package guildid

import (
	"net/http"
	"onyx/bot/api"
	"onyx/bot/core"

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

	c.JSON(http.StatusOK, gin.H{
		"id":          guild.ID.String(),
		"name":        guild.Name,
		"iconURL":     *guild.IconURL(),
		"memberCount": guild.MemberCount,
		"ownerId":     guild.OwnerID.String(),
	})
}
