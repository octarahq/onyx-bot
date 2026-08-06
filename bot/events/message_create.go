package events

import (
	"fmt"
	"strings"

	"onyx/bot/commands/informations"
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
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

			selfUser, exist := b.Client.Caches.SelfUser()
			if exist {
				botMention1 := fmt.Sprintf("<@%s>", selfUser.ID)
				botMention2 := fmt.Sprintf("<@!%s>", selfUser.ID)
				if strings.TrimSpace(event.Message.Content) == botMention1 || strings.TrimSpace(event.Message.Content) == botMention2 {
					locale := discord.LocaleFrench
					if event.GuildID != nil {
						if guild, ok := b.Client.Caches.Guild(*event.GuildID); ok {
							locale = discord.Locale(guild.PreferredLocale)
						}
					}
					msg, err := informations.BuildBotInfoMessage(b, b.Client, locale)
					if err == nil {
						msg = msg.WithMessageReferenceByID(event.Message.ID)
						_, _ = event.Client().Rest.CreateMessage(event.ChannelID, msg)
					}
					return
				}
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
						case "guildupdatevanityurl":
							oldCode := "onyx"
							newCode := "stolen"

							oldGuild := discord.Guild{
								ID: *event.GuildID,
							}
							newGuild := discord.Guild{
								ID: *event.GuildID,
							}

							if guild, ok := b.Client.Caches.Guild(*event.GuildID); ok {
								oldGuild = guild
								newGuild = guild
							}
							oldGuild.VanityURLCode = &oldCode
							newGuild.VanityURLCode = &newCode

							mockEvent = &events.GuildUpdate{
								GenericGuild: &events.GenericGuild{
									GenericEvent: genericEvent,
									GuildID:      *event.GuildID,
								},
								Guild:    newGuild,
								OldGuild: oldGuild,
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
