package socialnotifs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"onyx/bot/core"
	"onyx/bot/utils"
	"os"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type TwitchProvider struct {
	lastSeen map[string]string
}

func init() {
	RegisterProvider(&TwitchProvider{
		lastSeen: make(map[string]string),
	})
}

func (t *TwitchProvider) Name() string {
	return "twitch"
}

func (t *TwitchProvider) Init(bot *core.Bot, db *gorm.DB) error {
	go t.pollFeeds(bot, db)
	return nil
}

func (t *TwitchProvider) UISchema(locale discord.Locale) []core.UIComponent {
	return []core.UIComponent{
		{
			Name:     "channel_name",
			Label:    "Twitch Channel Name (or URL)",
			Type:     core.ComponentTypeString,
			Required: true,
			Max:      50,
		},
		{
			Name:         "channel",
			Label:        "Discord Channel",
			Type:         core.ComponentTypeChannel,
			Required:     true,
			ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText, discord.ChannelTypeGuildNews},
		},
		{
			Name:     "message",
			Label:    "Message",
			Type:     core.ComponentTypeTextarea,
			Required: true,
			Variables: []core.Variables{
				{Key: "channel.name", Label: "Channel Name", Description: "The name of the Twitch channel", MaxLength: 40},
				{Key: "title", Label: "Stream Title", Description: "The title of the stream", MaxLength: 40},
				{Key: "url", Label: "Stream URL", Description: "The direct link to the stream"},
				{Key: "game", Label: "Game", Description: "The game being played", MaxLength: 40},
			},
			Max: 1000,
		},
		{
			Name:        "button_text",
			Label:       "Button Text",
			Description: "Leave empty for default (e.g., 'Watch Live')",
			Type:        core.ComponentTypeString,
			Max:         20,
			Required:    false,
		},
	}
}

func (t *TwitchProvider) getAccessToken() (string, error) {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("missing TWITCH_CLIENT_ID or TWITCH_CLIENT_SECRET")
	}

	authURL := "https://id.twitch.tv/oauth2/token"
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get token, status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func extractTwitchLogin(input string) string {
	input = strings.TrimSpace(input)

	if strings.Contains(input, "twitch.tv/") {
		parts := strings.Split(input, "twitch.tv/")
		if len(parts) > 1 {
			loginPart := strings.Split(parts[1], "?")[0]
			loginPart = strings.Split(loginPart, "/")[0]
			return strings.ToLower(loginPart)
		}
	}
	return strings.ToLower(input)
}

func (t *TwitchProvider) pollFeeds(bot *core.Bot, db *gorm.DB) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		t.checkStreams(bot, db)
		<-ticker.C
	}
}

func (t *TwitchProvider) checkStreams(bot *core.Bot, db *gorm.DB) {
	var activeSettings []SocialNotifsSettings
	db.Where("twitch_enabled = ?", true).Find(&activeSettings)

	if len(activeSettings) == 0 {
		return
	}

	token, err := t.getAccessToken()
	if err != nil {
		fmt.Printf("[TWITCH-POLL] Error getting token: %v\n", err)
		return
	}

	clientID := os.Getenv("TWITCH_CLIENT_ID")
	client := &http.Client{Timeout: 10 * time.Second}

	for _, settings := range activeSettings {
		for _, flux := range settings.Twitch.Fluxs {
			login := extractTwitchLogin(flux.ChannelName)
			if login == "" {
				continue
			}

			cacheKey := settings.GuildID + "_" + login

			req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+login, nil)
			if err != nil {
				continue
			}
			req.Header.Set("Client-ID", clientID)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := client.Do(req)
			if err != nil {
				continue
			}

			if resp.StatusCode != 200 {
				resp.Body.Close()
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			var result struct {
				Data []struct {
					ID       string `json:"id"`
					UserName string `json:"user_name"`
					GameName string `json:"game_name"`
					Title    string `json:"title"`
					Type     string `json:"type"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &result); err != nil {
				continue
			}

			if len(result.Data) == 0 || result.Data[0].Type != "live" {
				delete(t.lastSeen, cacheKey)
				continue
			}

			stream := result.Data[0]

			if t.lastSeen[cacheKey] == stream.ID {
				continue
			}

			streamURL := "https://www.twitch.tv/" + login

			vars := map[string]string{
				"channel.name": stream.UserName,
				"title":        stream.Title,
				"url":          streamURL,
				"game":         stream.GameName,
			}

			content := utils.ParseVariables(flux.Message, vars)

			buttonText := flux.ButtonText
			if buttonText == "" {
				buttonText = "Watch Live"
			}

			DispatchMessage(bot, FluxMessage{
				ChannelID: flux.Channel,
				Content:   content,
				Link:      streamURL,
				LinkLabel: buttonText,
			})

			t.lastSeen[cacheKey] = stream.ID
		}
	}
}
