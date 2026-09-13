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

		if valid && m.Data.LastUser == userID {
			valid = false
		}

		if valid {
			m.Data.ServerCount = int64(nextCount)
			m.Data.LastUser = userID
			if m.Data.UserProgress == nil {
				m.Data.UserProgress = make(map[string]int64)
			}
			m.Data.UserProgress[userID]++

			_ = b.DB.GormDB.Save(&m.Data).Error
			_ = e.Client().Rest.AddReaction(e.ChannelID, e.MessageID, "✅")
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
	} else if m.Data.Main.Type == "strict" {
		if e.Message.Member != nil && !e.Client().Caches.MemberPermissions(*e.Message.Member).Has(discord.PermissionAdministrator) {
			e.Client().Rest.DeleteMessage(e.Message.ChannelID, e.MessageID)
		}

		channel, exist := e.Channel()
		guild, exist := e.Guild()
		if !exist {
			return true
		}

		perms, exist := channel.PermissionOverwrites().Role(guild.ID)
		if !exist || !perms.Deny.Has(discord.PermissionSendMessages) {
			var allow, deny discord.Permissions
			if exist {
				allow = perms.Allow.Remove(discord.PermissionSendMessages)
				deny = perms.Deny.Add(discord.PermissionSendMessages)
			} else {
				allow = discord.Permissions(0)
				deny = discord.PermissionSendMessages
			}

			_ = e.Client().Rest.UpdatePermissionOverwrite(e.Message.ChannelID, guild.ID, discord.RolePermissionOverwriteUpdate{
				Allow: &allow,
				Deny:  &deny,
			})
		}

		return true
	}

	return false
}

func (m *InfiniteCounterModule) HandleComponentInteractionCreate(b *core.Bot, e *events.ComponentInteractionCreate) bool {
	if e.Data.CustomID() == "module-infinite_counter-all-counterup" && m.Data.Enabled {
		if m.Data.Main.Type == "strict" {
			if e.GuildID() != nil {
				_ = m.LoadData(b.DB.GormDB, e.GuildID().String())
			}

			uid := e.User().ID.String()
			if m.Data.LastUser == uid {
				locale := discord.LocaleFrench
				if e.GuildID() != nil {
					if guild, ok := e.Client().Caches.Guild(*e.GuildID()); ok && guild.PreferredLocale != "" {
						locale = discord.Locale(guild.PreferredLocale)
					}
				}
				trad := locales.GetModule_InfiniteCounterModule(locale)
				e.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_double_click).WithFlags(discord.MessageFlagEphemeral))
				return false
			}

			m.Data.ServerCount = m.Data.ServerCount + 1
			m.Data.LastUser = uid
			if m.Data.UserProgress == nil {
				m.Data.UserProgress = make(map[string]int64)
			}
			m.Data.UserProgress[uid]++

			_ = b.DB.GormDB.Save(&m.Data).Error

			msg := CreatePanel(b, int(m.Data.ServerCount), e.User().EffectiveName(), e.User().EffectiveAvatarURL())
			e.CreateMessage(msg)
			return true
		}
	}

	return false
}
