package infinitecounter

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/locales"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func parseNewCounter(content string, allowChat bool, currentCount int) (int, bool) {
	nextCountStr := fmt.Sprintf("%d", currentCount+1)

	switch allowChat {
	case true:
		parts := strings.Split(content, " ")

		if slices.Contains(parts, nextCountStr) {
			return currentCount + 1, true
		}

		return 0, false
	case false:
		if content == nextCountStr {
			return currentCount + 1, true
		} else {
			return 0, false
		}
	}

	return 0, false
}

func (m *InfiniteCounterModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if !m.Data.Enabled || e.Message.Author.Bot {
		return false
	}

	if e.ChannelID.String() != m.Data.Main.ChannelID {
		return false
	}

	if m.Data.Main.Type == "free" {
		nextCount, valid := parseNewCounter(e.Message.Content, m.Data.Free.AllowChat, int(m.Data.ServerCount))
		userID := e.Message.Author.ID.String()

		if valid {
			m.Data.ServerCount = int64(nextCount)
			if m.Data.UserProgress == nil {
				m.Data.UserProgress = make(map[string]int64)
			}
			m.Data.UserProgress[userID]++

			_ = b.DB.GormDB.Save(&m.Data).Error
			return true
		} else {
			if m.Data.Free.RestartOnError {
				m.Data.ServerCount = 0
				_ = b.DB.GormDB.Save(&m.Data).Error

				locale := discord.LocaleFrench
				if e.GuildID != nil {
					if guild, ok := e.Client().Caches.Guild(*e.GuildID); ok && guild.PreferredLocale != "" {
						locale = discord.Locale(guild.PreferredLocale)
					}
				}
				trad := locales.GetModule_InfiniteCounterModule(locale)

				msgCreate := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay(trad.Error_title),
						discord.NewTextDisplay(trad.Error_desc),
						discord.NewTextDisplay(trad.Error_next),
					).WithAccentColor(0xED4245),
				)
				msgCreate.MessageReference = &discord.MessageReference{
					MessageID: &e.MessageID,
				}

				_, _ = e.Client().Rest.CreateMessage(e.ChannelID, msgCreate)
				return true
			} else {
				e.Client().Rest.DeleteMessage(e.Message.ChannelID, e.MessageID)
			}
		}
	}

	return false
}
