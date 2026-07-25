package fun

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
		Name:        "quote",
		Description: "Get a motivation quote",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "quote",
			Description: "Get a motivation quote",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()

			loc := plume.Def4("en")
			if event.Locale() == discord.LocaleFrench {
				loc = plume.Def4("fr")
			}

			res, _ := client.GetQuote(&plume.GetQuoteParams{
				Locale: &loc,
			})
			trad := locales.GetQuote(event.Locale())

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Quote, res.Quote)),
						discord.NewTextDisplayf(trad.Author, res.Author),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
