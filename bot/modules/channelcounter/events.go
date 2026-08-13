package channelcounter

import (
	"onyx/bot/core"
	"onyx/bot/utils"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func (m *ChannelCounterModule) HandleReady(b *core.Bot, e *events.Ready) bool {
	go func(bot *core.Bot) {
		for {
			var settings []ChannelCounterSettings
			bot.DB.GormDB.Find(&settings)

			if len(settings) == 0 {
				time.Sleep(5 * time.Minute)
				continue
			}

			delay := time.Hour / time.Duration(len(settings))
			if delay > 5*time.Minute {
				delay = 5 * time.Minute
			}

			for _, data := range settings {
				time.Sleep(delay)

				if !data.Enabled {
					continue
				}

				gid, err := snowflake.Parse(data.GuildID)
				if err != nil {
					continue
				}

				guild, exist := bot.Client.Caches.Guild(gid)
				if !exist {
					continue
				}

				if data.MemberCount.Enabled {
					scid, err := snowflake.Parse(data.MemberCount.Channel)
					if err == nil {
						channel, exist := bot.Client.Caches.Channel(scid)
						if exist {
							channelName := channel.Name()
							newChannelName := getChannelName(data.MemberCount.NameConv, guild.MemberCount)
							if channelName != newChannelName {
								bot.Client.Rest.UpdateChannel(scid, discord.GuildVoiceChannelUpdate{
									Name: &newChannelName,
								})
							}
						}
					}
				}

				var humansCount, botsCount int
				if data.HumansCount.Enabled || data.BotsCount.Enabled || data.StatusSettings.Enabled {
					for member := range bot.Client.Caches.Members(gid) {
						if member.User.Bot {
							botsCount++
						} else {
							humansCount++
						}
					}
				}

				if data.HumansCount.Enabled {
					scid, err := snowflake.Parse(data.HumansCount.Channel)
					if err == nil {
						channel, exist := bot.Client.Caches.Channel(scid)
						if exist {
							channelName := channel.Name()
							newChannelName := getChannelName(data.HumansCount.NameConv, humansCount)
							if channelName != newChannelName {
								bot.Client.Rest.UpdateChannel(scid, discord.GuildVoiceChannelUpdate{
									Name: &newChannelName,
								})
							}
						}
					}
				}

				if data.BotsCount.Enabled {
					scid, err := snowflake.Parse(data.BotsCount.Channel)
					if err == nil {
						channel, exist := bot.Client.Caches.Channel(scid)
						if exist {
							channelName := channel.Name()
							newChannelName := getChannelName(data.BotsCount.NameConv, botsCount)
							if channelName != newChannelName {
								bot.Client.Rest.UpdateChannel(scid, discord.GuildVoiceChannelUpdate{
									Name: &newChannelName,
								})
							}
						}
					}
				}

				var channelsCount int
				if data.ChannelsCount.Enabled || data.StatusSettings.Enabled {
					for gc := range bot.Client.Caches.ChannelsForGuild(gid) {
						_ = gc
						channelsCount++
					}
				}

				if data.ChannelsCount.Enabled {
					scid, err := snowflake.Parse(data.ChannelsCount.Channel)
					if err == nil {
						channel, exist := bot.Client.Caches.Channel(scid)
						if exist {
							channelName := channel.Name()
							newChannelName := getChannelName(data.ChannelsCount.NameConv, channelsCount)
							if channelName != newChannelName {
								bot.Client.Rest.UpdateChannel(scid, discord.GuildVoiceChannelUpdate{
									Name: &newChannelName,
								})
							}
						}
					}
				}

				var rolesCount int
				if data.RolesCount.Enabled || data.StatusSettings.Enabled {
					rolesCount = bot.Client.Caches.RolesLen(gid)
				}

				if data.RolesCount.Enabled {
					scid, err := snowflake.Parse(data.RolesCount.Channel)
					if err == nil {
						channel, exist := bot.Client.Caches.Channel(scid)
						if exist {
							channelName := channel.Name()
							newChannelName := getChannelName(data.RolesCount.NameConv, rolesCount)
							if channelName != newChannelName {
								bot.Client.Rest.UpdateChannel(scid, discord.GuildVoiceChannelUpdate{
									Name: &newChannelName,
								})
							}
						}
					}
				}

				if data.StatusSettings.Enabled && len(data.StatusSettings.Names) > 0 {
					scid, err := snowflake.Parse(data.StatusSettings.Channel)
					if err == nil {
						channel, exist := bot.Client.Caches.Channel(scid)
						if exist {
							currentIndex := data.StatusSettings.CurrentIndex
							if currentIndex >= len(data.StatusSettings.Names) {
								currentIndex = 0
							}

							nameTemplate := data.StatusSettings.Names[currentIndex]

							newChannelName := utils.ParseVariables(nameTemplate, map[string]string{
								"member_count":   utils.ParseCount(guild.MemberCount),
								"humans_count":   utils.ParseCount(humansCount),
								"bots_count":     utils.ParseCount(botsCount),
								"channels_count": utils.ParseCount(channelsCount),
								"roles_count":    utils.ParseCount(rolesCount),
							})

							channelName := channel.Name()
							if channelName != newChannelName {
								bot.Client.Rest.UpdateChannel(scid, discord.GuildVoiceChannelUpdate{
									Name: &newChannelName,
								})
							}

							bot.DB.GormDB.Model(&data).Update("statuscounter_current_index", currentIndex+1)
						}
					}
				}
			}
		}
	}(b)

	return false
}
