package infinitecounter

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *InfiniteCounterModule) Command() *discord.SlashCommandCreate {
	return nil
}

func (m *InfiniteCounterModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
