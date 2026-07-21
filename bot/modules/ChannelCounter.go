package modules

import (
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type ChannelCountMembersSettings struct {
	Enabled  bool   `json:"enabled"`
	Channel  string `json:"channel"`
	NameConv string `json:"nameconv" gorm:"default:'Members : {count}'"`
}

type ChannelCountHumansSettings struct {
	Enabled  bool   `json:"enabled"`
	Channel  string `json:"channel"`
	NameConv string `json:"nameconv" gorm:"default:'Humans : {count}'"`
}

type ChannelCountBotsSettings struct {
	Enabled  bool   `json:"enabled"`
	Channel  string `json:"channel"`
	NameConv string `json:"nameconv" gorm:"default:'Bots : {count}'"`
}

type ChannelCountChannelsSettings struct {
	Enabled  bool   `json:"enabled"`
	Channel  string `json:"channel"`
	NameConv string `json:"nameconv" gorm:"default:'Channels : {count}'"`
}

type ChannelCountRolesSettings struct {
	Enabled  bool   `json:"enabled"`
	Channel  string `json:"channel"`
	NameConv string `json:"nameconv" gorm:"default:'Roles : {count}'"`
}

type ChannelCountStatusSettings struct {
	Enabled      bool     `json:"enabled"`
	Channel      string   `json:"channel"`
	Names        []string `json:"names" gorm:"serializer:json"`
	Interval     int      `json:"interval" gorm:"default:10"`
	CurrentIndex int      `json:"currentIndex" gorm:"default:0"`
}

type ChannelCounterSettings struct {
	GuildID        string                       `gorm:"primaryKey" json:"guild_id"`
	Enabled        bool                         `gorm:"default:false" json:"enabled"`
	MemberCount    ChannelCountMembersSettings  `gorm:"embedded;embeddedPrefix:membercounter_" json:"membercounter"`
	HumansCount    ChannelCountHumansSettings   `gorm:"embedded;embeddedPrefix:humanscounter_" json:"humanscounter"`
	BotsCount      ChannelCountBotsSettings     `gorm:"embedded;embeddedPrefix:botscounter_" json:"botscounter"`
	ChannelsCount  ChannelCountChannelsSettings `gorm:"embedded;embeddedPrefix:channelscounter_" json:"channelscounter"`
	RolesCount     ChannelCountRolesSettings    `gorm:"embedded;embeddedPrefix:rolescounter_" json:"rolescounter"`
	StatusSettings ChannelCountStatusSettings   `gorm:"embedded;embeddedPrefix:statuscounter_" json:"statuscounter"`
}

type ChannelCounterModule struct {
	Data ChannelCounterSettings
}

func init() {
	Register(&ChannelCounterModule{})
}

func (m *ChannelCounterModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "ChannelCounter",
		Icon: "bar_chart",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_ChannelCounter").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_ChannelCounter").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_ChannelCounter")
			getSub := func(k string) core.SubmoduleMeta {
				if s, ok := meta.Submodules[k]; ok {
					return core.SubmoduleMeta{Label: s.Label, Description: s.Description}
				}
				return core.SubmoduleMeta{}
			}
			return map[string]core.SubmoduleMeta{
				"membercounter":   getSub("membercounter"),
				"humanscounter":   getSub("humanscounter"),
				"botscounter":     getSub("botscounter"),
				"channelscounter": getSub("channelscounter"),
				"rolescounter":    getSub("rolescounter"),
				"statuscounter":   getSub("statuscounter"),
			}
		},
	}
}

func (m *ChannelCounterModule) Priority() int   { return 1 }
func (m *ChannelCounterModule) IsEnabled() bool { return m.Data.Enabled }
func (m *ChannelCounterModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionManageChannels,
	}
}

func getChannelName(layout string, count int) string {
	var name string
	name = utils.ParseVariables(layout, map[string]string{
		"count": utils.ParseCount(count),
	})

	return name
}

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

