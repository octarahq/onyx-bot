package socialnotifs

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"onyx/bot/core"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type YouTubeProvider struct {
	db *gorm.DB
}

func init() {
	RegisterProvider(&YouTubeProvider{})
}

func (y *YouTubeProvider) Name() string {
	return "youtube"
}

func (y *YouTubeProvider) UISchema(locale discord.Locale) []core.UIComponent {
	return []core.UIComponent{
		{
			Name:        "channel_url",
			Label:       "YouTube Channel URL",
			Description: "e.g., https://youtube.com/@Squeezie or https://youtube.com/channel/UC...",
			Type:        core.ComponentTypeString,
			Required:    true,
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
			Label:    "Message (variables: {title}, {url}, {channel.name})",
			Type:     core.ComponentTypeTextarea,
			Required: true,
			Variables: []core.Variables{
				{Key: "channel.name", Label: "Channel Name", Description: "The name of the YouTube channel"},
				{Key: "title", Label: "Video Title", Description: "The title of the new video or live"},
				{Key: "url", Label: "Video URL", Description: "The link to the video"},
			},
		},
		{
			Name:        "button_text",
			Label:       "Button Text (Optional)",
			Description: "Leave empty for default (e.g., 'Voir la vidéo')",
			Type:        core.ComponentTypeString,
			Required:    false,
		},
		{
			Name:     "notify_videos",
			Label:    "Notify on New Videos",
			Type:     core.ComponentTypeBoolean,
			Required: false,
		},
		{
			Name:     "notify_shorts",
			Label:    "Notify on New Shorts",
			Type:     core.ComponentTypeBoolean,
			Required: false,
		},
		{
			Name:     "notify_lives",
			Label:    "Notify on Live Streams",
			Type:     core.ComponentTypeBoolean,
			Required: false,
		},
	}
}

func (y *YouTubeProvider) Init(bot *core.Bot, db *gorm.DB) error {
	y.db = db
	go y.renewSubscriptions(bot)
	return nil
}

func (y *YouTubeProvider) renewSubscriptions(bot *core.Bot) {
	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		y.subscribeAll()
		<-ticker.C
	}
}

func (y *YouTubeProvider) subscribeAll() {
	var activeSettings []SocialNotifsSettings
	y.db.Where("youtube_enabled = ?", true).Find(&activeSettings)

	apiURL := os.Getenv("SITE_URL")

	for _, settings := range activeSettings {
		for _, flux := range settings.YouTube.Fluxs {
			channelID := ExtractChannelID(flux.ChannelURL)
			if channelID == "" {
				continue
			}

			topicURL := fmt.Sprintf("https://www.youtube.com/xml/feeds/videos.xml?channel_id=%s", channelID)
			callbackURL := fmt.Sprintf("%s/api/webhooks/youtube", apiURL)

			data := url.Values{}
			data.Set("hub.callback", callbackURL)
			data.Set("hub.topic", topicURL)
			data.Set("hub.verify", "sync")
			data.Set("hub.mode", "subscribe")

			secret := os.Getenv("SOCIALNOTIFS_KEY")
			if secret != "" {
				data.Set("hub.secret", secret)
			}

			resp, err := http.PostForm("https://pubsubhubbub.appspot.com/subscribe", data)
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	}
}

func ExtractChannelID(channelURL string) string {
	if strings.Contains(channelURL, "/channel/UC") {
		parts := strings.Split(channelURL, "/channel/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "/")[0]
		}
	}

	resp, err := http.Get(channelURL)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	bodyStr := string(bodyBytes)

	re := regexp.MustCompile(`itemprop="channelId" content="([^"]+)"`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) > 1 {
		return matches[1]
	}

	re2 := regexp.MustCompile(`"externalId":"([^"]+)"`)
	matches2 := re2.FindStringSubmatch(bodyStr)
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
}
