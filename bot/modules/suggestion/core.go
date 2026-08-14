package suggestion

import (
	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type SuggestionMainSettings struct {
	Channel      string `json:"channel"`
	AllowDebate  bool   `json:"allow_debate"`
}

type SuggestionContentSettings struct {
	AllowImages bool `json:"allow_images"`
}


type SuggestionSettings struct {
	GuildID string                    `gorm:"primaryKey" json:"guild_id"`
	Enabled bool                      `gorm:"default:false" json:"enabled"`
	Main    SuggestionMainSettings    `gorm:"embedded;embeddedPrefix:main_" json:"main"`
	Content SuggestionContentSettings `gorm:"embedded;embeddedPrefix:content_" json:"content"`
}

type SuggestionModule struct {
	Data SuggestionSettings
}

func init() {
	core.Register(&SuggestionModule{})
}

func (m *SuggestionModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "SuggestionModule",
		Icon: "lightbulb_2",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_SuggestionModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_SuggestionModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_SuggestionModule")
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

func (m *SuggestionModule) Priority() int   { return 1 }
func (m *SuggestionModule) IsEnabled() bool { return m.Data.Enabled }
func (m *SuggestionModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionAddReactions,
		discord.PermissionAttachFiles,
		discord.PermissionCreatePublicThreads,
		discord.PermissionManageChannels,
		discord.PermissionManageThreads,
	}
}

func (m *SuggestionModule) Schema() interface{} {
	return []interface{}{&SuggestionSettings{}}
}
func (m *SuggestionModule) DataPtr() interface{} { return &m.Data }
func (m *SuggestionModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = SuggestionSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, SuggestionSettings{GuildID: guildID}).Error
}

func (m *SuggestionModule) UISchema(locale discord.Locale) core.UISchema {
	meta := locales.GetMeta(locale, "module_SuggestionModule")

	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "main",
				Label:       meta.Submodules["main"].Label,
				Description: meta.Submodules["main"].Description,
				Components: []core.UIComponent{
					{
						Name:         "channel",
						Label:        meta.Submodules["main"].Options["channel"].Label,
						Description:  meta.Submodules["main"].Options["channel"].Description,
						Type:         core.ComponentTypeChannel,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText, discord.ChannelTypeGuildForum},
						Required:     true,
					},
					{
						Name:        "allow_debate",
						Label:       meta.Submodules["main"].Options["allow_debate"].Label,
						Description: meta.Submodules["main"].Options["allow_debate"].Description,
						Type:        core.ComponentTypeBoolean,
						Required:    false,
					},

				},
			},
			{
				Name:        "content",
				Label:       meta.Submodules["content"].Label,
				Description: meta.Submodules["content"].Description,
				Components: []core.UIComponent{
					{
						Name:        "allow_images",
						Label:       meta.Submodules["content"].Options["allow_images"].Label,
						Description: meta.Submodules["content"].Options["allow_images"].Description,
						Type:        core.ComponentTypeBoolean,
						Required:    false,
					},
				},
			},
		},
	}
}
