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

type EpicProvider struct {
	lastSeen map[string]int
}

func init() {
	RegisterProvider(&EpicProvider{
		lastSeen: make(map[string]int),
	})
}

func (e *EpicProvider) Name() string {
	return "epic"
}

func (e *EpicProvider) Init(bot *core.Bot, db *gorm.DB) error {
	go e.pollFeeds(bot, db)
	return nil
}

func (e *EpicProvider) UISchema(locale discord.Locale) []core.UIComponent {
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

func (e *EpicProvider) pollFeeds(bot *core.Bot, db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	e.checkFeeds(bot, db, true)

	for {
		<-ticker.C
		e.checkFeeds(bot, db, false)
	}
}

func (e *EpicProvider) checkFeeds(bot *core.Bot, db *gorm.DB, initial bool) {
	var activeSettings []FreeGamesSettings
	db.Where("epic_enabled = ?", true).Find(&activeSettings)

	if len(activeSettings) == 0 {
		return
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	url := "https://www.gamerpower.com/api/giveaways?platform=epic-games-store"
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
		if !settings.Epic.Enabled || settings.Epic.Channel == "" {
			continue
		}

		cacheKey := settings.GuildID + "_epic"

		if e.lastSeen[cacheKey] == latestItem.ID {
			continue
		}

		if !initial && e.lastSeen[cacheKey] != 0 {
			content := fmt.Sprintf("A new free game is available! **%s**", latestItem.Title)
			
			if settings.Epic.Role != "" {
				content = fmt.Sprintf("<@&%s> %s", settings.Epic.Role, content)
			}

			DispatchMessage(bot, FluxMessage{
				ChannelID: settings.Epic.Channel,
				Content:   content,
				GameInfo: FreeGameInfo{
					Title:       latestItem.Title,
					Description: latestItem.Description,
					Thumbnail:   latestItem.Thumbnail,
					URL:         latestItem.URL,
					Worth:       latestItem.Worth,
				},
				LinkLabel: "Get Game on Epic Games",
			})
		}

		e.lastSeen[cacheKey] = latestItem.ID
	}
}
