package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"onyx/bot/api"
	"onyx/bot/core"
	"onyx/bot/db"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	oauthConfig     *oauth2.Config
	userGuildsCache = api.NewExpiringCache[string, []map[string]interface{}](2 * time.Minute)
)

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/auth/discord/login",
		Handler: handleLogin,
	})
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/auth/discord/callback",
		Handler: handleCallback,
	})

	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/auth/discord/@me",
		Handler: handleMe,
	})

	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/auth/discord/@me/guilds",
		Handler: handleMeGuilds,
	})

	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/auth/discord/logout",
		Handler: handleLogout,
	})
}

func getOAuthConfig() *oauth2.Config {
	if oauthConfig == nil {
		oauthConfig = &oauth2.Config{
			ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("DISCORD_REDIRECT_URI"),
			Scopes:       []string{"identify", "guilds"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		}
	}
	return oauthConfig
}

func handleLogin(c *gin.Context) {
	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	url := getOAuthConfig().AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func handleCallback(c *gin.Context) {
	state := c.Query("state")
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state token"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is missing"})
		return
	}

	token, err := getOAuthConfig().Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	defer resp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	sessionBytes := make([]byte, 32)
	rand.Read(sessionBytes)
	sessionID := hex.EncodeToString(sessionBytes)

	dbInstance := c.MustGet("db").(*db.DB)
	newSession := db.Session{
		SessionID:   sessionID,
		UserID:      user.ID,
		AccessToken: token.AccessToken,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	if err := dbInstance.GormDB.Create(&newSession).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database session"})
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("session_id", sessionID, 7*24*3600, "/", "", false, true)

	_ = godotenv.Load()
	c.Redirect(http.StatusPermanentRedirect, os.Getenv("SITE_DASHBOARD_URL"))
}

func handleMe(c *gin.Context) {
	api.AuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	user, exists := c.MustGet("user").(*discord.User)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user context not found", "error_code": "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"discriminator": user.Discriminator,
		"avatar":        user.Avatar,
		"avatarURL":     user.AvatarURL(),
	})
}

func handleMeGuilds(c *gin.Context) {
	api.AuthMiddleware()(c)
	if c.IsAborted() {
		return
	}

	session, exists := c.MustGet("session").(*db.Session)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	if session.AccessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Discord access token. Please log out and log back in.", "error_code": "AUTH_MISSING_TOKEN"})
		return
	}

	if cachedGuilds, ok := userGuildsCache.Get(session.UserID); ok {
		c.JSON(http.StatusOK, cachedGuilds)
		return
	}

	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/users/@me/guilds", nil)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println(resp)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch guilds"})
		return
	}
	defer resp.Body.Close()

	var discordGuilds []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		Owner       bool   `json:"owner"`
		Permissions string `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&discordGuilds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse guilds data"})
		return
	}

	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	var guilds []map[string]interface{}
	for _, g := range discordGuilds {
		guildID, err := snowflake.Parse(g.ID)
		if err != nil {
			continue
		}

		_, botInGuild := bot.Client.Caches.Guild(guildID)

		permsUint, _ := strconv.ParseUint(g.Permissions, 10, 64)
		hasManageServer := discord.Permissions(permsUint).Has(discord.PermissionManageGuild)

		if botInGuild || hasManageServer {
			avatarURL := ""
			if g.Icon != "" {
				avatarURL = "https://cdn.discordapp.com/icons/" + g.ID + "/" + g.Icon + ".png"
			} else {
				avatarURL = "https://cdn.discordapp.com/embed/avatars/1.png"
			}

			guilds = append(guilds, map[string]interface{}{
				"id":           g.ID,
				"name":         g.Name,
				"avatarURL":    avatarURL,
				"owner":        g.Owner,
				"permissions":  g.Permissions,
				"bot_in_guild": botInGuild,
			})
		}
	}

	userGuildsCache.Set(session.UserID, guilds)

	c.JSON(http.StatusOK, guilds)
}

func handleLogout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		database, exists := c.Get("db")
		if exists {
			gormDB := database.(*db.DB).GormDB
			gormDB.Where("session_id = ?", sessionID).Delete(&db.Session{})
		}
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}
