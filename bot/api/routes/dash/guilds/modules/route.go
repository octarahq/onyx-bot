package modules

import (
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

func getModuleSettingsContext(c *gin.Context) (*core.Bot, string, *db.Guild, reflect.Value, bool) {
	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found"})
		return nil, "", nil, reflect.Value{}, false
	}

	guildId := c.Param("guildId")
	guildIdSnowflake, err := snowflake.Parse(guildId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id"})
		return nil, "", nil, reflect.Value{}, false
	}

	if _, ok := bot.Client.Caches.GuildCache().Get(guildIdSnowflake); !ok {
		if _, err := bot.Client.Rest.GetGuild(guildIdSnowflake, false); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "guild not found (bot is not in this server)"})
			return nil, "", nil, reflect.Value{}, false
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
		return nil, "", nil, reflect.Value{}, false
	}

	guildData, err := db.LoadSettings(bot.DB.GormDB, guildId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return nil, "", nil, reflect.Value{}, false
	}

	settingsValue := reflect.ValueOf(guildData).Elem()
	fieldName := "Settings" + strings.TrimSuffix(mod.Name(), "Module")
	field := settingsValue.FieldByName(fieldName)

	return bot, guildId, guildData, field, true
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

	_, _, _, field, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if field.IsValid() {
		c.JSON(http.StatusOK, field.Interface())
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

	guildData, err := db.LoadSettings(bot.DB.GormDB, guildId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	settingsValue := reflect.ValueOf(guildData).Elem()
	modules := make([]gin.H, 0, len(bot.Modules))

	for _, m := range bot.Modules {
		moduleName := m.Name()
		fieldName := "Settings" + strings.TrimSuffix(moduleName, "Module")
		field := settingsValue.FieldByName(fieldName)
		active := false

		if field.IsValid() {
			if field.Kind() == reflect.Struct {
				f := field.FieldByName("Enabled")
				if f.IsValid() && f.Kind() == reflect.Bool {
					active = f.Bool()
				}
			} else if field.Kind() == reflect.Bool {
				active = field.Bool()
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

	bot, _, guildData, field, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if !field.IsValid() || !field.CanSet() {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found"})
		return
	}

	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}

	updated := false
	if enabled, ok := payload["enabled"]; ok {
		if v, ok := enabled.(bool); ok {
			if field.Kind() == reflect.Struct {
				f := field.FieldByName("Enabled")
				if f.IsValid() && f.CanSet() && f.Kind() == reflect.Bool {
					f.SetBool(v)
					updated = true
				}
			} else if field.Kind() == reflect.Bool {
				field.SetBool(v)
				updated = true
			}
		}
	}

	if !updated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no supported fields to update"})
		return
	}

	if err := db.UpdateSettings(bot.DB.GormDB, guildData); err != nil {
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

	bot, _, guildData, field, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if !field.IsValid() || !field.CanSet() {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found"})
		return
	}

	updated := false
	if field.Kind() == reflect.Struct {
		f := field.FieldByName("Enabled")
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.Bool {
			f.SetBool(status)
			updated = true
		}
	} else if field.Kind() == reflect.Bool {
		field.SetBool(status)
		updated = true
	}

	if !updated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module does not support enabling/disabling"})
		return
	}

	if err := db.UpdateSettings(bot.DB.GormDB, guildData); err != nil {
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
