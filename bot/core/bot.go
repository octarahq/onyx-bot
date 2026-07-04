package core

import (
	"github.com/disgoorg/disgo/bot"
)

type Bot struct {
	Client   *bot.Client
	AdminIDs []string
}
