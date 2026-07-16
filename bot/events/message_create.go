package events

import (
	"strings"

	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "MessageCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.MessageCreate)
			if !ok {
				return
			}

			if event.Message.Author.Bot {
				return
			}

			prefix := ","
			if strings.HasPrefix(event.Message.Content, prefix) {
				isAdmin := false
				for _, id := range b.AdminIDs {
					if event.Message.Author.ID.String() == id {
						isAdmin = true
						break
					}
				}

				if isAdmin {
					content := strings.TrimPrefix(event.Message.Content, prefix)
					args := strings.Fields(content)
					if len(args) == 0 {
						return
					}

					command := strings.ToLower(args[0])
					_ = command

					switch command {
					case "emit":
						if len(args) < 2 {
							return
						}
						eventToEmit := strings.ToLower(args[1])

						if event.GuildID == nil || event.Message.Member == nil {
							return
						}

						var mockEvent bot.Event
						genericEvent := events.NewGenericEvent(b.Client, 0, 0)

						switch eventToEmit {
						case "guildmemberjoin":
							mockEvent = &events.GuildMemberJoin{
								GenericGuildMember: &events.GenericGuildMember{
									GenericEvent: genericEvent,
									GuildID:      *event.GuildID,
									Member:       *event.Message.Member,
								},
							}
						case "guildmemberleave":
							mockEvent = &events.GuildMemberLeave{
								GenericEvent: genericEvent,
								GuildID:      *event.GuildID,
								User:         event.Message.Author,
								Member:       *event.Message.Member,
							}
						}

						if mockEvent != nil {
							b.Client.EventManager.DispatchEvent(mockEvent)
						}
					}
				}
			}
		},
	})
}
