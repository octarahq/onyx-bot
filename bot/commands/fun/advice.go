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
		Name:        "advice",
		Description: "Get a random advice",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "advice",
			Description: "Get a random advice",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			res, _ := client.GetAdvice(&plume.GetAdviceParams{})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# Advice"),
						discord.NewTextDisplay(fmt.Sprintf("> My advice : %s", res.Advice)),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
