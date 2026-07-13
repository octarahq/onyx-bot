package modules

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
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

func (t *TranslationSettings) Validate() error {
	if t.Main.Lang != "" {
		valid := false
		for _, v := range utils.TranslateLangs {
			if t.Main.Lang == v.Value {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("invalid language: %s", t.Main.Lang)
		}
	}

	if t.Main.Channels != "" {
		channels := strings.Split(t.Main.Channels, ",")
		if len(channels) > 5 {
			return fmt.Errorf("maximum 5 channels allowed")
		}

		for _, ch := range channels {
			ch = strings.TrimSpace(ch)
			if len(ch) < 17 || len(ch) > 19 {
				return fmt.Errorf("invalid channel id: '%s' must be 17-19 characters", ch)
			}
			for _, r := range ch {
				if r < '0' || r > '9' {
					return fmt.Errorf("invalid channel id: '%s' must be numeric", ch)
				}
			}
		}
	}
	return nil
}

type TranslationModule struct {
	Data TranslationSettings
}

func init() {
	Register(&TranslationModule{})
}

func (m *TranslationModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "TranslationModule",
		Icon: "translate",
		Label: func(locale discord.Locale) string {
			return locales.GetModule_TranslationModule(locale).Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetModule_TranslationModule(locale).Description
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

func (m *TranslationModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if e.Message.Author.Bot {
		return false
	}

	channel, ok := e.Client().Caches.Channel(e.ChannelID)
	if !ok || channel.Type() != discord.ChannelTypeGuildNews {
		return false
	}

	if m.Data.Main.Channels == "" {
		return false
	}
	channelIDs := strings.Split(m.Data.Main.Channels, ",")

	var ch string
	for _, c := range channelIDs {
		c = strings.TrimSpace(c)
		if c == e.ChannelID.String() {
			ch = e.ChannelID.String()
		}
	}

	if ch != e.ChannelID.String() {
		return false
	}

	params := discord.ThreadCreateFromMessage{
		Name:                fmt.Sprintf("Traduction %s", m.Data.Main.Lang),
		AutoArchiveDuration: discord.AutoArchiveDuration1w,
	}

	thread, err := e.Client().Rest.CreateThreadFromMessage(e.ChannelID, e.MessageID, params)
	if err != nil {
		return false
	}

	content := e.Message.Content

	if len(content) > 2000 {
		content = fmt.Sprintf("%s...", content[0:1996])
	}

	t := utils.Translate(utils.TranslateParams{
		Query:  content,
		Source: "auto",
		Target: m.Data.Main.Lang,
	})

	trad := t.TranslatedText
	if len(trad) > 2000 {
		trad = fmt.Sprintf("%s...", trad[0:1996])
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplay(trad),
			discord.NewTextDisplayf("-# %s %s Translation", utils.TranslateLangs[m.Data.Main.Lang].Flag, utils.TranslateLangs[m.Data.Main.Lang].Name),
			discord.NewActionRow(
				discord.NewSecondaryButton("Translate", "translate-all-ephemeral"),
			),
		),
	)

	if _, err := e.Client().Rest.CreateMessage(thread.ID(), msg); err != nil {
		fmt.Printf("Error %s\n", err.Error())
		return false
	}

	return false
}

func (m *TranslationModule) Schema() interface{}  { return &TranslationSettings{} }
func (m *TranslationModule) DataPtr() interface{} { return &m.Data }
func (m *TranslationModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = TranslationSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, TranslationSettings{GuildID: guildID}).Error
}

func (m *TranslationModule) UISchema() core.UISchema {
	var langOptions []core.UISelectOption
	for _, v := range utils.TranslateLangs {
		langOptions = append(langOptions, core.UISelectOption{
			Label: v.Name,
			Value: v.Value,
		})
	}

	maxChannels := 5

	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:  "main",
				Label: "Main Settings",
				Components: []core.UIComponent{
					{
						Name:     "lang",
						Label:    "Language",
						Type:     core.ComponentTypeSelect,
						Required: true,
						Options:  langOptions,
					},
					{
						Name:        "channels",
						Label:       "Channels",
						Description: "Select up to 5 channels for automatic translation.",
						Type:        core.ComponentTypeChannel,
						Required:    false,
						Multiple:    true,
						Max:         &maxChannels,
					},
				},
			},
		},
	}
}
