package modules

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
	"gorm.io/gorm"

	plume "github.com/lotus64yt/goplume/api/v1"
)

type WelcomeJoinSettings struct {
	Enabled       bool     `gorm:"default:false" json:"enabled"`
	Channel       string   `json:"channel"`
	Content       string   `json:"content"`
	Links         []string `gorm:"serializer:json" json:"links"`
	MentionMember bool     `gorm:"default:true" json:"mentionMember"`
}

type WelcomeLeaveSettings struct {
	Enabled       bool   `gorm:"default:false" json:"enabled"`
	Channel       string `json:"channel"`
	Content       string `json:"content"`
	MentionMember bool   `gorm:"default:true" json:"mentionMember"`
}

type WelcomeSettings struct {
	GuildID string               `gorm:"primaryKey" json:"guild_id"`
	Enabled bool                 `gorm:"default:false" json:"enabled"`
	Join    WelcomeJoinSettings  `gorm:"embedded;embeddedPrefix:join_" json:"join"`
	Leave   WelcomeLeaveSettings `gorm:"embedded;embeddedPrefix:leave_" json:"leave"`
}

type WelcomeModule struct {
	Data WelcomeSettings
}

func init() {
	Register(&WelcomeModule{})
}

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

func (m *WelcomeModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "WelcomeModule",
		Icon: "waving_hand",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_WelcomeModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_WelcomeModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_WelcomeModule")
			subs := make(map[string]core.SubmoduleMeta)
			for k, v := range meta.Submodules {
				subs[k] = core.SubmoduleMeta{
					Label:       v.Label,
					Description: v.Description,
				}
			}
			return subs
		},
	}
}

func (m *WelcomeModule) Priority() int   { return 1 }
func (m *WelcomeModule) IsEnabled() bool { return m.Data.Enabled }
func (m *WelcomeModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionSendMessages,
	}
}

func (m *WelcomeModule) Schema() interface{}  { return &WelcomeSettings{} }
func (m *WelcomeModule) DataPtr() interface{} { return &m.Data }
func (m *WelcomeModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = WelcomeSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, WelcomeSettings{GuildID: guildID}).Error
}

