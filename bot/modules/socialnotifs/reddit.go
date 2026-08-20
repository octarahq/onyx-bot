package socialnotifs

import (
	"fmt"
	"net/http"
	"onyx/bot/core"
	"onyx/bot/utils"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

type RedditProvider struct {
	lastSeen map[string]string
}

func init() {
	RegisterProvider(&RedditProvider{
		lastSeen: make(map[string]string),
	})
}

func (r *RedditProvider) Name() string {
	return "reddit"
}

func (r *RedditProvider) Init(bot *core.Bot, db *gorm.DB) error {
	go r.pollFeeds(bot, db)
	return nil
}

func (r *RedditProvider) UISchema(locale discord.Locale) []core.UIComponent {
	return []core.UIComponent{
		{
			Name:     "subreddit",
			Label:    "Subreddit (e.g. r/gaming or gaming)",
			Type:     core.ComponentTypeString,
			Required: true,
			Max:      50,
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
				{Key: "feed.title", Label: "Feed Title", Description: "The name of the subreddit", MaxLength: 40},
				{Key: "title", Label: "Post Title", Description: "The title of the new post", MaxLength: 40},
				{Key: "url", Label: "Post URL", Description: "The direct link to the new post"},
			},
			Max: 1000,
		},
		{
			Name:        "button_text",
			Label:       "Button Text",
			Description: "Leave empty for default (e.g., 'Read Post')",
			Type:        core.ComponentTypeString,
			Max:         20,
			Required:    false,
		},
	}
}

func (r *RedditProvider) pollFeeds(bot *core.Bot, db *gorm.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		r.checkFeeds(bot, db)
		<-ticker.C
	}
}

func (r *RedditProvider) checkFeeds(bot *core.Bot, db *gorm.DB) {
	var activeSettings []SocialNotifsSettings
	db.Where("reddit_enabled = ?", true).Find(&activeSettings)

	fp := gofeed.NewParser()
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	for _, settings := range activeSettings {
		for _, flux := range settings.Reddit.Fluxs {
			subName := strings.TrimSpace(flux.Subreddit)
			if strings.HasPrefix(subName, "r/") {
				subName = subName[2:]
			}

			if subName == "" {
				continue
			}

			url := "https://www.reddit.com/r/" + subName + "/new.rss"

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				continue
			}
			req.Header.Set("User-Agent", fmt.Sprintf("OnyxBot/%s", bot.Version))

			resp, err := client.Do(req)
			if err != nil {
				continue
			}

			if resp.StatusCode != 200 {
				resp.Body.Close()
				continue
			}

			feed, err := fp.Parse(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			if feed == nil || len(feed.Items) == 0 {
				continue
			}

			latestItem := feed.Items[0]
			cacheKey := settings.GuildID + "_" + url

			if r.lastSeen[cacheKey] == latestItem.GUID || r.lastSeen[cacheKey] == latestItem.Link {
				continue
			}

			if r.lastSeen[cacheKey] != "" {
				vars := map[string]string{
					"feed.title": feed.Title,
					"title":      latestItem.Title,
					"url":        latestItem.Link,
				}

				content := utils.ParseVariables(flux.Message, vars)

				buttonText := flux.ButtonText
				if buttonText == "" {
					buttonText = "Read Post"
				}

				DispatchMessage(bot, FluxMessage{
					ChannelID: flux.Channel,
					Content:   content,
					Link:      latestItem.Link,
					LinkLabel: buttonText,
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
