package infinitecounter

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

type userProgressEntry struct {
	userID string
	count  int64
}

func (m *InfiniteCounterModule) Command() *discord.SlashCommandCreate {
	return &discord.SlashCommandCreate{
		Name:        "counter",
		Description: "Commandes du module Compteur Infini",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "leaderboard",
				Description: "Affiche le classement des membres du serveur",
			},
		},
	}
}

func (m *InfiniteCounterModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	if !m.IsEnabled() {
		return false
	}

	if event.GuildID() == nil {
		return false
	}

	subCmd := event.SlashCommandInteractionData().SubCommandName
	if subCmd != nil && *subCmd != "leaderboard" && *subCmd != "classement" {
		return false
	}

	_ = m.LoadData(b.DB.GormDB, event.GuildID().String())

	trad := locales.GetModule_InfiniteCounterModule(event.Locale())

	var entries []userProgressEntry
	for uid, cnt := range m.Data.UserProgress {
		if cnt > 0 {
			entries = append(entries, userProgressEntry{userID: uid, count: cnt})
		}
	}

	slices.SortFunc(entries, func(a, b userProgressEntry) int {
		if b.count > a.count {
			return 1
		} else if b.count < a.count {
			return -1
		}
		return 0
	})

	var containerComps []discord.ContainerSubComponent
	containerComps = append(containerComps, discord.NewTextDisplay(trad.Leaderboard_title))

	if len(entries) == 0 {
		containerComps = append(containerComps, discord.NewTextDisplay(trad.Leaderboard_empty))
	} else {
		top1 := entries[0]
		top1ID, _ := snowflake.Parse(top1.userID)

		name := fmt.Sprintf("<@%s>", top1.userID)
		avatarURL := ""

		if member, exist := event.Client().Caches.Member(*event.GuildID(), top1ID); exist {
			name = member.EffectiveName()
			if member.User.AvatarURL() != nil {
				avatarURL = *member.User.AvatarURL()
			}
		} else if user, err := event.Client().Rest.GetUser(top1ID); err == nil {
			name = user.EffectiveName()
			if user.AvatarURL() != nil {
				avatarURL = *user.AvatarURL()
			}
		}

		topText := fmt.Sprintf(trad.Leaderboard_first, name, top1.count)
		section := discord.NewSection(discord.NewTextDisplay(topText))
		if avatarURL != "" {
			section = section.WithAccessory(discord.NewThumbnail(avatarURL))
		}
		containerComps = append(containerComps, section)

		if len(entries) > 1 {
			limit := len(entries)
			if limit > 10 {
				limit = 10
			}

			var otherLines []string
			for i := 1; i < limit; i++ {
				e := entries[i]
				eID, _ := snowflake.Parse(e.userID)
				uName := fmt.Sprintf("<@%s>", e.userID)
				if member, exist := event.Client().Caches.Member(*event.GuildID(), eID); exist {
					uName = member.EffectiveName()
				}
				otherLines = append(otherLines, fmt.Sprintf("**#%d** %s — **%d**", i+1, uName, e.count))
			}
			containerComps = append(containerComps, discord.NewTextDisplay(strings.Join(otherLines, "\n")))
		}
	}

	msg := discord.NewMessageCreateV2(discord.NewContainer(containerComps...))
	_ = event.CreateMessage(msg)
	return true
}