func (m *ChannelCounterModule) Schema() interface{}  { return &ChannelCounterSettings{} }
func (m *ChannelCounterModule) DataPtr() interface{} { return &m.Data }
func (m *ChannelCounterModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = ChannelCounterSettings{GuildID: guildID}
	err := db.FirstOrCreate(&m.Data, ChannelCounterSettings{GuildID: guildID}).Error

	if m.Data.MemberCount.NameConv == "" {
		m.Data.MemberCount.NameConv = "Members: {count}"
	}
	if m.Data.HumansCount.NameConv == "" {
		m.Data.HumansCount.NameConv = "Humans: {count}"
	}
	if m.Data.BotsCount.NameConv == "" {
		m.Data.BotsCount.NameConv = "Bots: {count}"
	}
	if m.Data.ChannelsCount.NameConv == "" {
		m.Data.ChannelsCount.NameConv = "Channels: {count}"
	}
	if m.Data.RolesCount.NameConv == "" {
		m.Data.RolesCount.NameConv = "Roles: {count}"
	}
	if m.Data.StatusSettings.Interval < 1 {
		m.Data.StatusSettings.Interval = 1
	}

	return err
}

func (m *ChannelCounterModule) UISchema(locale discord.Locale) core.UISchema {
	maxNameLength := 95
	maxStatuses := 10
	meta := locales.GetMeta(locale, "module_ChannelCounter")

	getLabelDesc := func(subK, optK, defL, defD string) (string, string) {
		l, d := defL, defD
		if s, ok := meta.Submodules[subK]; ok {
			if o, ok := s.Options[optK]; ok {
				if o.Label != "" {
					l = o.Label
				}
				if o.Description != "" {
					d = o.Description
				}
			}
		}
		return l, d
	}

	getVarLabelDesc := func(subK, optK, varK, defL, defD string) (string, string) {
		l, d := defL, defD
		if s, ok := meta.Submodules[subK]; ok {
			if o, ok := s.Options[optK]; ok {
				if vOpt, ok := o.Options[varK]; ok {
					if vOpt.Label != "" {
						l = vOpt.Label
					}
					if vOpt.Description != "" {
						d = vOpt.Description
					}
				}
			}
		}
		return l, d
	}

	createCounterComponents := func(subKey, defaultName string) []core.UIComponent {
		enabledL, _ := getLabelDesc(subKey, "enabled", "Enabled", "")
		channelL, channelD := getLabelDesc(subKey, "channel", "Channel", "Select the channel to display this counter.")
		nameconvL, nameconvD := getLabelDesc(subKey, "nameconv", "Channel Name Format", "Use {count} to display the number.")
		countL, countD := getVarLabelDesc(subKey, "nameconv", "count", "Count", "The number to display.")

		return []core.UIComponent{
			{
				Name:     "enabled",
				Label:    enabledL,
				Type:     core.ComponentTypeBoolean,
				Required: false,
			},
			{
				Name:         "channel",
				Label:        channelL,
				Description:  channelD,
				Type:         core.ComponentTypeChannel,
				ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildVoice},
				Required:     true,
				UniqueGroup:  "channel_counter_group",
			},
			{
				Name:        "nameconv",
				Label:       nameconvL,
				Description: nameconvD,
				Placeholder: defaultName,
				Type:        core.ComponentTypeString,
				Required:    true,
				Max:         &maxNameLength,
				Variables: []core.Variables{
					{
						Key:         "count",
						Label:       countL,
						Description: countD,
						Length:      5,
					},
				},
			},
		}
	}

	statusEnabledL, _ := getLabelDesc("statuscounter", "enabled", "Enabled", "")
	statusChannelL, statusChannelD := getLabelDesc("statuscounter", "channel", "Channel", "Select the channel to display this rotating status.")
	statusIntervalL, statusIntervalD := getLabelDesc("statuscounter", "interval", "Rotation Interval (Minutes)", "How often the status changes (1 minute to 1 week).")
	statusNamesL, statusNamesD := getLabelDesc("statuscounter", "names", "Rotating Statuses", "List of statuses.")

	memL, memD := getVarLabelDesc("statuscounter", "names", "member_count", "Member Count", "Total members in the server")
	humL, humD := getVarLabelDesc("statuscounter", "names", "humans_count", "Humans Count", "Total humans in the server")
	botL, botD := getVarLabelDesc("statuscounter", "names", "bots_count", "Bots Count", "Total bots in the server")
	chanL, chanD := getVarLabelDesc("statuscounter", "names", "channels_count", "Channels Count", "Total channels in the server")
	roleL, roleD := getVarLabelDesc("statuscounter", "names", "roles_count", "Roles Count", "Total roles in the server")

	statusVars := []core.Variables{
		{Key: "member_count", Label: memL, Description: memD, Length: 5},
		{Key: "humans_count", Label: humL, Description: humD, Length: 5},
		{Key: "bots_count", Label: botL, Description: botD, Length: 5},
		{Key: "channels_count", Label: chanL, Description: chanD, Length: 5},
		{Key: "roles_count", Label: roleL, Description: roleD, Length: 5},
	}

	getSubLabelDesc := func(k, defL, defD string) (string, string) {
		if s, ok := meta.Submodules[k]; ok {
			return s.Label, s.Description
		}
		return defL, defD
	}

	memSubL, memSubD := getSubLabelDesc("membercounter", "Total Members Counter", "Displays total member count in a channel.")
	humSubL, humSubD := getSubLabelDesc("humanscounter", "Humans Counter", "Displays total human count in a channel.")
	botSubL, botSubD := getSubLabelDesc("botscounter", "Bots Counter", "Displays total bot count in a channel.")
	chanSubL, chanSubD := getSubLabelDesc("channelscounter", "Channels Counter", "Displays total channel count in a channel.")
	roleSubL, roleSubD := getSubLabelDesc("rolescounter", "Roles Counter", "Displays total role count in a channel.")
	statSubL, statSubD := getSubLabelDesc("statuscounter", "Rotating Status Channel", "Configures a channel whose name rotates to display server stats.")

	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "membercounter",
				Label:       memSubL,
				Description: memSubD,
				Components:  createCounterComponents("membercounter", "Members: {count}"),
			},
			{
				Name:        "humanscounter",
				Label:       humSubL,
				Description: humSubD,
				Components:  createCounterComponents("humanscounter", "Humans: {count}"),
			},
			{
				Name:        "botscounter",
				Label:       botSubL,
				Description: botSubD,
				Components:  createCounterComponents("botscounter", "Bots: {count}"),
			},
			{
				Name:        "channelscounter",
				Label:       chanSubL,
				Description: chanSubD,
				Components:  createCounterComponents("channelscounter", "Channels: {count}"),
			},
			{
				Name:        "rolescounter",
				Label:       roleSubL,
				Description: roleSubD,
				Components:  createCounterComponents("rolescounter", "Roles: {count}"),
			},
			{
				Name:        "statuscounter",
				Label:       statSubL,
				Description: statSubD,
				Components: []core.UIComponent{
					{
						Name:     "enabled",
						Label:    statusEnabledL,
						Type:     core.ComponentTypeBoolean,
						Required: false,
					},
					{
						Name:         "channel",
						Label:        statusChannelL,
						Description:  statusChannelD,
						Type:         core.ComponentTypeChannel,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildVoice},
						Required:     true,
						UniqueGroup:  "channel_counter_group",
					},
					{
						Name:        "interval",
						Label:       statusIntervalL,
						Description: statusIntervalD,
						Placeholder: "10",
						Type:        core.ComponentTypeNumber,
						Required:    true,
						Min:         func() *int { v := 1; return &v }(),
						Max:         func() *int { v := 10080; return &v }(),
					},
					{
						Name:        "names",
						Label:       statusNamesL,
						Description: statusNamesD,
						Type:        core.ComponentTypeList,
						ListType:    "string",
						Required:    true,
						Max:         &maxStatuses,
						ItemMax:     &maxNameLength,
						Variables:   statusVars,
					},
				},
			},
		},
	}
}
