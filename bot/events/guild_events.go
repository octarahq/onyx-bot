package events

import (
	auth "onyx/bot/api/routes/auth/discord"
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "GuildJoin",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			_, ok := e.(*events.GuildJoin)
			if !ok {
				return
			}
			auth.UserGuildsCache.Clear()
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "GuildLeave",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			_, ok := e.(*events.GuildLeave)
			if !ok {
				return
			}
			auth.UserGuildsCache.Clear()
		},
	})
}
