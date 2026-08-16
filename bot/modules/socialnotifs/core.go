package socialnotifs

import (
	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"gorm.io/gorm"
)

type GeneralNotifSettings[f any] struct {
	Enabled bool `gorm:"default:false" json:"enabled"`
	Fluxs   []f  `gorm:"serializer:json" json:"fluxs"`
}

type GithubFlux struct {
	RepoName string `json:"reponame"`
	Channel  string `json:"channel"`
	Message  string `json:"message"`
}

type RSSFlux struct {
	FeedURL    string `json:"feed_url"`
	Channel    string `json:"channel"`
	Message    string `json:"message"`
	ButtonText string `json:"button_text"`
}

type SocialNotifsSettings struct {
	GuildID string                           `gorm:"primaryKey" json:"guild_id"`
	Enabled bool                             `gorm:"default:false" json:"enabled"`
	Github  GeneralNotifSettings[GithubFlux] `gorm:"embedded;embeddedPrefix:github_" json:"github"`
	RSS     GeneralNotifSettings[RSSFlux]    `gorm:"embedded;embeddedPrefix:rss_" json:"rss"`
}

type SocialNotifsModule struct {
	Data SocialNotifsSettings
}

func init() {
	core.Register(&SocialNotifsModule{})
}

func (m *SocialNotifsModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "SocialNotifsModule",
		Icon: "bigtop_updates",
		Label: func(locale discord.Locale) string {
			return "Social Notifications"
		},
		Description: func(locale discord.Locale) string {
			return "Manage social media notifications"
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			subs := make(map[string]core.SubmoduleMeta)
			meta := locales.GetMeta(locale, "module_SocialNotifsModule")

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

func (m *SocialNotifsModule) Priority() int   { return 1 }
func (m *SocialNotifsModule) IsEnabled() bool { return m.Data.Enabled }
func (m *SocialNotifsModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionCreatePublicThreads,
		discord.PermissionSendMessages,
	}
}

func (m *SocialNotifsModule) Schema() interface{}  { return &SocialNotifsSettings{} }
func (m *SocialNotifsModule) DataPtr() interface{} { return &m.Data }
func (m *SocialNotifsModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = SocialNotifsSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, SocialNotifsSettings{GuildID: guildID}).Error
}

func (m *SocialNotifsModule) UISchema(locale discord.Locale) core.UISchema {
	schema := core.UISchema{
		SubModules: []core.UISubModule{},
	}
	meta := locales.GetMeta(locale, "module_SocialNotifsModule")

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
				{
					Name:          "fluxs",
					Label:         "Notification Fluxes",
					Description:   "Add sources to listen to",
					Type:          core.ComponentTypeObjectList,
					Required:      false,
					SubComponents: provider.UISchema(locale),
				},
			},
		}

		schema.SubModules = append(schema.SubModules, subModule)
	}

	return schema
}

func (m *SocialNotifsModule) HandleReady(b *core.Bot, event *events.Ready) bool {
	for _, provider := range Providers {
		provider.Init(b, b.DB.GormDB)
	}
	return false
}
