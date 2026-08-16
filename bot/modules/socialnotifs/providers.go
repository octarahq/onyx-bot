package socialnotifs

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

type FluxMessage struct {
	ChannelID string
	Content   string
	Link      string
	LinkLabel string
}

func DispatchMessage(bot *core.Bot, msg FluxMessage) {
	if msg.ChannelID == "" || msg.Content == "" {
		return
	}

	components := []discord.ContainerSubComponent{
		discord.NewTextDisplay(msg.Content),
	}

	if msg.Link != "" {
		label := msg.LinkLabel
		if label == "" {
			label = "Open link"
		}
		actionRow := discord.NewActionRow(
			discord.NewLinkButton(label, msg.Link),
		)
		components = append(components, actionRow)
	}

	payload := discord.NewMessageCreateV2(discord.NewContainer(components...))
	bot.SendMessage(msg.ChannelID, payload)
}
