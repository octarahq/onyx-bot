package modules

import (
	"encoding/json"
	"fmt"
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
		if strings.EqualFold(m.Metadata().Name, module) || strings.EqualFold(strings.TrimSuffix(m.Metadata().Name, "Module"), module) {
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

	_, _, mod, dataPtr, ok := getModuleSettingsContext(c)
	if !ok {
		return
	}

	lang := c.Query("lang")
	headerLang := c.GetHeader("Accept-Language")

	if lang == "" {
		lang = headerLang
		if len(lang) > 2 {
			lang = lang[:2]
		}
	}
	if lang == "" {
		lang = string(discord.LocaleEnglishUS)
	}
	locale := discord.Locale(lang)

	moduledata := mod.Metadata()
	var label, description string
	if moduledata.Label != nil {
		label = moduledata.Label(locale)
	}
	if moduledata.Description != nil {
		description = moduledata.Description(locale)
	}

	var submodules map[string]core.SubmoduleMeta
	if moduledata.Submodules != nil && c.Query("submodules") == "true" {
		submodules = moduledata.Submodules(locale)
	}

	if dataPtr != nil {
		c.JSON(http.StatusOK, gin.H{
			"metadata": gin.H{
				"name":        moduledata.Name,
				"icon":        moduledata.Icon,
				"label":       label,
				"description": description,
				"submodules":  submodules,
			},
			"data": dataPtr,
		})
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

	lang := c.Query("lang")
	headerLang := c.GetHeader("Accept-Language")

	if lang == "" {
		lang = headerLang
		if len(lang) > 2 {
			lang = lang[:2]
		}
	}
	if lang == "" {
		lang = string(discord.LocaleEnglishUS)
	}
	locale := discord.Locale(lang)

	c.JSON(http.StatusOK, provider.UISchema(locale))
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

	bot, guildId, mod, dataPtr, ok := getModuleSettingsContext(c)
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

	if provider, ok := mod.(core.UIProvider); ok {
		schema := provider.UISchema(discord.LocaleEnglishUS)
		for _, sub := range schema.SubModules {
			subData, ok := payload[sub.Name].(map[string]any)
			if !ok {
				continue
			}

			isEnabled := true
			if enabledVal, hasEnabled := subData["enabled"]; hasEnabled {
				if bVal, ok := enabledVal.(bool); ok {
					isEnabled = bVal
				}
			}

			for _, comp := range sub.Components {
				val, exists := subData[comp.Name]
				if !exists {
					if comp.Required && isEnabled {
						c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s is required", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
						return
					}
					continue
				}

				var strVal string
				var isString bool

				if s, ok := val.(string); ok {
					strVal = s
					isString = true
				} else if arr, ok := val.([]interface{}); ok {
					if comp.Multiple {
						var strArr []string
						for _, item := range arr {
							if s, ok := item.(string); ok {
								strArr = append(strArr, s)
							}
						}
						strVal = strings.Join(strArr, ",")
						isString = true
						subData[comp.Name] = strVal
					}
				}

				if isString {
					if comp.Required && strVal == "" && isEnabled {
						c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s is required", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
						return
					}

					if comp.Type == core.ComponentTypeString || comp.Type == core.ComponentTypeTextarea {
						if comp.Min != nil && len(strVal) < *comp.Min {
							c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s is too short", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
							return
						}
						if comp.Max != nil && len(strVal) > *comp.Max {
							c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s is too long", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
							return
						}
					}

					if comp.Type == core.ComponentTypeChannel {
						channels := []string{strVal}
						if comp.Multiple {
							channels = strings.Split(strVal, ",")
						}

						if comp.Max != nil && len(channels) > *comp.Max {
							c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s has too many channels", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
							return
						}

						for _, ch := range channels {
							ch = strings.TrimSpace(ch)
							if ch == "" {
								continue
							}

							chSF, err := snowflake.Parse(ch)
							if err != nil {
								c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid channel ID: %s", ch), "error_code": "VALIDATION_FAILED"})
								return
							}

							var channelGuild discord.GuildChannel
							var cOk bool
							channelGuild, cOk = bot.Client.Caches.Channel(chSF)
							if !cOk {
								chInt, err := bot.Client.Rest.GetChannel(chSF)
								if err != nil {
									c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("channel %s does not exist", ch), "error_code": "VALIDATION_FAILED"})
									return
								}
								var isGuild bool
								channelGuild, isGuild = chInt.(discord.GuildChannel)
								if !isGuild {
									c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("channel %s is not a guild channel", ch), "error_code": "VALIDATION_FAILED"})
									return
								}
							}

							if channelGuild.GuildID().String() != guildId {
								c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("channel %s is not in this server", ch), "error_code": "VALIDATION_FAILED"})
								return
							}
						}
					}

					if comp.Type == core.ComponentTypeRole {
						roles := []string{strVal}
						if comp.Multiple {
							roles = strings.Split(strVal, ",")
						}

						if comp.Max != nil && len(roles) > *comp.Max {
							c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s has too many roles", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
							return
						}

						guildSF, _ := snowflake.Parse(guildId)
						for _, r := range roles {
							r = strings.TrimSpace(r)
							if r == "" {
								continue
							}

							rSF, err := snowflake.Parse(r)
							if err != nil {
								c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid role ID: %s", r), "error_code": "VALIDATION_FAILED"})
								return
							}

							_, ok := bot.Client.Caches.Role(guildSF, rSF)
							if !ok {
								roleRoles, err := bot.Client.Rest.GetRoles(guildSF)
								if err != nil {
									c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("role %s does not exist", r), "error_code": "VALIDATION_FAILED"})
									return
								}
								found := false
								for _, gr := range roleRoles {
									if gr.ID == rSF {
										found = true
										break
									}
								}
								if !found {
									c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("role %s does not exist", r), "error_code": "VALIDATION_FAILED"})
									return
								}
							}
						}
					}
				}

				if floatVal, ok := val.(float64); ok {
					if comp.Type == core.ComponentTypeNumber {
						if comp.Min != nil && int(floatVal) < *comp.Min {
							c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s is too small", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
							return
						}
						if comp.Max != nil && int(floatVal) > *comp.Max {
							c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("field %s.%s is too large", sub.Name, comp.Name), "error_code": "VALIDATION_FAILED"})
							return
						}
					}
				}
			}
		}
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
				"error":      "The bot lacks the required permissions to enable this module.",
				"error_code": "BOT_MISSING_PERMISSIONS",
				"missing":    missingPerms,
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
