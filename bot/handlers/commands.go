package handlers

import (
	"onyx/bot/core"

	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Command struct {
	Name                string
	Description         string
	Category            string
	Create              discord.ApplicationCommandCreate
	Execute             func(b *core.Bot, event *events.ApplicationCommandInteractionCreate)
	ExecuteButton       func(b *core.Bot, event *events.ComponentInteractionCreate)
	ExecuteMenu         func(b *core.Bot, event *events.ComponentInteractionCreate)
	ExecuteModal        func(b *core.Bot, event *events.ModalSubmitInteractionCreate)
	ExecuteAutocomplete func(b *core.Bot, event *events.AutocompleteInteractionCreate)
}

var Commands = make(map[string]Command)

func RegisterCommand(cmd Command) {
	Commands[cmd.Name] = cmd
}

func SyncCommands(b *core.Bot) error {
	var cmds []discord.ApplicationCommandCreate
	for name, cmd := range Commands {
		create := cmd.Create

		if slashCmd, ok := create.(discord.SlashCommandCreate); ok {
			if slashCmd.NameLocalizations == nil {
				slashCmd.NameLocalizations = make(map[discord.Locale]string)
			}
			if slashCmd.DescriptionLocalizations == nil {
				slashCmd.DescriptionLocalizations = make(map[discord.Locale]string)
			}

			for locale, metas := range locales.Metas {
				if meta, exists := metas[name]; exists {
					if meta.Name != "" {
						slashCmd.NameLocalizations[locale] = meta.Name
					}
					if meta.Description != "" {
						slashCmd.DescriptionLocalizations[locale] = meta.Description
					}
					if len(meta.Options) > 0 && len(slashCmd.Options) > 0 {
						slashCmd.Options = localizeOptions(slashCmd.Options, locale, meta)
					}
				}
			}
			cmds = append(cmds, slashCmd)
		} else {
			cmds = append(cmds, create)
		}
	}

	_, err := b.Client.Rest.SetGlobalCommands(b.Client.ApplicationID, cmds)
	return err
}
