package socialnotifs

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *SocialNotifsModule) Command() *discord.SlashCommandCreate {
	return nil
}

func (m *SocialNotifsModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	return false
}
