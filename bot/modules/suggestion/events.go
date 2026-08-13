package suggestion

import (
	"fmt"
	"onyx/bot/core"

	"github.com/disgoorg/disgo/events"
)

func (m *SuggestionModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if m.Data.Enabled {
		fmt.Println("SuggestionModule message create handled!")
	}
	return false
}
