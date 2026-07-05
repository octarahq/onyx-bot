package ping

import (
	"net/http"
	"onyx/bot/api"

	"github.com/gin-gonic/gin"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/ping",
		Handler: handlePing,
	})
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
