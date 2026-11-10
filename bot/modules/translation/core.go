package translation

import (
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

type TranslationMainSettings struct {
	Channels string `json:"channels"`
	Lang     string `gorm:"default:'en'" json:"lang"`
}

type TranslationSettings struct {
	GuildID string                  `gorm:"primaryKey" json:"guild_id"`
	Enabled bool                    `gorm:"default:false" json:"enabled"`
	Main    TranslationMainSettings `gorm:"embedded;embeddedPrefix:main_" json:"main"`
}

type TranslationModule struct {
	Data TranslationSettings
}

func init() {
	core.Register(&TranslationModule{})
}

func (m *TranslationModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "TranslationModule",
		Icon: "translate",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_TranslationModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_TranslationModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_TranslationModule")
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

func (m *TranslationModule) Priority() int   { return 1 }
func (m *TranslationModule) IsEnabled() bool { return m.Data.Enabled }
func (m *TranslationModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionCreatePublicThreads,
		discord.PermissionSendMessages,
	}
}

func (m *TranslationModule) Schema() interface{}  { return &TranslationSettings{} }
func (m *TranslationModule) DataPtr() interface{} { return &m.Data }
func (m *TranslationModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = TranslationSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, TranslationSettings{GuildID: guildID}).Error
}

func (m *TranslationModule) UISchema(locale discord.Locale) core.UISchema {
	var langOptions []core.UISelectOption
	for _, v := range utils.TranslateLangs {
		langOptions = append(langOptions, core.UISelectOption{
			Label: v.Name,
			Value: v.Value,
		})
	}

	meta := locales.GetMeta(locale, "module_TranslationModule")
	mainLabel := "Main Settings"
	mainDesc := ""
	if sub, ok := meta.Submodules["main"]; ok {
		if sub.Label != "" {
			mainLabel = sub.Label
		}
		if sub.Description != "" {
			mainDesc = sub.Description
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
						Name:     "lang",
						Label:    "Language",
						Type:     core.ComponentTypeSelect,
						Required: true,
						Options:  langOptions,
					},
					{
						Name:         "channels",
						Label:        "Channels",
						Description:  "Select up to 5 channels for automatic translation.",
						Type:         core.ComponentTypeChannel,
						Required:     false,
						Multiple:     true,
						Max:          5,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildNews},
					},
				},
			},
		},
	}
}
