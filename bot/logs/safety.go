package logs

import (
	"net/url"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
)

func (l Logger) SendSafetyPhishingLogs(urls []*url.URL, message discord.Message) Code {
	code := GenerateCode(time.Now(), "safety", "phishing")
	guild, exist := l.Client.Caches.Guild(*message.GuildID)
	var guildName string = "unknown"
	if exist {
		guildName = guild.Name
	}

	linkList := make([]string, 0, len(urls))
	for _, u := range urls {
		linkList = append(linkList, u.String())
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplayf("## Anti Phishing `%s`", code),
			discord.NewTextDisplayf("Links  : %s", strings.Join(linkList, ",")),
			discord.NewTextDisplayf("Author : %s (<@%s>)", message.Author.Username, message.Author.ID),
			discord.NewTextDisplayf("Guild  : %s (%s)", guildName, message.GuildID),
		),
		discord.NewContainer(
			discord.NewTextDisplayf("```%s```", message.Content),
		),
	)

	l.SendLog(SafetyLogsChannel, msg)

	return code
}
