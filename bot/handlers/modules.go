package handlers

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func ExecModulesEvent(b *core.Bot, event bot.Event) bool {
	if e, ok := event.(*events.Ready); ok {
		for _, mod := range b.Modules {
			if handler, ok := mod.(core.OnReady); ok {
				if handler.HandleReady(b, e) {
					return true
				}
			}
		}
		return false
	}

	var guildIDStr string

	switch e := event.(type) {
	case *events.MessageCreate:
		if e.GuildID != nil {
			guildIDStr = e.GuildID.String()
		}
	}

	if guildIDStr == "" {
		return false
	}

	for _, mod := range b.Modules {
		if dbAware, ok := mod.(core.DatabaseAware); ok {
			if err := dbAware.LoadData(b.DB.GormDB, guildIDStr); err != nil {
				continue
			}
		}

		if !mod.IsEnabled() {
			continue
		}

		switch e := event.(type) {
		case *events.MessageCreate:
			if handler, ok := mod.(core.OnMessageCreate); ok {
				if handler.HandleMessageCreate(b, e) {
					return true
				}
			}
		}
	}

	return false
}
