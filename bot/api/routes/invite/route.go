package invite

import (
	"fmt"
	"net/http"
	"net/url"
	"onyx/bot/api"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/invite",
		Handler: handleInvite,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/invite/callback",
		Handler: handleInviteCallback,
	})
}

func handleInvite(c *gin.Context) {
	_ = godotenv.Load()
	gid := c.Query("guild_id")
	authorizeURL := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&scope=bot+applications.commands&permissions=1099511627767", os.Getenv("DISCORD_CLIENT_ID"))
	if gid != "" {
		authorizeURL = fmt.Sprintf("%s&guild_id=%s&disable_guild_select=true", authorizeURL, gid)
	}

	rurl := c.Query("redirect_url")
	if rurl != "" {
		if strings.HasPrefix(rurl, "/") {
			redirectURI := fmt.Sprintf("%s/api/invite/callback", os.Getenv("SITE_URL"))

			authorizeURL := fmt.Sprintf("%s&redirect_uri=%s&state=%s&response_type=code", authorizeURL, url.QueryEscape(redirectURI), url.QueryEscape(rurl))
			c.Redirect(http.StatusTemporaryRedirect, authorizeURL)
			return
		}
	}

	c.Redirect(http.StatusTemporaryRedirect, authorizeURL)
}

func handleInviteCallback(c *gin.Context) {
	_ = godotenv.Load()

	code := c.Query("code")
	if code != "" {
		data := url.Values{}
		data.Set("client_id", os.Getenv("DISCORD_CLIENT_ID"))
		data.Set("client_secret", os.Getenv("DISCORD_CLIENT_SECRET"))
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", fmt.Sprintf("%s/api/invite/callback", os.Getenv("SITE_URL")))

		req, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(data.Encode()))
		if err == nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		}
	}

	rurl := c.Query("state")
	if rurl == "" {
		rurl = c.Query("redirect_url")
	}

	if rurl != "" && strings.HasPrefix(rurl, "/") {
		c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("%s%s", os.Getenv("SITE_URL"), rurl))
		return
	}

	c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("%s/dashboard", os.Getenv("SITE_URL")))
}
