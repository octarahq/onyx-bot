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
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId/modules/:module/schema",
		Handler: handleGetModuleSchema,
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

func checkBotPermission(b *core.Bot, module core.Module, me discord.Member) []string {
	var missing []string
	botPerms := b.Client.Caches.MemberPermissions(me)

	if botPerms.Has(discord.PermissionAdministrator) {
		return missing
	}

	permissions := module.Permissions()
	for _, p := range permissions {
		if !botPerms.Has(p) {
			missing = append(missing, p.String())
		}
	}

	return missing
}

func getModuleSettingsContext(c *gin.Context) (*core.Bot, string, core.Module, interface{}, bool) {
	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return nil, "", nil, nil, false
	}

	guildId := c.Param("guildId")
	guildIdSnowflake, err := snowflake.Parse(guildId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id", "error_code": "INVALID_GUILD_ID"})
		return nil, "", nil, nil, false
	}

	if _, ok := bot.Client.Caches.GuildCache().Get(guildIdSnowflake); !ok {
		if _, err := bot.Client.Rest.GetGuild(guildIdSnowflake, false); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "guild not found (bot is not in this server)", "error_code": "GUILD_NOT_FOUND"})
			return nil, "", nil, nil, false
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
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found", "error_code": "MODULE_NOT_FOUND"})
		return nil, "", nil, nil, false
	}

	dbAware, ok := mod.(core.DatabaseAware)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module does not support settings", "error_code": "NOT_TOGGLABLE"})
		return nil, "", nil, nil, false
	}

	if err := dbAware.LoadData(bot.DB.GormDB, guildId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings", "error_code": "INTERNAL_ERROR"})
		return nil, "", nil, nil, false
	}

	return bot, guildId, mod, dbAware.DataPtr(), true
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

	_, _, _, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if dataPtr != nil {
		c.JSON(http.StatusOK, dataPtr)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found", "error_code": "MODULE_DATA_NOT_FOUND"})
	}
}

func handleGetModuleSchema(c *gin.Context) {
	api.GuildAuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	api.PermissionMiddleware(discord.PermissionManageGuild)(c)
	if c.IsAborted() {
		return
	}

	_, _, mod, _, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	provider, ok := mod.(core.UIProvider)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "module does not provide a UI schema", "error_code": "NO_SCHEMA"})
		return
	}

	c.JSON(http.StatusOK, provider.UISchema())
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	guildId := c.Param("guildId")
	guildIdSnowflake, err := snowflake.Parse(guildId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id", "error_code": "INVALID_GUILD_ID"})
		return
	}

	if _, ok := bot.Client.Caches.GuildCache().Get(guildIdSnowflake); !ok {
		if _, err := bot.Client.Rest.GetGuild(guildIdSnowflake, false); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "guild not found (bot is not in this server)", "error_code": "GUILD_NOT_FOUND"})
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

	bot, _, _, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if dataPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found", "error_code": "MODULE_DATA_NOT_FOUND"})
		return
	}

	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body", "error_code": "INVALID_PAYLOAD"})
		return
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payload", "error_code": "INTERNAL_ERROR"})
		return
	}

	if err := json.Unmarshal(jsonBytes, dataPtr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to apply updates", "error_code": "VALIDATION_FAILED"})
		return
	}

	if v, ok := dataPtr.(db.Validatable); ok {
		if err := v.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "error_code": "VALIDATION_FAILED"})
			return
		}
	}

	if err := bot.DB.GormDB.Save(dataPtr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings", "error_code": "INTERNAL_ERROR"})
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

	bot, guildId, mod, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	if dataPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module data not found", "error_code": "MODULE_DATA_NOT_FOUND"})
		return
	}

	if status {
		guildIdSnowflake, err := snowflake.Parse(guildId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id", "error_code": "INVALID_GUILD_ID"})
			return
		}

		me, ok := bot.Client.Caches.Member(guildIdSnowflake, bot.Client.ID())
		if !ok {
			m, err := bot.Client.Rest.GetMember(guildIdSnowflake, bot.Client.ID())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bot member", "error_code": "INTERNAL_ERROR"})
				return
			}
			me = *m
		}

		missingPerms := checkBotPermission(bot, mod, me)
		if len(missingPerms) > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "The bot lacks the required permissions to enable this module.",
				"error_code": "BOT_MISSING_PERMISSIONS",
				"missing": missingPerms,
			})
			return
		}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "module does not support enabling/disabling", "error_code": "NOT_TOGGLABLE"})
		return
	}

	if err := bot.DB.GormDB.Save(dataPtr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings", "error_code": "INTERNAL_ERROR"})
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
