package utils

import (
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "ping",
		Description: "Replies with Pong!",
		Category:    "Utils",
		Create: discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			msg := discord.NewMessageCreate().
				WithContent("Pong!")

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
