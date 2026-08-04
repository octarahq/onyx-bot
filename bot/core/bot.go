package core

import (
	"onyx/bot/db"
	"onyx/bot/logs"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
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

	ConnectedSince time.Time
	Version        string

	Logger logs.Logger
	ModuleLogger ModuleLogger
}

func (b *Bot) LogModuleInfo(gid string, moduleName string, title string, logs []string) {
	if b.ModuleLogger != nil {
		b.ModuleLogger.LogInfo(b, gid, moduleName, title, logs)
	}
}

func (b *Bot) LogModuleImportant(gid string, moduleName string, title string, logs []string) {
	if b.ModuleLogger != nil {
		b.ModuleLogger.LogImportant(b, gid, moduleName, title, logs)
	}
}

func (b *Bot) SendMessage(cid string, msg discord.MessageCreate) {
	scid, err := snowflake.Parse(cid)
	if err != nil {
		return
	}

	_, err = b.Client.Rest.CreateMessage(scid, msg)
	if err != nil {
		return
	}
}
