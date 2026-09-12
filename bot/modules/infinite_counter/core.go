package infinitecounter

import (
	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type MainCounterSettings struct {
	ChannelID string `json:"channel_id"`
	Type      string `gorm:"default:'free'" json:"type"`
}

type FreeCounterSettings struct {
	Enabled        bool `gorm:"default:false" json:"enabled"`
	RestartOnError bool `gorm:"default:false" json:"restart_on_error"`
	AllowChat      bool `gorm:"default:true" json:"allow_chat"`
}

type StrictCounterSettings struct {
	Enabled bool `gorm:"default:false" json:"enabled"`
}

type InfiniteCounterSettings struct {
	GuildID string                `gorm:"primaryKey" json:"guild_id"`
	Enabled bool                  `gorm:"default:false" json:"enabled"`
	Main    MainCounterSettings   `gorm:"embedded;embeddedPrefix:main_" json:"main"`
	Free    FreeCounterSettings   `gorm:"embedded;embeddedPrefix:free_" json:"free"`
	Strict  StrictCounterSettings `gorm:"embedded;embeddedPrefix:strict_" json:"strict"`
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
		Icon: "plus",
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
	}
}

func (m *InfiniteCounterModule) Schema() interface{}  { return &InfiniteCounterSettings{} }
func (m *InfiniteCounterModule) DataPtr() interface{} { return &m.Data }
func (m *InfiniteCounterModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = InfiniteCounterSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, InfiniteCounterSettings{GuildID: guildID}).Error
}

func (m *InfiniteCounterModule) HandleAction(b *core.Bot, guildID string, action string, payload map[string]any) (any, error) {
	if action == "send_panel" {
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
