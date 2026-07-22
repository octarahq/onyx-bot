package events

import (
	"context"
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
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
				fmt.Printf("Successfully synced %d commands!\n", len(b.Commands))
				fmt.Printf("Successfully %d modules!\n", len(b.Modules))
			}

			go func() {
				ticker := time.NewTicker(15 * time.Second)
				defer ticker.Stop()

				state := 0
				for {
					var status string
					switch state {
					case 0:
						status = "onyx.octara.xyz"
					case 1:
						serverCount := b.Client.Caches.GuildsLen()
						status = fmt.Sprintf("I'm in %d servers", serverCount)
					case 2:
						status = "Open-Source bot"
					}

					err := b.Client.SetPresence(context.Background(), gateway.WithCustomActivity(status))
					if err != nil {
						fmt.Printf("Failed to set presence: %v\n", err)
					}

					state = (state + 1) % 3
					<-ticker.C
				}
			}()
		},
	})
}
