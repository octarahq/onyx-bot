package freegames

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type Provider interface {
	Name() string
	Init(bot *core.Bot, db *gorm.DB) error
	UISchema(locale discord.Locale) []core.UIComponent
}

var Providers = []Provider{}

func RegisterProvider(p Provider) {
	Providers = append(Providers, p)
}

type FreeGameInfo struct {
	Title       string
	Description string
	Thumbnail   string
	URL         string
	Worth       string
}

type FluxMessage struct {
	ChannelID string
	Content   string
	GameInfo  FreeGameInfo
	LinkLabel string
}

func DispatchMessage(bot *core.Bot, msg FluxMessage) {
	if msg.ChannelID == "" {
		return
	}

	var texts []discord.SectionSubComponent
	texts = append(texts, discord.NewTextDisplayf("## %s", msg.GameInfo.Title))
	texts = append(texts, discord.NewTextDisplay(msg.GameInfo.Description))

	if msg.GameInfo.Worth != "" && msg.GameInfo.Worth != "N/A" {
		texts = append(texts, discord.NewTextDisplayf("**Worth:** %s", msg.GameInfo.Worth))
	}

	section := discord.NewSection(texts...)

	if msg.GameInfo.Thumbnail != "" {
		section = section.WithAccessory(discord.NewThumbnail(msg.GameInfo.Thumbnail))
	}

	var containerComps []discord.ContainerSubComponent
	containerComps = append(containerComps, section)

	if msg.GameInfo.URL != "" {
		label := msg.LinkLabel
		if label == "" {
			label = "Get Game"
		}
		containerComps = append(containerComps, discord.NewActionRow(
			discord.NewLinkButton(label, msg.GameInfo.URL),
		))
	}

	payload := discord.NewMessageCreateV2(discord.NewContainer(containerComps...))
	payload.Content = msg.Content

	bot.SendMessage(msg.ChannelID, payload)
}
