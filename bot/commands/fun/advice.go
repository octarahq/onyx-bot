package commands

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"

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

			loc := plume.Def4("en")
			if event.Locale() == discord.LocaleFrench {
				loc = plume.Def4("fr")
			}

			res, _ := client.GetAdvice(&plume.GetAdviceParams{
				Locale: &loc,
			})
			trad := locales.GetAdvice(event.Locale())

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Advice, res.Advice)),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
