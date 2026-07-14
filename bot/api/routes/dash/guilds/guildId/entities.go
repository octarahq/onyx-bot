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
		Path:    "/dash/guilds/:guildId/roles",
		Handler: handleGetRoles,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId/channels",
		Handler: handleGetChannels,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId/emojis",
		Handler: handleGetEmojis,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/dash/guilds/:guildId/members",
		Handler: handleGetMembers,
	})
}

func handleGetRoles(c *gin.Context) {
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

	roles, err := bot.Client.Rest.GetRoles(sgid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get roles", "error_code": "DISCORD_API_ERROR"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func handleGetChannels(c *gin.Context) {
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

	channels, err := bot.Client.Rest.GetGuildChannels(sgid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get channels", "error_code": "DISCORD_API_ERROR"})
		return
	}

	c.JSON(http.StatusOK, channels)
}

func handleGetEmojis(c *gin.Context) {
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

	emojis, err := bot.Client.Rest.GetEmojis(sgid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get emojis", "error_code": "DISCORD_API_ERROR"})
		return
	}

	c.JSON(http.StatusOK, emojis)
}

func handleGetMembers(c *gin.Context) {
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

	members, err := bot.Client.Rest.GetMembers(sgid, 1000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get members", "error_code": "DISCORD_API_ERROR"})
		return
	}

	c.JSON(http.StatusOK, members)
}
