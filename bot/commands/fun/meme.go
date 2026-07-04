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
		Name:        "meme",
		Description: "Get a meme",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "meme",
			Description: "Get a meme",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "random",
					Description: "Get a random meme",
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()

			var msg discord.MessageCreate
			switch *cmd.SubCommandName {
			case "random":
				res, _ := client.GetMeme()

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# Meme"),
						discord.NewTextDisplay(fmt.Sprintf("### %s", res.Title)),
						discord.NewMediaGallery(
							discord.MediaGalleryItem{
								Media: discord.UnfurledMediaItem{
									URL: res.ImageUrl,
								},
							},
						),
						discord.NewTextDisplay(fmt.Sprintf("-# %s", res.Author)),
					),
				)
			}
			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
