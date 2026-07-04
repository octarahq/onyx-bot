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
		Name:        "eightball",
		Description: "Get answer with 8ball",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "eightball",
			Description: "Get answer with 8ball",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "question",
					Description: "Your question to ask",
					Required:    true,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			res, _ := client.Get8ball(&plume.Get8ballParams{})

			question, _ := event.SlashCommandInteractionData().OptString("question")

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay("# 8Ball"),
						discord.NewTextDisplay(fmt.Sprintf("Your question : %s", question)),
						discord.NewTextDisplay(fmt.Sprintf("> My answer : %s", res.Answer)),
					).WithAccessory(discord.NewThumbnail(*event.User().AvatarURL())),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
