package commands

import (
	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	plume "github.com/lotus64yt/goplume/api/v1"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "ascii",
		Description: "Create ascii art",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "ascii",
			Description: "Create ascii art",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "image",
					Description: "Create an ASCII Image",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "text",
							Description: "The text",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "font",
							Description: "The font for the art",
							Required:    false,
							Choices: []discord.ApplicationCommandOptionChoiceString{
								{
									Name:  "Standard",
									Value: "standard",
								},
								{
									Name:  "Isometric",
									Value: "isometric",
								},
								{
									Name:  "3D",
									Value: "3d",
								},
								{
									Name:  "Letters",
									Value: "letters",
								},
							},
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "text",
					Description: "Create an ASCII text",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "text",
							Description: "The text",
							Required:    true,
						},
					},
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()
			trad := locales.GetAscii(event.Locale())

			var msg discord.MessageCreate
			switch *cmd.SubCommandName {
			case "text":
				text, _ := cmd.OptString("text")
				res, _ := client.GetAsciiText(&plume.GetAsciiTextParams{
					Text: text,
				})

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(res.Text),
					),
				)

			case "image":
				text, _ := cmd.OptString("text")
				font, _ := cmd.OptString("font")

				var fontPtr *plume.GetAsciiImageParamsFont
				if font != "" {
					f := plume.GetAsciiImageParamsFont(font)
					fontPtr = &f
				}

				req, _ := plume.NewGetAsciiImageRequest("https://plume.voctal.dev/api", &plume.GetAsciiImageParams{
					Text: text,
					Font: fontPtr,
				})

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewMediaGallery(
							discord.MediaGalleryItem{
								Media: discord.UnfurledMediaItem{
									URL: req.URL.String(),
								},
							},
						),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
