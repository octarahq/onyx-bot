package core

import (
	"onyx/bot/db"

	"github.com/disgoorg/disgo/events"
)

type ModuleConfig struct {
	IsFetchable bool
	IsEditable  bool
}

type Module interface {
	Name() string
	Priority() int
	IsEnabled() bool
	Config() ModuleConfig
}

type DataAware interface {
	SetData(data db.Guild)
}

type OnMessageCreate interface {
	HandleMessageCreate(b *Bot, event *events.MessageCreate)
}

type OnReady interface {
	HandleReady(b *Bot, event *events.Ready)
}
