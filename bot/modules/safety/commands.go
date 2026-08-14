package safety

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *SafetyModule) Command() *discord.SlashCommandCreate {
	return nil
}

func (m *SafetyModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
