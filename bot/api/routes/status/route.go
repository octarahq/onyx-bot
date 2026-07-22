package ping

import (
	"net/http"
	"onyx/bot/api"
	"onyx/bot/core"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/status",
		Handler: handleStatus,
	})

	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/status/ping",
		Handler: handlePing,
	})

	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/status/bot",
		Handler: handleStatusBot,
	})
}

func handleStatus(c *gin.Context) {
	_ = godotenv.Load()
	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version":  bot.Version,
		"servers":  bot.Client.Caches.GuildsLen(),
		"users":    bot.Client.Caches.MembersAllLen(),
		"commands": len(bot.Commands),
	})
}

func handleStatusBot(c *gin.Context) {
	_, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	if exists {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
