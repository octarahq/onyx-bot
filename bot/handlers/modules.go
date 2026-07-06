package handlers

import (
	"onyx/bot/core"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
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

	var guildID *snowflake.ID

	switch e := event.(type) {
	case *events.MessageCreate:
		guildID = e.GuildID
	}

	if guildID == nil {
		return false
	}

	guildIDStr := guildID.String()

	me, ok := event.Client().Caches.Member(*guildID, event.Client().ID())
	if !ok {
		m, err := event.Client().Rest.GetMember(*guildID, event.Client().ID())
		if err != nil {
			return false
		}
		me = *m
	}

	for _, mod := range b.Modules {
		if !mod.IsEnabled() {
			continue
		}

		if dbAware, ok := mod.(core.DatabaseAware); ok {
			if err := dbAware.LoadData(b.DB.GormDB, guildIDStr); err != nil {
				continue
			}
		}

		if !CheckPerms(b, mod, me) {
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

func CheckPerms(b *core.Bot, module core.Module, me discord.Member) bool {
	var missing []string
	botPerms := b.Client.Caches.MemberPermissions(me)

	if botPerms.Has(discord.PermissionAdministrator) {
		return true
	}

	permissions := module.Permissions()
	for _, p := range permissions {
		if !botPerms.Has(p) {
			missing = append(missing, p.String())
		}
	}

	if len(missing) == 0 {
		return true
	}
	return false
}
