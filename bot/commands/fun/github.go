package commands

import (
	"onyx/bot/core"
	"onyx/bot/handlers"

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
							discord.NewTextDisplayf("# Github User : %s", res.Username),
						).WithAccessory(discord.NewThumbnail(res.AvatarUrl)),
						discord.NewActionRow(
							discord.NewLinkButton("User page", res.PageUrl),
						),
					),
				)
			case "repository":
				repository, _ := cmd.OptString("repository")
				res, _ := client.GetGithubRepository(&plume.GetGithubRepositoryParams{
					Name: repository,
				})

				descStr, _ := res.Description.AsGetGithubRepository200JSONResponseBodyDescription0()
				description := "No description provided."
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
				language := "Unknown"
				if res.Language != nil {
					language = *res.Language
				}
				license := "None"
				if res.LicenseName != nil {
					license = *res.LicenseName
				}

				msg = discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplayf("# Github repository : %s", res.FullName),
							discord.NewTextDisplayf("> %s\n\n**Stars:** %d\n**Forks:** %d\n**Issues:** %d\n**Language:** %s", description, stars, forks, issues, language),
						).WithAccessory(discord.NewThumbnail(res.OwnerAvatarUrl)),
						discord.NewTextDisplayf("-# License : %s", license),
						discord.NewActionRow(
							discord.NewLinkButton("Repository", res.Url),
						),
					),
				)
			}

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
