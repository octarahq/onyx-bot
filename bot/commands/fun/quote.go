package commands

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	plume "github.com/lotus64yt/goplume/api/v1"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "quote",
		Description: "Get a motivation quote",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "quote",
			Description: "Get a motivation quote",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			res, _ := client.GetQuote(&plume.GetQuoteParams{})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# Quote"),
						discord.NewTextDisplay(fmt.Sprintf("> %s", res.Quote)),
						discord.NewTextDisplayf("-# %s", res.Author),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
