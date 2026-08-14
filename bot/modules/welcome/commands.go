package welcome

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *WelcomeModule) Command() *discord.SlashCommandCreate {
	return nil
}

func (m *WelcomeModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
