package core

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"gorm.io/gorm"
)

type Module interface {
	Name() string
	Priority() int
	IsEnabled() bool
	Permissions() []discord.Permissions
}

type DatabaseAware interface {
	Schema() interface{}
	LoadData(db *gorm.DB, guildID string) error
	DataPtr() interface{}
}

type OnMessageCreate interface {
	HandleMessageCreate(b *Bot, event *events.MessageCreate) bool
}

type OnReady interface {
	HandleReady(b *Bot, event *events.Ready) bool
}
