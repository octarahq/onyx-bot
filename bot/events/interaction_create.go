package events

import (
	"onyx/bot/core"
	"onyx/bot/handlers"

	"github.com/disgoorg/disgo/bot"
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

			// In a real bot, we would parse the custom ID to find the command or component handler
			// For this boilerplate, we'll iterate over commands to see if they have ExecuteButton or ExecuteMenu (very simplified)
			// A robust implementation would store component handlers separately or route by a prefix.
			for _, cmd := range handlers.Commands {
				if cmd.ExecuteButton != nil {
					cmd.ExecuteButton(b, event)
				}
				if cmd.ExecuteMenu != nil {
					cmd.ExecuteMenu(b, event)
				}
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

			for _, cmd := range handlers.Commands {
				if cmd.ExecuteModal != nil {
					cmd.ExecuteModal(b, event)
				}
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
