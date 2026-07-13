package invite

import (
	"fmt"
	"net/http"
	"onyx/bot/api"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/invite",
		Handler: handleInvite,
	})
}

func handleInvite(c *gin.Context) {
	_ = godotenv.Load()
	gid := c.Query("guild_id")
	url := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s", os.Getenv("DISCORD_CLIENT_ID"))
	if gid != "" {
		url = fmt.Sprintf("%s&guild_id=%s", url, gid)
	}

	c.Redirect(http.StatusPermanentRedirect, url)
}
