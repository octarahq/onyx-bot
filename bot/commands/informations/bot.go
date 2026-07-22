package informations

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/joho/godotenv"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "bot",
		Description: "Get bot info",
		Category:    "Informations",
		Create: discord.SlashCommandCreate{
			Name:        "bot",
			Description: "Get bot info",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			_ = godotenv.Load()
			t := locales.GetBot(event.Locale())

			self, exist := event.Client().Caches.SelfUser()
			if !exist {
				event.CreateMessage(discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay(t.Error_self_user),
					),
				).WithFlags(discord.MessageFlagEphemeral))
				return
			}

			goVersion := runtime.Version()
			disgoVersion := "unknown"
			if bi, ok := debug.ReadBuildInfo(); ok {
				for _, d := range bi.Deps {
					if d.Path == "github.com/disgoorg/disgo" {
						disgoVersion = d.Version
						break
					}
				}
			}

			dbPath := "bot.db"
			var dbSize int64
			if info, err := os.Stat(dbPath); err == nil {
				dbSize = info.Size()
			}

			serversCount := event.Client().Caches.GuildsLen()
			usersCount := event.Client().Caches.MembersAllLen()
			commandsCount := len(handlers.Commands)

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplayf(t.Title, self.EffectiveName()),
						discord.NewTextDisplay(t.Description),
					).WithAccessory(discord.NewThumbnail(self.EffectiveAvatarURL())),
					discord.NewTextDisplay(t.Details_title),
					discord.NewTextDisplayf(t.Version_onyx, os.Getenv("VERSION")),
					discord.NewTextDisplayf(t.Last_boot, utils.GenerateTimestamp(int(b.ConnectedSince.Unix()), utils.TimestampLongDate)),
					discord.NewSeparator(discord.SeparatorSpacingSizeLarge),
					discord.NewTextDisplay(t.Programming_title),
					discord.NewTextDisplayf(t.Versions, goVersion, disgoVersion),
					discord.NewTextDisplayf(t.Db_size, utils.ParseUnit(dbSize, "AUTO")),
					discord.NewSeparator(discord.SeparatorSpacingSizeLarge),
					discord.NewTextDisplay(t.Stats_title),
					discord.NewTextDisplayf(t.Stats, serversCount, usersCount, commandsCount),
					discord.NewSeparator(discord.SeparatorSpacingSizeLarge),
					discord.NewActionRow(
						discord.NewLinkButton(t.Website, os.Getenv("SITE_URL")),
						discord.NewLinkButton(t.Invite_me, fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&scope=bot+applications.commands&permissions=1099511627767", os.Getenv("DISCORD_CLIENT_ID"))),
					),
				),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
