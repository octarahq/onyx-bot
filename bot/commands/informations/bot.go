package informations

import (
	"fmt"

	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/joho/godotenv"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "bot",
		Description: "Get bot info",
		Category:    "Informations",
		Create: discord.SlashCommandCreate{
			Name:        "bot",
			Description: "Get bot info",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			_ = godotenv.Load()
			msg, err := BuildBotInfoMessage(b, event.Client(), event.Locale())
			if err != nil {
				event.CreateMessage(msg)
				return
			}

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
