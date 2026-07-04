package commands

import (
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	plume "github.com/lotus64yt/goplume/api/v1"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "emoji",
		Description: "Emoji commands",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "emoji",
			Description: "Emoji commands",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "emojify",
					Description: "Emojify some text",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "text",
							Description: "The text to emojify",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommandGroup{
					Name:        "mix",
					Description: "Mix emojis",
					Options: []discord.ApplicationCommandOptionSubCommand{
						{
							Name:        "custom",
							Description: "Mix two emojis",
							Options: []discord.ApplicationCommandOption{
								discord.ApplicationCommandOptionString{
									Name:        "emoji1",
									Description: "First emoji",
									Required:    true,
								},
								discord.ApplicationCommandOptionString{
									Name:        "emoji2",
									Description: "Second emoji",
									Required:    true,
								},
							},
						},
						{
							Name:        "random",
							Description: "Mix two random emojis",
						},
					},
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()

			var msg discord.MessageCreate

			switch *cmd.SubCommandName {
			case "emojify":
				text, _ := cmd.OptString("text")
				res, err := client.GetEmojify(&plume.GetEmojifyParams{
					Text: text,
				})
				if err != nil || res == nil {
					msg = discord.NewMessageCreateV2(
						discord.NewContainer(
							discord.NewTextDisplay("Error emojifying text."),
						),
					)
				} else {
					msg = discord.NewMessageCreateV2(
						discord.NewContainer(
							discord.NewTextDisplay("# Emojify"),
							discord.NewTextDisplay(res.Text),
						),
					)
				}

			case "custom":
				emoji1, _ := cmd.OptString("emoji1")
				emoji2, _ := cmd.OptString("emoji2")
				res, err := client.GetEmojiMix(&plume.GetEmojiMixParams{
					Left:  emoji1,
					Right: emoji2,
				})
				if err != nil || res == nil {
					msg = discord.NewMessageCreateV2(
						discord.NewContainer(
							discord.NewTextDisplay("Error mixing emojis. They might not be compatible!"),
						),
					)
				} else {
					msg = discord.NewMessageCreateV2(
						discord.NewContainer(
							discord.NewTextDisplay("# Emoji Mix"),
							discord.NewMediaGallery(
								discord.MediaGalleryItem{
									Media: discord.UnfurledMediaItem{
										URL: res.EmojiUrl,
									},
								},
							),
						),
					)
				}

			case "random":
				res, err := client.GetRandomEmojiMix()
				if err != nil || res == nil {
					msg = discord.NewMessageCreateV2(
						discord.NewContainer(
							discord.NewTextDisplay("Error getting random emoji mix."),
						),
					)
				} else {
					msg = discord.NewMessageCreateV2(
						discord.NewContainer(
							discord.NewTextDisplay("# Random Emoji Mix"),
							discord.NewMediaGallery(
								discord.MediaGalleryItem{
									Media: discord.UnfurledMediaItem{
										URL: res.EmojiUrl,
									},
								},
							),
						),
					)
				}
			}

			if err := event.CreateMessage(msg); err != nil {
				// Failed to send message, nothing we can do here usually
			}
		},
	})
}
