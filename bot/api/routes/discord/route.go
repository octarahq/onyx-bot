package discord

import (
	"net/http"
	"onyx/bot/api"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/discord",
		Handler: handleDiscord,
	})
}

func handleDiscord(c *gin.Context) {
	_ = godotenv.Load()
	c.Redirect(http.StatusPermanentRedirect, os.Getenv("SUPPORT_INVITE"))
}
