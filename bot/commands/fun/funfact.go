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
		Name:        "funfact",
		Description: "Get a random funfact",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "funfact",
			Description: "Get a random fun fact",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()

			loc := plume.Def4("en")
			if event.Locale() == discord.LocaleFrench {
				loc = plume.Def4("fr")
			}

			res, _ := client.GetFunfact(&plume.GetFunfactParams{
				Locale: &loc,
			})
			trad := locales.GetFunfact(event.Locale())

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Text, res.Text)),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
					discord.NewActionRow(
						discord.NewLinkButton(trad.Source, res.SourceUrl),
					),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
