package suggestion

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *SuggestionModule) Command() *discord.SlashCommandCreate {
	return &discord.SlashCommandCreate{
		Name:        "suggestion",
		Description: "Module de suggestions",
	}
}

func (m *SuggestionModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
