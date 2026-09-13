package infinitecounter

import (
	"errors"
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type MainCounterSettings struct {
	ChannelID string `json:"channel_id"`
	Type      string `gorm:"default:'strict'" json:"type"`
}

type FreeCounterSettings struct {
	RestartOnError bool `gorm:"default:false" json:"restart_on_error"`
	AllowChat      bool `gorm:"default:true" json:"allow_chat"`
}

type StrictCounterSettings struct {
}

type InfiniteCounterSettings struct {
	GuildID      string                `gorm:"primaryKey" json:"guild_id"`
	Enabled      bool                  `gorm:"default:false" json:"enabled"`
	ServerCount  int64                 `gorm:"default:0" json:"-"`
	LastUser     string                `json:"-"`
	UserProgress map[string]int64      `gorm:"serializer:json" json:"-"`
	Main         MainCounterSettings   `gorm:"embedded;embeddedPrefix:main_" json:"main"`
	Free         FreeCounterSettings   `gorm:"embedded;embeddedPrefix:free_" json:"free"`
	Strict       StrictCounterSettings `gorm:"embedded;embeddedPrefix:strict_" json:"strict"`
}

type InfiniteCounterModule struct {
	Data InfiniteCounterSettings
}

func init() {
	core.Register(&InfiniteCounterModule{})
}

func (m *InfiniteCounterModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "InfiniteCounterModule",
		Icon: "format_list_numbered",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_InfiniteCounterModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_InfiniteCounterModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_InfiniteCounterModule")
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

func (m *InfiniteCounterModule) Priority() int   { return 1 }
func (m *InfiniteCounterModule) IsEnabled() bool { return m.Data.Enabled }
func (m *InfiniteCounterModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionSendMessages,
		discord.PermissionManageMessages,
		discord.PermissionManageChannels,
	}
}

func (m *InfiniteCounterModule) Schema() interface{}  { return &InfiniteCounterSettings{} }
func (m *InfiniteCounterModule) DataPtr() interface{} { return &m.Data }
func (m *InfiniteCounterModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = InfiniteCounterSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, InfiniteCounterSettings{GuildID: guildID}).Error
}

func CreatePanel(b *core.Bot, currentCount int, name, avatarURL string) discord.MessageCreate {
	return discord.NewMessageCreate().AddEmbeds(
		discord.Embed{
			Color:       utils.ParseStrColor("#2c92b8"),
			Description: fmt.Sprintf("## %d", currentCount),
			Author: &discord.EmbedAuthor{
				Name:    name,
				IconURL: avatarURL,
			},
		},
	).AddActionRow(
		discord.NewPrimaryButton("+1", "module-infinite_counter-all-counterup"),
	)
}

func CreatePanelUpdate(b *core.Bot, currentCount int, name, avatarURL string) discord.MessageUpdate {
	return discord.NewMessageUpdate().ClearEmbeds().AddEmbeds(
		discord.Embed{
			Color:       utils.ParseStrColor("#2c92b8"),
			Description: fmt.Sprintf("## %d", currentCount),
			Author: &discord.EmbedAuthor{
				Name:    name,
				IconURL: avatarURL,
			},
		},
	).AddActionRow(
		discord.NewPrimaryButton("+1", "module-infinite_counter-all-counterup"),
	)
}

func SendPanel(b *core.Bot, currentCount int, name, avatarURL string, cid snowflake.ID) error {
	msg := CreatePanel(b, currentCount, name, avatarURL)
	_, err := b.Client.Rest.CreateMessage(cid, msg)
	return err
}

func (m *InfiniteCounterModule) HandleAction(b *core.Bot, guildID string, action string, payload map[string]any) (any, error) {
	if action == "send_panel" {
		locale := discord.LocaleEnglishUS
		if payload != nil {
			if l, ok := payload["_lang"].(string); ok && l != "" {
				if strings.HasPrefix(l, "fr") {
					locale = discord.LocaleFrench
				} else if strings.HasPrefix(l, "en") {
					locale = discord.LocaleEnglishUS
				} else {
					locale = discord.Locale(l)
				}
			} else if l, ok := payload["lang"].(string); ok && l != "" {
				if strings.HasPrefix(l, "fr") {
					locale = discord.LocaleFrench
				} else if strings.HasPrefix(l, "en") {
					locale = discord.LocaleEnglishUS
				} else {
					locale = discord.Locale(l)
				}
			}
		} else if gid, err := snowflake.Parse(guildID); err == nil {
			if guild, ok := b.Client.Caches.Guild(gid); ok && guild.PreferredLocale != "" {
				locale = discord.Locale(guild.PreferredLocale)
			}
		}

		trad := locales.GetModule_InfiniteCounterModule(locale)

		cid, err := snowflake.Parse(m.Data.Main.ChannelID)
		if err != nil {
			return map[string]any{"success": false}, errors.New(trad.Action_invalid_channel)
		}

		self, ok := b.Client.Caches.SelfUser()
		if !ok {
			return map[string]any{"success": false}, errors.New(trad.Action_bot_not_found)
		}

		globalName := self.Username
		if self.GlobalName != nil && *self.GlobalName != "" {
			globalName = *self.GlobalName
		}

		avatarURL := ""
		if self.AvatarURL() != nil {
			avatarURL = *self.AvatarURL()
		}

		err = SendPanel(b, int(m.Data.ServerCount), globalName, avatarURL, cid)
		if err != nil {
			return map[string]any{"success": false}, errors.New(trad.Action_cannot_send_panel)
		}
		return map[string]any{"success": true}, nil
	}
	return nil, nil
}

