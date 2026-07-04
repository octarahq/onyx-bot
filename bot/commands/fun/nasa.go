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
			trad := locales.GetNasa(event.Locale())

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
						discord.NewTextDisplay("# " + trad.Iss_title),
						discord.NewTextDisplayf(trad.Iss_position+"\n"+trad.Iss_altitude+"\n("+trad.Iss_updated+" <t:%d:R>)", res.Latitude, res.Longitude, res.Altitude, res.Velocity, int64(res.Timestamp)),
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
						discord.NewTextDisplay("# " + trad.Apod_title),
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
							discord.NewLinkButton(trad.Source, res.Url),
						),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
