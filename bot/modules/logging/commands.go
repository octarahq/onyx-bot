package logging

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *LoggingModule) Command() *discord.SlashCommandCreate {
	return nil
}

func (m *LoggingModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
