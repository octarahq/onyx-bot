package handlers

import (
	"reflect"

	"onyx/bot/core"

	"github.com/disgoorg/disgo/bot"
)

type Event = core.Event

var Events []Event

func RegisterEvent(e Event) {
	Events = append(Events, e)
}

func SetupEvents(b *core.Bot) {
	executedEvents := make(map[string]bool)

	b.Client.EventManager.AddEventListeners(bot.NewListenerFunc(func(e bot.Event) {
		eventName := reflect.TypeOf(e).Elem().Name()

		go ExecModulesEvent(b, e)

		for _, ev := range Events {
			if ev.Name == eventName {
				if ev.ExecOnce {
					if executedEvents[ev.Name] {
						continue
					}
					executedEvents[ev.Name] = true
				}

				go ev.Execute(b, e)
			}
		}
	}))
}
