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
		Name:        "nasa",
		Description: "Get nasa data",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "nasa",
			Description: "Get nasa data",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "iss",
					Description: "Get the current position and velocity of the ISS",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "apod",
					Description: "Get the image of the day.",
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()

			var msg discord.MessageCreate
			switch *cmd.SubCommandName {
			case "iss":
				res, _ := client.GetIss()

				circle := true
				req, _ := plume.NewGetIssImageRequest("https://plume.voctal.dev/api", &plume.GetIssImageParams{
					Circle: &circle,
				})

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# ISS"),
						discord.NewTextDisplayf("The International Space Station is currently located at %.4f, %.4f.\nIt is orbiting at an altitude of %.0f km with a speed of %.0f km/h. (Updated <t:%d:R>)", res.Latitude, res.Longitude, res.Altitude, res.Velocity, int64(res.Timestamp)),
						discord.NewMediaGallery(
							discord.MediaGalleryItem{
								Media: discord.UnfurledMediaItem{
									URL: req.URL.String(),
								},
							},
						),
					),
				)
			case "apod":
				res, _ := client.GetNasaApod()

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# NASA Apod"),
						discord.NewTextDisplayf("### %s", res.Title),
						discord.NewTextDisplayf("> %s", res.Explanation),
						discord.NewMediaGallery(
							discord.MediaGalleryItem{
								Media: discord.UnfurledMediaItem{
									URL: *res.HdUrl,
								},
							},
						),
						discord.NewActionRow(
							discord.NewLinkButton("Source", res.Url),
						),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
