package events

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "Ready",
		ExecOnce: true,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.Ready)
			if !ok {
				return
			}

			fmt.Println("Bot is ready! Logged in as", event.User.Username)

			if err := handlers.SyncCommands(b); err != nil {
				fmt.Printf("Failed to sync commands: %v\n", err)
			} else {
				fmt.Println("Successfully synced commands!")
			}
		},
	})
}
