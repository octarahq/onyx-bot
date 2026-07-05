package modules

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/db"

	"github.com/disgoorg/disgo/events"
)

type TestModule struct {
	Data *db.Guild
}

func init() {
	Register(&TestModule{})
}

func (m *TestModule) Name() string    { return "TestModule" }
func (m *TestModule) Priority() int   { return 1 }
func (m *TestModule) IsEnabled() bool { return m.Data.SettingsTest.Enabled }
func (m *TestModule) Config() core.ModuleConfig {
	return core.ModuleConfig{
		IsFetchable: true,
		IsEditable:  true,
	}
}

func (m *TestModule) SetData(data db.Guild) {
	m.Data = &data
}

func (m *TestModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) {
	fmt.Println("Test !")
}
