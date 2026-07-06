package modules

import (
	"fmt"
	"strings"
	"onyx/bot/core"

	"github.com/disgoorg/disgo/events"
	"gorm.io/gorm"
)

type TranslationSettings struct {
	GuildID  string `gorm:"primaryKey" json:"guild_id"`
	Enabled  bool   `gorm:"default:false" json:"enabled"`
	Channels string `json:"channels"`
	Lang     string `gorm:"default:'en-US'" json:"lang"`
}

func (t *TranslationSettings) Validate() error {
	if len(t.Lang) > 10 {
		return fmt.Errorf("lang exceeds maximum length of 10 characters")
	}

	if t.Channels != "" {
		channels := strings.Split(t.Channels, ",")
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

func (m *TranslationModule) Name() string    { return "TranslationModule" }
func (m *TranslationModule) Priority() int   { return 1 }
func (m *TranslationModule) IsEnabled() bool { return m.Data.Enabled }

func (m *TranslationModule) Schema() interface{} { return &TranslationSettings{} }
func (m *TranslationModule) DataPtr() interface{} { return &m.Data }
func (m *TranslationModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = TranslationSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, TranslationSettings{GuildID: guildID}).Error
}

func (m *TranslationModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	fmt.Println("Test !")
	return false
}
