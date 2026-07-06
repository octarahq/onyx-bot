package events

import (
	"strings"

	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterEvent(handlers.Event{
		Name:     "ApplicationCommandInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.ApplicationCommandInteractionCreate)
			if !ok {
				return
			}

			cmd, exists := handlers.Commands[event.Data.CommandName()]
			if exists && cmd.Execute != nil {
				cmd.Execute(b, event)
			}
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "ComponentInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.ComponentInteractionCreate)
			if !ok {
				return
			}

			customID := event.Data.CustomID()
			parts := strings.Split(customID, "-")
			if len(parts) < 2 {
				return
			}

			commandName := parts[0]
			userID := parts[1]

			if userID != "all" && userID != event.User().ID.String() {
				trad := locales.GetInteraction(event.Locale())
				event.CreateMessage(discord.MessageCreate{
					Content: trad.Not_allowed_component,
					Flags:   discord.MessageFlagEphemeral,
				})
				return
			}

			cmd, exists := handlers.Commands[commandName]
			if !exists {
				return
			}

			if cmd.ExecuteButton != nil {
				cmd.ExecuteButton(b, event)
			}
			if cmd.ExecuteMenu != nil {
				cmd.ExecuteMenu(b, event)
			}
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "ModalSubmitInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.ModalSubmitInteractionCreate)
			if !ok {
				return
			}

			customID := event.Data.CustomID
			parts := strings.Split(customID, "-")
			if len(parts) < 2 {
				return
			}

			commandName := parts[0]
			userID := parts[1]

			if userID != "all" && userID != event.User().ID.String() {
				trad := locales.GetInteraction(event.Locale())
				event.CreateMessage(discord.MessageCreate{
					Content: trad.Not_allowed_modal,
					Flags:   discord.MessageFlagEphemeral,
				})
				return
			}

			cmd, exists := handlers.Commands[commandName]
			if !exists {
				return
			}

			if cmd.ExecuteModal != nil {
				cmd.ExecuteModal(b, event)
			}
		},
	})

	handlers.RegisterEvent(handlers.Event{
		Name:     "AutocompleteInteractionCreate",
		ExecOnce: false,
		Execute: func(b *core.Bot, e bot.Event) {
			event, ok := e.(*events.AutocompleteInteractionCreate)
			if !ok {
				return
			}

			cmd, exists := handlers.Commands[event.Data.CommandName]
			if exists && cmd.ExecuteAutocomplete != nil {
				cmd.ExecuteAutocomplete(b, event)
			}
		},
	})
}
