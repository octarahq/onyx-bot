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
		Name:        "upsidedown",
		Description: "Put a text upside down",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "upsidedown",
			Description: "Put a text upside down",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "text",
					Description: "The text",
					Required:    true,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			text, _ := event.SlashCommandInteractionData().OptString("text")

			res, _ := client.GetUpsideDown(&plume.GetUpsideDownParams{
				Text: text,
			})
			trad := locales.GetUpsidedown(event.Locale())

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Text, res.Text)),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
