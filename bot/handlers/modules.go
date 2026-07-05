package handlers

import (
	"onyx/bot/core"
	"onyx/bot/db"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func ExecModulesEvent(b *core.Bot, event bot.Event) {
	if e, ok := event.(*events.Ready); ok {
		for _, mod := range b.Modules {
			if handler, ok := mod.(core.OnReady); ok {
				go handler.HandleReady(b, e)
			}
		}
		return
	}

	var guildIDStr string

	switch e := event.(type) {
	case *events.MessageCreate:
		if e.GuildID != nil {
			guildIDStr = e.GuildID.String()
		}
	}

	if guildIDStr == "" {
		return
	}

	data, err := db.LoadSettings(b.DB.GormDB, guildIDStr)
	if err != nil {
		return
	}

	for _, mod := range b.Modules {
		if dataAware, ok := mod.(core.DataAware); ok {
			dataAware.SetData(*data)
		}

		if !mod.IsEnabled() {
			continue
		}

		switch e := event.(type) {
		case *events.MessageCreate:
			if handler, ok := mod.(core.OnMessageCreate); ok {
				go handler.HandleMessageCreate(b, e)
			}
		}
	}
}
