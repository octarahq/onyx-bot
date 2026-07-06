package modules

import (
	"encoding/json"
	"net/http"
	"onyx/bot/api"
	"onyx/bot/core"
	"onyx/bot/db"
	"reflect"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId/modules/:module",
		Handler: handleGetModuleData,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodPatch,
		Path:    "/dash/guilds/:guildId/modules/:module",
		Handler: handlePatchModuleData,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodPost,
		Path:    "/dash/guilds/:guildId/modules/:module",
		Handler: handlePostModuleData,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodDelete,
		Path:    "/dash/guilds/:guildId/modules/:module",
		Handler: handleDeleteModuleData,
	})

	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId/modules",
		Handler: handleGetModuleList,
	})
}

func getModuleSettingsContext(c *gin.Context) (*core.Bot, string, interface{}, bool) {
	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found"})
		return nil, "", nil, false
	}

	guildId := c.Param("guildId")
	guildIdSnowflake, err := snowflake.Parse(guildId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id"})
		return nil, "", nil, false
	}

	if _, ok := bot.Client.Caches.GuildCache().Get(guildIdSnowflake); !ok {
		if _, err := bot.Client.Rest.GetGuild(guildIdSnowflake, false); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "guild not found (bot is not in this server)"})
			return nil, "", nil, false
		}
	}

	module := c.Param("module")
	var mod core.Module
	for _, m := range bot.Modules {
		if strings.EqualFold(m.Name(), module) || strings.EqualFold(strings.TrimSuffix(m.Name(), "Module"), module) {
			mod = m
			break
		}
	}

	if mod == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return nil, "", nil, false
	}

	dbAware, ok := mod.(core.DatabaseAware)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module does not support settings"})
		return nil, "", nil, false
	}

	if err := dbAware.LoadData(bot.DB.GormDB, guildId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return nil, "", nil, false
	}

	return bot, guildId, dbAware.DataPtr(), true
}

func handleGetModuleData(c *gin.Context) {
	api.GuildAuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	api.PermissionMiddleware(discord.PermissionManageGuild)(c)
	if c.IsAborted() {
		return
	}

	_, _, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if dataPtr != nil {
		c.JSON(http.StatusOK, dataPtr)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found"})
	}
}

func handleGetModuleList(c *gin.Context) {
	api.GuildAuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	api.PermissionMiddleware(discord.PermissionManageGuild)(c)
	if c.IsAborted() {
		return
	}

	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found"})
		return
	}

	guildId := c.Param("guildId")
	guildIdSnowflake, err := snowflake.Parse(guildId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id"})
		return
	}

	if _, ok := bot.Client.Caches.GuildCache().Get(guildIdSnowflake); !ok {
		if _, err := bot.Client.Rest.GetGuild(guildIdSnowflake, false); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "guild not found (bot is not in this server)"})
			return
		}
	}

	modules := make([]gin.H, 0, len(bot.Modules))

	for _, m := range bot.Modules {
		moduleName := m.Name()
		active := false

		if dbAware, ok := m.(core.DatabaseAware); ok {
			if err := dbAware.LoadData(bot.DB.GormDB, guildId); err == nil {
				ptr := dbAware.DataPtr()
				val := reflect.ValueOf(ptr).Elem()

				if val.Kind() == reflect.Struct {
					f := val.FieldByName("Enabled")
					if f.IsValid() && f.Kind() == reflect.Bool {
						active = f.Bool()
					}
				} else if val.Kind() == reflect.Bool {
					active = val.Bool()
				}
			}
		}

		modules = append(modules, gin.H{
			"name":    moduleName,
			"enabled": active,
		})
	}

	c.JSON(http.StatusOK, gin.H{"modules": modules})
}

func handlePatchModuleData(c *gin.Context) {
	api.GuildAuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	api.PermissionMiddleware(discord.PermissionManageGuild)(c)
	if c.IsAborted() {
		return
	}

	bot, _, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if dataPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found"})
		return
	}

	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payload"})
		return
	}

	if err := json.Unmarshal(jsonBytes, dataPtr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to apply updates"})
		return
	}

	if v, ok := dataPtr.(db.Validatable); ok {
		if err := v.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := bot.DB.GormDB.Save(dataPtr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func setModuleStatus(c *gin.Context, status bool) {
	api.GuildAuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	api.PermissionMiddleware(discord.PermissionManageGuild)(c)
	if c.IsAborted() {
		return
	}

	bot, _, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if dataPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found"})
		return
	}

	updated := false
	val := reflect.ValueOf(dataPtr).Elem()
	if val.Kind() == reflect.Struct {
		f := val.FieldByName("Enabled")
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.Bool {
			f.SetBool(status)
			updated = true
		}
	} else if val.Kind() == reflect.Bool {
		val.SetBool(status)
		updated = true
	}

	if !updated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module does not support enabling/disabling"})
		return
	}

	if err := bot.DB.GormDB.Save(dataPtr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func handlePostModuleData(c *gin.Context) {
	setModuleStatus(c, true)
}

func handleDeleteModuleData(c *gin.Context) {
	setModuleStatus(c, false)
}
