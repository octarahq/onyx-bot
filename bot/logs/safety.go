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

func (l Logger) SendSafetyBlockedInviteLogs(urls []*url.URL, message discord.Message) Code {
	code := GenerateCode(time.Now(), "safety", "invite")
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
			discord.NewTextDisplayf("## Blocked Invite `%s`", code),
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

func (l Logger) SendSafetyZalgoLogs(ratio float64, message discord.Message) Code {
	code := GenerateCode(time.Now(), "safety", "zalgo")
	guild, exist := l.Client.Caches.Guild(*message.GuildID)
	var guildName string = "unknown"
	if exist {
		guildName = guild.Name
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplayf("## Blocked Zalgo `%s`", code),
			discord.NewTextDisplayf("Ratio  : %.02f", ratio),
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

func (l Logger) SendSafetyMentionSpamLogs(message discord.Message) Code {
	code := GenerateCode(time.Now(), "safety", "mentions")
	guild, exist := l.Client.Caches.Guild(*message.GuildID)
	var guildName string = "unknown"
	if exist {
		guildName = guild.Name
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplayf("## Mention Spam `%s`", code),
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

func (l Logger) SendSafetyRaidBotLogs(member discord.Member) Code {
	code := GenerateCode(time.Now(), "safety", "bot")
	guild, exist := l.Client.Caches.Guild(member.GuildID)
	var guildName string = "unknown"
	if exist {
		guildName = guild.Name
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplayf("## Anti Bot `%s`", code),
			discord.NewTextDisplayf("Bot : %s (<@%s>)", member.User.Username, member.User.ID),
			discord.NewTextDisplayf("Guild  : %s (%s)", guildName, guild.ID),
		),
	)

	l.SendLog(SafetyLogsChannel, msg)

	return code
}

func (l Logger) SendSafetyRaidAltLogs(member discord.Member) Code {
	code := GenerateCode(time.Now(), "safety", "alt")
	guild, exist := l.Client.Caches.Guild(member.GuildID)
	var guildName string = "unknown"
	if exist {
		guildName = guild.Name
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplayf("## Anti Bot `%s`", code),
			discord.NewTextDisplayf("Bot : %s (<@%s>)", member.User.Username, member.User.ID),
			discord.NewTextDisplayf("Guild  : %s (%s)", guildName, guild.ID),
		),
	)

	l.SendLog(SafetyLogsChannel, msg)

	return code
}
