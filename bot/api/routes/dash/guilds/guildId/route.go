package guildid

import (
	"net/http"
	"onyx/bot/api"
	"onyx/bot/core"
	"reflect"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId",
		Handler: handleGetGuildInfo,
	})
}

func handleGetGuildInfo(c *gin.Context) {
	api.GuildAuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	api.PermissionMiddleware(discord.PermissionManageGuild)(c)
	if c.IsAborted() {
		return
	}

	gid := c.Param("guildId")
	sgid, err := snowflake.Parse(gid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id", "error_code": "INVALID_GUILD_ID"})
		return
	}

	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	var guild discord.Guild
	cguild, ok := bot.Client.Caches.GuildCache().Get(sgid)
	if ok {
		guild = cguild
	} else {
		rguid, err := bot.Client.Rest.GetGuild(sgid, false)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "guild not found (bot is not in this server)", "error_code": "GUILD_NOT_FOUND"})
			return
		}
		guild = rguid.Guild
	}

	modules := make([]gin.H, 0, len(bot.Modules))

	for _, m := range bot.Modules {
		moduledata := m.Metadata()
		active := false

		if dbAware, ok := m.(core.DatabaseAware); ok {
			if err := dbAware.LoadData(bot.DB.GormDB, gid); err == nil {
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

		var label, description string
		if moduledata.Label != nil {
			label = moduledata.Label(locale)
		}
		if moduledata.Description != nil {
			description = moduledata.Description(locale)
		}

		var submodules map[string]core.SubmoduleMeta
		if moduledata.Submodules != nil {
			submodules = moduledata.Submodules(locale)
		}

		modules = append(modules, gin.H{
			"name":        moduledata.Name,
			"enabled":     active,
			"icon":        moduledata.Icon,
			"label":       label,
			"description": description,
			"submodules":  submodules,
		})
	}

	iconURL := "https://cdn.discordapp.com/embed/avatars/1.png"
	if url := guild.IconURL(); url != nil {
		iconURL = *url
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          guild.ID.String(),
		"name":        guild.Name,
		"iconURL":     iconURL,
		"memberCount": guild.MemberCount,
		"ownerId":     guild.OwnerID.String(),
		"modules":     modules,
	})
}
