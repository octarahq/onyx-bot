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
		Name:        "math",
		Description: "Calculate the result of a mathematical expression",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "math",
			Description: "Get answer with 8ball",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "expr",
					Description: "Calculate the result of a mathematical expression",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "expr",
							Description: "The expression to calculate",
							Required:    true,
							MaxLength:   func() *int { i := 100; return &i }(),
						},
					},
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()
			trad := locales.GetMath(event.Locale())

			var msg discord.MessageCreate
			switch *cmd.SubCommandName {
			case "expr":
				expr, _ := cmd.OptString("expr")

				res, _ := client.GetMath(&plume.GetMathParams{
					Expr: expr,
				})

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("# "+trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Expression, expr)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Result, res.Result)),
					),
				)
			}
			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
