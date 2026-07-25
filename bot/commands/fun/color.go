package fun

import (
	"fmt"
	"math/rand"
	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	plume "github.com/lotus64yt/goplume/api/v1"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "color",
		Description: "Get color info",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "color",
			Description: "Get color info",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "info",
					Description: "Get color info",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "hex",
							Description: "A valid hexadecimal color (without the #)",
							Required:    true,
							MinLength:   func() *int { i := 6; return &i }(),
							MaxLength:   func() *int { i := 6; return &i }(),
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "random",
					Description: "Get random color info",
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()

			var color string
			switch *cmd.SubCommandName {
			case "info":
				color, _ = cmd.OptString("hex")
			case "random":
				color = fmt.Sprintf("%06x", rand.Intn(0xFFFFFF))
			}

			res, _ := client.GetColor(&plume.GetColorParams{
				Hex: &color,
			})
			trad := locales.GetColor(event.Locale())

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(fmt.Sprintf("# %s", res.Name)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Hex, res.Hex.Value)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Rgb, res.Rgb.Value)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Hsl, res.Hsl.Value)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Hsv, res.Hsv.Value)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Cmyk, res.Cmyk.Value)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Decimal, res.Int)),
					).WithAccessory(discord.NewThumbnail(res.Url)),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
