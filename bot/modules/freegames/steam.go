package freegames

import (
	"encoding/json"
	"fmt"
	"net/http"
	"onyx/bot/core"
	"time"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type SteamProvider struct {
	lastSeen map[string]int
}

func init() {
	RegisterProvider(&SteamProvider{
		lastSeen: make(map[string]int),
	})
}

func (s *SteamProvider) Name() string {
	return "steam"
}

func (s *SteamProvider) Init(bot *core.Bot, db *gorm.DB) error {
	go s.pollFeeds(bot, db)
	return nil
}

func (s *SteamProvider) UISchema(locale discord.Locale) []core.UIComponent {
	return []core.UIComponent{
		{
			Name:         "channel",
			Label:        "Channel",
			Type:         core.ComponentTypeChannel,
			Required:     true,
			ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText, discord.ChannelTypeGuildNews},
		},
		{
			Name:     "role",
			Label:    "Role to mention (optional)",
			Type:     core.ComponentTypeRole,
			Required: false,
		},
	}
}

type GamerPowerGiveaway struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"open_giveaway_url"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Worth       string `json:"worth"`
}

func (s *SteamProvider) pollFeeds(bot *core.Bot, db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	s.checkFeeds(bot, db, true)

	for {
		<-ticker.C
		s.checkFeeds(bot, db, false)
	}
}

func (s *SteamProvider) checkFeeds(bot *core.Bot, db *gorm.DB, initial bool) {
	var activeSettings []FreeGamesSettings
	db.Where("steam_enabled = ?", true).Find(&activeSettings)

	if len(activeSettings) == 0 {
		return
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	url := "https://www.gamerpower.com/api/giveaways?platform=steam"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", fmt.Sprintf("OnyxBot/%s", bot.Version))

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var giveaways []GamerPowerGiveaway
	if err := json.NewDecoder(resp.Body).Decode(&giveaways); err != nil {
		return
	}

	if len(giveaways) == 0 {
		return
	}

	latestItem := giveaways[0]

	for _, settings := range activeSettings {
		if !settings.Steam.Enabled || settings.Steam.Channel == "" {
			continue
		}

		cacheKey := settings.GuildID + "_steam"

		if s.lastSeen[cacheKey] == latestItem.ID {
			continue
		}

		if !initial && s.lastSeen[cacheKey] != 0 {
			content := fmt.Sprintf("A new free game is available! **%s**", latestItem.Title)
			
			if settings.Steam.Role != "" {
				content = fmt.Sprintf("<@&%s> %s", settings.Steam.Role, content)
			}

			DispatchMessage(bot, FluxMessage{
				ChannelID: settings.Steam.Channel,
				Content:   content,
				GameInfo: FreeGameInfo{
					Title:       latestItem.Title,
					Description: latestItem.Description,
					Thumbnail:   latestItem.Thumbnail,
					URL:         latestItem.URL,
					Worth:       latestItem.Worth,
				},
				LinkLabel: "Get Game on Steam",
			})
		}

		s.lastSeen[cacheKey] = latestItem.ID
	}
}