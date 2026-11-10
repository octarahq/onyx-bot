package socialnotifs

import (
	"onyx/bot/core"
	"onyx/bot/utils"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

type RSSProvider struct {
	lastSeen map[string]string
}

func init() {
	RegisterProvider(&RSSProvider{
		lastSeen: make(map[string]string),
	})
}

func (r *RSSProvider) Name() string {
	return "rss"
}

func (r *RSSProvider) Init(bot *core.Bot, db *gorm.DB) error {
	go r.pollFeeds(bot, db)
	return nil
}

func (r *RSSProvider) UISchema(locale discord.Locale) []core.UIComponent {

	return []core.UIComponent{
		{
			Name:     "feed_url",
			Label:    "RSS Feed URL",
			Type:     core.ComponentTypeString,
			Required: true,
			Max:      100,
		},
		{
			Name:         "channel",
			Label:        "Channel",
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
				{Key: "feed.title", Label: "Feed Title", Description: "The name of the RSS feed", MaxLength: 40},
				{Key: "title", Label: "Item Title", Description: "The title of the new item/post", MaxLength: 40},
				{Key: "url", Label: "Item URL", Description: "The direct link to the new item/post"},
				{Key: "description", Label: "Description", Description: "The snippet or description of the item", MaxLength: 1000},
			},
			Max: 100,
		},
		{
			Name:        "button_text",
			Label:       "Button Text",
			Description: "Leave empty for default (e.g., 'Lire l'article')",
			Type:        core.ComponentTypeString,
			Max:         20,
			Required:    false,
		},
	}
}

func (r *RSSProvider) pollFeeds(bot *core.Bot, db *gorm.DB) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		r.checkFeeds(bot, db)
		<-ticker.C
	}
}

func (r *RSSProvider) checkFeeds(bot *core.Bot, db *gorm.DB) {
	var activeSettings []SocialNotifsSettings
	db.Where("rss_enabled = ?", true).Find(&activeSettings)

	fp := gofeed.NewParser()

	for _, settings := range activeSettings {
		for _, flux := range settings.RSS.Fluxs {

			feed, err := fp.ParseURL(flux.FeedURL)
			if err != nil {
				continue
			}
			if feed == nil || len(feed.Items) == 0 {
				continue
			}

			latestItem := feed.Items[0]
			cacheKey := settings.GuildID + "_" + flux.FeedURL

			if r.lastSeen[cacheKey] == latestItem.GUID || r.lastSeen[cacheKey] == latestItem.Link {
				continue
			}

			if r.lastSeen[cacheKey] != "" || true {
				vars := map[string]string{
					"feed.title":  feed.Title,
					"title":       latestItem.Title,
					"url":         latestItem.Link,
					"description": latestItem.Description,
				}

				content := utils.ParseVariables(flux.Message, vars)

				DispatchMessageV2(bot, FluxMessage{
					ChannelID: flux.Channel,
					Content:   content,
					Link:      latestItem.Link,
					LinkLabel: flux.ButtonText,
				})

			}

			if latestItem.GUID != "" {
				r.lastSeen[cacheKey] = latestItem.GUID
			} else {
				r.lastSeen[cacheKey] = latestItem.Link
			}
		}
	}
}
