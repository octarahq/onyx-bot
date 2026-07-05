package core

import (
	"onyx/bot/db"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Command struct {
	Name                string
	Description         string
	Category            string
	Create              discord.ApplicationCommandCreate
	Execute             func(b *Bot, event *events.ApplicationCommandInteractionCreate)
	ExecuteButton       func(b *Bot, event *events.ComponentInteractionCreate)
	ExecuteMenu         func(b *Bot, event *events.ComponentInteractionCreate)
	ExecuteModal        func(b *Bot, event *events.ModalSubmitInteractionCreate)
	ExecuteAutocomplete func(b *Bot, event *events.AutocompleteInteractionCreate)
}

type Event struct {
	Name     string
	ExecOnce bool
	Execute  func(b *Bot, e bot.Event)
}

type Bot struct {
	Client   *bot.Client
	AdminIDs []string
	DB       *db.DB

	Commands map[string]Command
	Events   []Event

	Modules []Module
}