func (m *WelcomeModule) UISchema(locale discord.Locale) core.UISchema {
	meta := locales.GetMeta(locale, "module_WelcomeModule")

	joinLabel := "Join Settings"
	joinDesc := ""
	joinEnabledLabel := "Enabled"
	joinChannelLabel := "Channel"
	joinChannelDesc := "The channel where the message will be sent."
	joinContentLabel := "Message Content"
	joinLinksLabel := "Links"
	joinLinksDesc := ""
	joinMentionLabel := "Mention Member"

	joinVars := []core.Variables{
		{Key: "user.id", Label: "User ID", Description: "The ID of the user."},
		{Key: "user.mention", Label: "User Mention", Description: "Mentions the user."},
		{Key: "user.name", Label: "User Name", Description: "The username of the user."},
		{Key: "server.name", Label: "Server Name", Description: "The name of the server."},
		{Key: "server.member_count", Label: "Member Count", Description: "The number of members in the server."},
	}

	if sub, ok := meta.Submodules["join"]; ok {
		if sub.Label != "" {
			joinLabel = sub.Label
		}
		if sub.Description != "" {
			joinDesc = sub.Description
		}
		if opt, ok := sub.Options["enabled"]; ok && opt.Label != "" {
			joinEnabledLabel = opt.Label
		}
		if opt, ok := sub.Options["channel"]; ok {
			if opt.Label != "" {
				joinChannelLabel = opt.Label
			}
			if opt.Description != "" {
				joinChannelDesc = opt.Description
			}
		}
		if opt, ok := sub.Options["content"]; ok {
			if opt.Label != "" {
				joinContentLabel = opt.Label
			}
			for i, v := range joinVars {
				if vOpt, ok := opt.Options[v.Key]; ok {
					if vOpt.Label != "" {
						joinVars[i].Label = vOpt.Label
					}
					if vOpt.Description != "" {
						joinVars[i].Description = vOpt.Description
					}
				}
			}
		}
		if opt, ok := sub.Options["links"]; ok {
			if opt.Label != "" {
				joinLinksLabel = opt.Label
			}
			if opt.Description != "" {
				joinLinksDesc = opt.Description
			}
		}
		if opt, ok := sub.Options["mentionMember"]; ok && opt.Label != "" {
			joinMentionLabel = opt.Label
		}
	}

	leaveLabel := "Leave Settings"
	leaveDesc := ""
	leaveEnabledLabel := "Enabled"
	leaveChannelLabel := "Channel"
	leaveChannelDesc := "The channel where the message will be sent."
	leaveContentLabel := "Message Content"
	leaveLinksLabel := "Links"
	leaveLinksDesc := ""
	leaveMentionLabel := "Mention Member"

	leaveVars := []core.Variables{
		{Key: "user.id", Label: "User ID", Description: "The ID of the user."},
		{Key: "user.name", Label: "User Name", Description: "The username of the user."},
		{Key: "server.name", Label: "Server Name", Description: "The name of the server."},
		{Key: "server.member_count", Label: "Member Count", Description: "The number of members in the server."},
	}

	if sub, ok := meta.Submodules["leave"]; ok {
		if sub.Label != "" {
			leaveLabel = sub.Label
		}
		if sub.Description != "" {
			leaveDesc = sub.Description
		}
		if opt, ok := sub.Options["enabled"]; ok && opt.Label != "" {
			leaveEnabledLabel = opt.Label
		}
		if opt, ok := sub.Options["channel"]; ok {
			if opt.Label != "" {
				leaveChannelLabel = opt.Label
			}
			if opt.Description != "" {
				leaveChannelDesc = opt.Description
			}
		}
		if opt, ok := sub.Options["content"]; ok {
			if opt.Label != "" {
				leaveContentLabel = opt.Label
			}
			for i, v := range leaveVars {
				if vOpt, ok := opt.Options[v.Key]; ok {
					if vOpt.Label != "" {
						leaveVars[i].Label = vOpt.Label
					}
					if vOpt.Description != "" {
						leaveVars[i].Description = vOpt.Description
					}
				}
			}
		}
		if opt, ok := sub.Options["mentionMember"]; ok && opt.Label != "" {
			leaveMentionLabel = opt.Label
		}
	}

	maxContent := 1000
	maxLinks := 5

	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "join",
				Label:       joinLabel,
				Description: joinDesc,
				Components: []core.UIComponent{
					{
						Name:     "enabled",
						Label:    joinEnabledLabel,
						Type:     core.ComponentTypeBoolean,
						Required: false,
					},
					{
						Name:         "channel",
						Label:        joinChannelLabel,
						Description:  joinChannelDesc,
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
					{
						Name:      "content",
						Label:     joinContentLabel,
						Type:      core.ComponentTypeTextarea,
						Max:       &maxContent,
						Required:  true,
						Variables: joinVars,
					},
					{
						Name:        "links",
						Label:       joinLinksLabel,
						Description: joinLinksDesc,
						Type:        core.ComponentTypeList,
						ListType:    "link",
						Max:         &maxLinks,
						Required:    false,
					},
					{
						Name:     "mentionMember",
						Label:    joinMentionLabel,
						Type:     core.ComponentTypeBoolean,
						Required: false,
					},
				},
			},
			{
				Name:        "leave",
				Label:       leaveLabel,
				Description: leaveDesc,
				Components: []core.UIComponent{
					{
						Name:     "enabled",
						Label:    leaveEnabledLabel,
						Type:     core.ComponentTypeBoolean,
						Required: false,
					},
					{
						Name:         "channel",
						Label:        leaveChannelLabel,
						Description:  leaveChannelDesc,
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
					{
						Name:      "content",
						Label:     leaveContentLabel,
						Type:      core.ComponentTypeTextarea,
						Max:       &maxContent,
						Required:  true,
						Variables: leaveVars,
					},
					{
						Name:        "links",
						Label:       leaveLinksLabel,
						Description: leaveLinksDesc,
						Type:        core.ComponentTypeList,
						ListType:    "link",
						Max:         &maxLinks,
						Required:    false,
					},
					{
						Name:     "mentionMember",
						Label:    leaveMentionLabel,
						Type:     core.ComponentTypeBoolean,
						Required: false,
					},
				},
			},
		},
	}
}
