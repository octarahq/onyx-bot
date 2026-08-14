package translation

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *TranslationModule) Command() *discord.SlashCommandCreate {
	return nil
}

func (m *TranslationModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
