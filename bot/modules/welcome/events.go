package welcome

import (
	"fmt"
	"net/url"
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	plume "github.com/lotus64yt/goplume/api/v1"
)

func (m *WelcomeModule) HandleGuildMemberJoin(b *core.Bot, e *events.GuildMemberJoin) bool {
	if m.Data.Enabled && m.Data.Join.Enabled {
		scid, err := snowflake.Parse(m.Data.Join.Channel)
		if err != nil {
			fmt.Printf("[ERROR] : %s", err)
			return false
		}

		channel, exist := b.Client.Caches.GuildTextChannel(scid)
		if !exist {
			fmt.Printf("[ERROR] : %s", err)
			return false
		}

		guild, exist := e.Client().Caches.Guild(e.GuildID)
		if !exist {
			return false
		}

		userName := e.Member.User.Username
		if e.Member.User.GlobalName != nil {
			userName = *e.Member.User.GlobalName
		}

		vars := map[string]string{
			"user.id":             e.Member.User.ID.String(),
			"user.mention":        fmt.Sprintf("<@%s>", e.Member.User.ID),
			"user.name":           userName,
			"server.name":         guild.Name,
			"server.member_count": strconv.Itoa(guild.MemberCount),
		}

		var addedLinks int
		actionRow := discord.NewActionRow()

		for _, l := range m.Data.Join.Links {
			url, err := url.Parse(l)
			if err != nil {
				continue
			}
			actionRow = actionRow.AddComponents(
				discord.NewLinkButton(url.Host, l),
			)
			addedLinks++
		}

		trad := locales.GetModule_WelcomeModule(discord.Locale(guild.PreferredLocale))

		t1 := trad.Card_t1
		t2 := trad.Card_t2
		t3 := guild.Name
		avatar := e.Member.EffectiveAvatarURL()
		req, _ := plume.NewGetCardsWelcomeRequest("https://plume.voctal.dev/api/api", &plume.GetCardsWelcomeParams{
			Avatar: &avatar,
			Text1:  t1,
			Text2:  &t2,
			Text3:  &t3,
		})

		components := []discord.ContainerSubComponent{
			discord.NewTextDisplay(trad.Embed_title),
			discord.NewTextDisplay(utils.ParseVariables(m.Data.Join.Content, vars)),
			discord.NewMediaGallery(
				discord.MediaGalleryItem{
					Media: discord.UnfurledMediaItem{
						URL: req.URL.String(),
					},
				},
			),
		}

		if addedLinks > 0 {
			components = append(components, actionRow)
		}

		components = append(components, discord.NewTextDisplayf(trad.Embed_footer, guild.MemberCount))

		msg := discord.NewMessageCreateV2(discord.NewContainer(components...))

		if m.Data.Join.MentionMember {
			msg = msg.WithAllowedMentions(&discord.AllowedMentions{
				Users: []snowflake.ID{e.Member.User.ID},
			})
		}

		if _, err := e.Client().Rest.CreateMessage(channel.ID(), msg); err != nil {
			fmt.Printf("[ERROR] : %s", err)
		}
	}
	return false
}

func (m *WelcomeModule) HandleGuildMemberLeave(b *core.Bot, e *events.GuildMemberLeave) bool {
	if m.Data.Enabled && m.Data.Leave.Enabled {
		scid, err := snowflake.Parse(m.Data.Leave.Channel)
		if err != nil {
			fmt.Printf("[ERROR] parsing channel ID: %s\n", err)
			return false
		}

		channel, exist := b.Client.Caches.GuildTextChannel(scid)
		if !exist {
			fmt.Printf("[ERROR] channel not found in cache: %s\n", m.Data.Leave.Channel)
			return false
		}

		guild, exist := e.Client().Caches.Guild(e.GuildID)
		if !exist {
			fmt.Println("[ERROR] guild not found in cache")
			return false
		}

		userName := e.User.Username
		if e.User.GlobalName != nil {
			userName = *e.User.GlobalName
		}

		vars := map[string]string{
			"user.id":             e.User.ID.String(),
			"user.name":           userName,
			"server.name":         guild.Name,
			"server.member_count": strconv.Itoa(guild.MemberCount),
		}

		trad := locales.GetModule_WelcomeModule(discord.Locale(guild.PreferredLocale))

		t1 := trad.Leave_card_t1
		t2 := trad.Leave_card_t2
		t3 := guild.Name
		color := "ff0000"
		avatar := e.User.EffectiveAvatarURL()
		req, _ := plume.NewGetCardsWelcomeRequest("https://plume.voctal.dev/api/api", &plume.GetCardsWelcomeParams{
			Avatar:            &avatar,
			Text1:             t1,
			Text2:             &t2,
			Text3:             &t3,
			AvatarBorderColor: &color,
		})

		components := []discord.ContainerSubComponent{
			discord.NewTextDisplay(trad.Leave_embed_title),
			discord.NewTextDisplay(utils.ParseVariables(m.Data.Leave.Content, vars)),
			discord.NewMediaGallery(
				discord.MediaGalleryItem{
					Media: discord.UnfurledMediaItem{
						URL: req.URL.String(),
					},
				},
			),
			discord.NewTextDisplayf(trad.Leave_embed_footer, guild.MemberCount),
		}

		msg := discord.NewMessageCreateV2(discord.NewContainer(components...))

		if m.Data.Leave.MentionMember {
			msg = msg.WithAllowedMentions(&discord.AllowedMentions{
				Users: []snowflake.ID{e.User.ID},
			})
		}

		if _, err := e.Client().Rest.CreateMessage(channel.ID(), msg); err != nil {
			fmt.Printf("[ERROR] : %s", err)
		}
	}
	return false
}
