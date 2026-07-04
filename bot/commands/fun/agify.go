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
		Name:        "agify",
		Description: "Get user age",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "agify",
			Description: "Get user age",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The user",
					Required:    false,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()

			user, exist := event.SlashCommandInteractionData().OptUser("user")
			if !exist {
				user = event.User()
			}

			res, _ := client.GetAgify(&plume.GetAgifyParams{
				Name: user.Username,
			})
			trad := locales.GetAgify(event.Locale())

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Age, user.ID, res.Age)),
					).WithAccessory(discord.NewThumbnail(*user.AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
