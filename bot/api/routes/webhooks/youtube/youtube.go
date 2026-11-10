package youtube

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"onyx/bot/api"
	"onyx/bot/core"
	"onyx/bot/modules/socialnotifs"
)

func init() {
	api.AddRoute(api.Route{
		Method:  "GET",
		Path:    "/webhooks/youtube",
		Handler: handleVerify,
	})
	api.AddRoute(api.Route{
		Method:  "POST",
		Path:    "/webhooks/youtube",
		Handler: handlePayload,
	})
}

func handleVerify(c *gin.Context) {
	challenge := c.Query("hub.challenge")
	if challenge != "" {
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusBadRequest)
}

func handlePayload(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	defer c.Request.Body.Close()

	secret := os.Getenv("SOCIALNOTIFS_KEY")
	if secret != "" {
		signatureHeader := c.GetHeader("X-Hub-Signature")
		if signatureHeader == "" {
			c.Status(http.StatusForbidden)
			return
		}

		parts := strings.Split(signatureHeader, "=")
		if len(parts) != 2 || parts[0] != "sha1" {
			c.Status(http.StatusForbidden)
			return
		}

		mac := hmac.New(sha1.New, []byte(secret))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
			c.Status(http.StatusForbidden)
			return
		}
	}

	type Entry struct {
		VideoId   string `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
		ChannelId string `xml:"http://www.youtube.com/xml/schemas/2015 channelId"`
		Title     string `xml:"title"`
		Link      struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Author struct {
			Name string `xml:"name"`
		} `xml:"author"`
		Published string `xml:"published"`
		Updated   string `xml:"updated"`
	}

	type Feed struct {
		Entry []Entry `xml:"entry"`
	}

	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		c.Status(http.StatusOK)
		return
	}

	if len(feed.Entry) == 0 {
		c.Status(http.StatusOK)
		return
	}

	entry := feed.Entry[0]

	b, ok := c.Get("bot")
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}
	bot := b.(*core.Bot)

	isShort := strings.Contains(entry.Link.Href, "/shorts/")

	var activeSettings []socialnotifs.SocialNotifsSettings
	bot.DB.GormDB.Where("youtube_enabled = ?", true).Find(&activeSettings)

	for _, settings := range activeSettings {
		for _, flux := range settings.YouTube.Fluxs {
			fluxChannelID := socialnotifs.ExtractChannelID(flux.ChannelURL)

			if fluxChannelID != entry.ChannelId {
				continue
			}

			if isShort && !flux.NotifyShorts {
				continue
			}
			if !isShort && !flux.NotifyVideos {
				continue
			}

			buttonText := flux.ButtonText
			if buttonText == "" {
				if isShort {
					buttonText = "Watch Short"
				} else {
					buttonText = "Watch Video"
				}
			}

			msg := flux.Message
			msg = strings.ReplaceAll(msg, "{channel.name}", entry.Author.Name)
			msg = strings.ReplaceAll(msg, "{title}", entry.Title)
			msg = strings.ReplaceAll(msg, "{url}", entry.Link.Href)

			socialnotifs.DispatchMessage(bot, socialnotifs.FluxMessage{
				ChannelID: flux.Channel,
				Content:   msg,
				Link:      entry.Link.Href,
				LinkLabel: buttonText,
			})
		}
	}

	c.Status(http.StatusOK)
}
