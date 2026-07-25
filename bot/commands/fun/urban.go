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
	maxLength := 50
	handlers.RegisterCommand(handlers.Command{
		Name:        "urban",
		Description: "Get the Urban definition of a word",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "urban",
			Description: "Get the Urban definition of a word",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "word",
					Description: "The word",
					Required:    true,
					MaxLength:   &maxLength,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			word, _ := event.SlashCommandInteractionData().OptString("word")
			trad := locales.GetUrban(event.Locale())

			res, _ := client.GetUrban(&plume.GetUrbanParams{
				Word: word,
			})

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Definition, res.Definition)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Example, res.Example)),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
					discord.NewTextDisplayf(trad.Author, res.Author),
					discord.NewActionRow(
						discord.NewLinkButton(trad.Source, res.Url),
					),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
