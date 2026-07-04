package commands

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"
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

			res, err := client.GetNpm(&plume.GetNpmParams{
				Name: pkgName,
			})

			var msg discord.MessageCreate

			if err != nil || res == nil {
				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay("Could not find package **" + pkgName + "** on NPM."),
					),
				)
			} else {
				author := res.AuthorUsername
				if author == "" {
					author = "Unknown"
				}

				keywords := "None"
				if len(res.Keywords) > 0 {
					keywords = strings.Join(res.Keywords, ", ")
				}

				if len(keywords) > 100 {
					keywords = keywords[:97] + "..."
				}

				buttons := []discord.InteractiveComponent{
					discord.NewLinkButton("NPM Page", res.NpmUrl),
				}
				if res.RepositoryUrl != nil && strings.HasPrefix(*res.RepositoryUrl, "http") {
					buttons = append(buttons, discord.NewLinkButton("Repository", *res.RepositoryUrl))
				}

				downloadsYearly := "N/A"
				if res.DownloadsYearly != nil {
					downloadsYearly = fmt.Sprintf("%d", *res.DownloadsYearly)
				}

				downloadsMonthly := fmt.Sprintf("%d", res.DownloadsMonthly)
				downloadsWeekly := fmt.Sprintf("%d", res.DownloadsWeekly)

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplayf("# NPM Package: %s (v%s)", res.Name, res.Version),
						discord.NewTextDisplayf("> %s\n\n**Author:** %s\n**Keywords:** %s\n**Dependents:** %d", res.Description, author, keywords, res.Dependents),
						discord.NewTextDisplayf("**Downloads:**\n- **Weekly:** %s\n- **Monthly:** %s\n- **Yearly:** %s", downloadsWeekly, downloadsMonthly, downloadsYearly),
						discord.NewTextDisplayf("**Last Published:** <t:%d:R>", int64(res.LastPublished)),
						discord.NewActionRow(buttons...),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
