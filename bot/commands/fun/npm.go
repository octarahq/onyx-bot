package commands

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	plume "github.com/lotus64yt/goplume/api/v1"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "npm",
		Description: "Get information about an NPM package",
		Category:    "Utils",
		Create: discord.SlashCommandCreate{
			Name:        "npm",
			Description: "Get information about an NPM package",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "The name of the package",
					Required:    true,
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()
			pkgName, _ := cmd.OptString("name")
			trad := locales.GetNpm(event.Locale())

			res, err := client.GetNpm(&plume.GetNpmParams{
				Name: pkgName,
			})

			var msg discord.MessageCreate

			if err != nil || res == nil {
				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplayf(trad.Not_found, pkgName),
					),
				)
			} else {
				author := res.AuthorUsername
				if author == "" {
					author = trad.Unknown
				}

				keywords := trad.None
				if len(res.Keywords) > 0 {
					keywords = strings.Join(res.Keywords, ", ")
				}

				if len(keywords) > 100 {
					keywords = keywords[:97] + "..."
				}

				buttons := []discord.InteractiveComponent{
					discord.NewLinkButton(trad.Npm_page, res.NpmUrl),
				}
				if res.RepositoryUrl != nil && strings.HasPrefix(*res.RepositoryUrl, "http") {
					buttons = append(buttons, discord.NewLinkButton(trad.Repository, *res.RepositoryUrl))
				}

				downloadsYearly := trad.Na
				if res.DownloadsYearly != nil {
					downloadsYearly = fmt.Sprintf("%d", *res.DownloadsYearly)
				}

				downloadsMonthly := fmt.Sprintf("%d", res.DownloadsMonthly)
				downloadsWeekly := fmt.Sprintf("%d", res.DownloadsWeekly)

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplayf("# "+trad.Title, res.Name, res.Version),
						discord.NewTextDisplayf(trad.Desc, res.Description, author, keywords, res.Dependents),
						discord.NewTextDisplayf(trad.Downloads, downloadsWeekly, downloadsMonthly, downloadsYearly),
						discord.NewTextDisplayf(trad.Last_published, int64(res.LastPublished)),
						discord.NewActionRow(buttons...),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
