package freegames

import (
	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"gorm.io/gorm"
)

type FreeGameConfig struct {
	Enabled bool   `gorm:"default:false" json:"enabled"`
	Channel string `json:"channel"`
	Role    string `json:"role"`
}

type FreeGamesSettings struct {
	GuildID string         `gorm:"primaryKey" json:"guild_id"`
	Enabled bool           `gorm:"default:false" json:"enabled"`
	Steam   FreeGameConfig `gorm:"embedded;embeddedPrefix:steam_" json:"steam"`
	Epic    FreeGameConfig `gorm:"embedded;embeddedPrefix:epic_" json:"epic"`
}

type FreeGamesModule struct {
	Data FreeGamesSettings
}

func init() {
	core.Register(&FreeGamesModule{})
}

func (m *FreeGamesModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "FreeGamesModule",
		Icon: "videogame_asset",
		Label: func(locale discord.Locale) string {
			return "Free Games Notifications"
		},
		Description: func(locale discord.Locale) string {
			return "Manage free games notifications (Steam, Epic, etc.)"
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			subs := make(map[string]core.SubmoduleMeta)
			meta := locales.GetMeta(locale, "module_FreeGamesModule")

			for _, provider := range Providers {
				name := provider.Name()

				label := name
				desc := ""
				if subMeta, ok := meta.Submodules[name]; ok {
					label = subMeta.Label
					desc = subMeta.Description
				}

				subs[name] = core.SubmoduleMeta{
					Label:       label,
					Description: desc,
				}
			}
			return subs
		},
	}
}

func (m *FreeGamesModule) Priority() int   { return 1 }
func (m *FreeGamesModule) IsEnabled() bool { return m.Data.Enabled }
func (m *FreeGamesModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionSendMessages,
	}
}

func (m *FreeGamesModule) Schema() interface{}  { return &FreeGamesSettings{} }
func (m *FreeGamesModule) DataPtr() interface{} { return &m.Data }
func (m *FreeGamesModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = FreeGamesSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, FreeGamesSettings{GuildID: guildID}).Error
}

func (m *FreeGamesModule) UISchema(locale discord.Locale) core.UISchema {
	schema := core.UISchema{
		SubModules: []core.UISubModule{},
	}
	meta := locales.GetMeta(locale, "module_FreeGamesModule")

	for _, provider := range Providers {
		name := provider.Name()

		label := name
		desc := ""
		if subMeta, ok := meta.Submodules[name]; ok {
			label = subMeta.Label
			desc = subMeta.Description
		}

		subModule := core.UISubModule{
			Name:        name,
			Label:       label,
			Description: desc,
			Components: []core.UIComponent{
				{
					Name:     "enabled",
					Label:    "Enable " + label,
					Type:     core.ComponentTypeBoolean,
					Required: false,
				},
			},
		}

		subModule.Components = append(subModule.Components, provider.UISchema(locale)...)

		schema.SubModules = append(schema.SubModules, subModule)
	}

	return schema
}

func (m *FreeGamesModule) HandleReady(b *core.Bot, event *events.Ready) bool {
	for _, provider := range Providers {
		provider.Init(b, b.DB.GormDB)
	}
	return false
}
