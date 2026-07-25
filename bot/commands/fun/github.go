package fun

import (
	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	plume "github.com/lotus64yt/goplume/api/v1"
)

func init() {
	maxLength := 100
	handlers.RegisterCommand(handlers.Command{
		Name:        "github",
		Description: "Search github data",
		Category:    "Fun",
		Create: discord.SlashCommandCreate{
			Name:        "github",
			Description: "Search github data",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "user",
					Description: "Search a GitHub user data",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "user",
							Description: "The user name",
							Required:    true,
							MaxLength:   &maxLength,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "repository",
					Description: "Search a GitHub repository data",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "repository",
							Description: "The repository name",
							Required:    true,
							MaxLength:   &maxLength,
						},
					},
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			client, _ := plume.NewAPIClient()
			cmd := event.SlashCommandInteractionData()
			trad := locales.GetGithub(event.Locale())

			var msg discord.MessageCreate
			switch *cmd.SubCommandName {
			case "user":
				user, _ := cmd.OptString("user")
				res, _ := client.GetGithubUser(&plume.GetGithubUserParams{
					Name: user,
				})

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplayf("# "+trad.User_title, res.Username),
						).WithAccessory(discord.NewThumbnail(res.AvatarUrl)),
						discord.NewActionRow(
							discord.NewLinkButton(trad.User_page, res.PageUrl),
						),
					),
				)
			case "repository":
				repository, _ := cmd.OptString("repository")
				res, _ := client.GetGithubRepository(&plume.GetGithubRepositoryParams{
					Name: repository,
				})

				descStr, _ := res.Description.AsGetGithubRepository200JSONResponseBodyDescription0()
				description := trad.No_desc
				if descStr != "" {
					description = descStr
				}

				stars := 0
				if res.StargazersCount != nil {
					stars = int(*res.StargazersCount)
				}
				forks := 0
				if res.ForksCount != nil {
					forks = int(*res.ForksCount)
				}
				issues := 0
				if res.OpenIssuesCount != nil {
					issues = int(*res.OpenIssuesCount)
				}
				language := trad.Unknown
				if res.Language != nil {
					language = *res.Language
				}
				license := trad.None
				if res.LicenseName != nil {
					license = *res.LicenseName
				}

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplayf("# "+trad.Repo_title, res.FullName),
							discord.NewTextDisplayf(trad.Repo_desc, description, stars, forks, issues, language),
						).WithAccessory(discord.NewThumbnail(res.OwnerAvatarUrl)),
						discord.NewTextDisplayf(trad.Repo_license, license),
						discord.NewActionRow(
							discord.NewLinkButton(trad.Repo_button, res.Url),
						),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
