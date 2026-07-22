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

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

func BuildBotInfoMessage(b *core.Bot, client *bot.Client, locale discord.Locale) (discord.MessageCreate, error) {
	t := locales.GetBot(locale)

	self, exist := client.Caches.SelfUser()
	if !exist {
		return discord.NewMessageCreateV2(
			discord.NewContainer(
				discord.NewTextDisplay(t.Error_self_user),
			),
		).WithFlags(discord.MessageFlagEphemeral), fmt.Errorf("self user not found")
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

	serversCount := client.Caches.GuildsLen()
	usersCount := client.Caches.MembersAllLen()
	commandsCount := len(handlers.Commands)

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplayf(t.Title, self.EffectiveName()),
				discord.NewTextDisplay(t.Description),
			).WithAccessory(discord.NewThumbnail(self.EffectiveAvatarURL())),
			discord.NewTextDisplay(t.Details_title),
			discord.NewTextDisplayf(t.Version_onyx, b.Version),
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
	return msg, nil
}