func (m *InfiniteCounterModule) UISchema(locale discord.Locale) core.UISchema {
	meta := locales.GetMeta(locale, "module_InfiniteCounterModule")

	mainLabel := "Main Configuration"
	mainDesc := "Configure the counter channel and mode."
	mainChannelLabel := "Counter Channel"
	mainChannelDesc := "The text channel for the counter."
	mainTypeLabel := "Counter Mode"
	mainTypeDesc := "Select the counter mode for your server."
	freeTypeLabel := "Free Counter"
	strictTypeLabel := "Strict Counter"

	if sub, ok := meta.Submodules["main"]; ok {
		if sub.Label != "" {
			mainLabel = sub.Label
		}
		if sub.Description != "" {
			mainDesc = sub.Description
		}
		if opt, ok := sub.Options["channel_id"]; ok {
			if opt.Label != "" {
				mainChannelLabel = opt.Label
			}
			if opt.Description != "" {
				mainChannelDesc = opt.Description
			}
		}
		if opt, ok := sub.Options["type"]; ok {
			if opt.Label != "" {
				mainTypeLabel = opt.Label
			}
			if opt.Description != "" {
				mainTypeDesc = opt.Description
			}
			if o, ok := opt.Options["free"]; ok && o.Label != "" {
				freeTypeLabel = o.Label
			}
			if o, ok := opt.Options["strict"]; ok && o.Label != "" {
				strictTypeLabel = o.Label
			}
		}
	}

	freeLabel := "Free Counter Settings"
	freeDesc := "Users write the numbers themselves in the chat."
	freeRestartLabel := "Restart on Error"
	freeRestartDesc := "Reset the counter to 0 if a user makes a mistake."
	freeAllowChatLabel := "Allow Chat"
	freeAllowChatDesc := "Allow regular chat messages alongside numbers."

	if sub, ok := meta.Submodules["free"]; ok {
		if sub.Label != "" {
			freeLabel = sub.Label
		}
		if sub.Description != "" {
			freeDesc = sub.Description
		}
		if opt, ok := sub.Options["restart_on_error"]; ok {
			if opt.Label != "" {
				freeRestartLabel = opt.Label
			}
			if opt.Description != "" {
				freeRestartDesc = opt.Description
			}
		}
		if opt, ok := sub.Options["allow_chat"]; ok {
			if opt.Label != "" {
				freeAllowChatLabel = opt.Label
			}
			if opt.Description != "" {
				freeAllowChatDesc = opt.Description
			}
		}
	}

	strictLabel := "Strict Counter Settings"
	strictDesc := "The bot handles counting via an interactive panel button."
	strictSendPanelLabel := "Send Panel"
	strictSendPanelDesc := "Send or update the panel message with the button in the channel."

	if sub, ok := meta.Submodules["strict"]; ok {
		if sub.Label != "" {
			strictLabel = sub.Label
		}
		if sub.Description != "" {
			strictDesc = sub.Description
		}
		if opt, ok := sub.Options["send_panel"]; ok {
			if opt.Label != "" {
				strictSendPanelLabel = opt.Label
			}
			if opt.Description != "" {
				strictSendPanelDesc = opt.Description
			}
		}
	}

	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "main",
				Label:       mainLabel,
				Description: mainDesc,
				Components: []core.UIComponent{
					{
						Name:         "channel_id",
						Label:        mainChannelLabel,
						Description:  mainChannelDesc,
						Type:         core.ComponentTypeChannel,
						Required:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
					{
						Name:        "type",
						Label:       mainTypeLabel,
						Description: mainTypeDesc,
						Type:        core.ComponentTypeSelect,
						Required:    true,
						Options: []core.UISelectOption{
							{Label: freeTypeLabel, Value: "free"},
							{Label: strictTypeLabel, Value: "strict"},
						},
					},
				},
			},
			{
				Name:        "free",
				Label:       freeLabel,
				Description: freeDesc,
				Components: []core.UIComponent{
					{
						Name:        "restart_on_error",
						Label:       freeRestartLabel,
						Description: freeRestartDesc,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "allow_chat",
						Label:       freeAllowChatLabel,
						Description: freeAllowChatDesc,
						Type:        core.ComponentTypeBoolean,
					},
				},
			},
			{
				Name:        "strict",
				Label:       strictLabel,
				Description: strictDesc,
				Components: []core.UIComponent{
					{
						Name:        "send_panel",
						Label:       strictSendPanelLabel,
						Description: strictSendPanelDesc,
						Type:        core.ComponentTypeButton,
						Action:      "send_panel",
						Variant:     "primary",
					},
				},
			},
		},
	}
}
