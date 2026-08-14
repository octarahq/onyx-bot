package safety

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"onyx/bot/core"
	"onyx/bot/locales"
	
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type SafetyAntiMassJoinLevel int

const (
	SafetyAntiMassJoinLevelNone   SafetyAntiMassJoinLevel = 0
	SafetyAntiMassJoinLevelSoft   SafetyAntiMassJoinLevel = 1
	SafetyAntiMassJoinLevelMedium SafetyAntiMassJoinLevel = 2
	SafetyAntiMassJoinLevelHight  SafetyAntiMassJoinLevel = 3
)

type SafetyAntiSpamLevel int

const (
	SafetyAntiSpamLevelNone   SafetyAntiSpamLevel = 0
	SafetyAntiSpamLevelSoft   SafetyAntiSpamLevel = 1
	SafetyAntiSpamLevelMedium SafetyAntiSpamLevel = 2
	SafetyAntiSpamLevelHight  SafetyAntiSpamLevel = 3
)

func (l *SafetyAntiMassJoinLevel) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*l = SafetyAntiMassJoinLevel(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if parsed, err := strconv.Atoi(s); err == nil {
			*l = SafetyAntiMassJoinLevel(parsed)
			return nil
		}
	}
	return fmt.Errorf("invalid type for SafetyAntiMassJoinLevel")
}

func (l SafetyAntiMassJoinLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(l)))
}

func (l *SafetyAntiSpamLevel) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*l = SafetyAntiSpamLevel(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if parsed, err := strconv.Atoi(s); err == nil {
			*l = SafetyAntiSpamLevel(parsed)
			return nil
		}
	}
	return fmt.Errorf("invalid type for SafetyAntiSpamLevel")
}

func (l SafetyAntiSpamLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(l)))
}

type SafetyARaidSettings struct {
	AltDetector       bool                    `json:"alt_detector"`
	AntiMassJoinLevel SafetyAntiMassJoinLevel `json:"anti_massjoin_level"`
	AntiBot           bool                    `json:"anti_bot"`
}

type SafetyASpamSettings struct {
	QuarentineRole  string              `json:"quarentine_role"`
	AntiSpamLevel   SafetyAntiSpamLevel `json:"anti_spam"`
	AntiPhishing    bool                `json:"anti_phishing"`
	BlockInviteLink bool                `json:"anti_invite"`
	AntiMention     bool                `json:"anti_mention"`
	AntiMassEmoji   bool                `json:"anti_mass_emoji"`
	AntiZalgo       bool                `json:"anti_zalgo"`
	IgnoredChannels string              `json:"ignored_channels"`
}

type SafetyANukeSettings struct {
	AntiMassKick             bool `json:"anti_mass_kick"`
	AntiMassChannelD         bool `json:"anti_mass_channel_delete"`
	AntiMassRoleD            bool `json:"anti_mass_role_delete"`
	AntiVanityUrlEdit        bool `json:"anti_vanity_url_edit"`
	AntiDangerousPermissions bool `json:"anti_danger_permission"`
}

type SafetyCaptchaSettings struct {
	Enabled       bool   `json:"enabled"`
	Channel       string `json:"channel"`
	VerifiedRole  string `json:"vrole"`
	ShowToSusUser bool   `json:"show_to_sus"`
}

type CaptchaSession struct {
	UserID        string         `json:"user_id"`
	CorrectAnswer string         `json:"correct_answer"`
	StartedAt     time.Time      `json:"started_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	Attempts      int            `json:"attempts"`
	MaxAttempts   int            `json:"max_attempts"`
	Status        string         `json:"status"`
	IsSuspect     bool           `json:"is_suspect"`
	BackupRoles   []snowflake.ID `json:"backup_roles"`
}

type SafetySaveGuildState struct {
	AntiMassJoinOldVerifLevel   discord.VerificationLevel `json:"anti_mass_join_old_verif_level"`
	AntiMassJoinOldEveryonePerm discord.Permissions       `json:"anti_mass_join_old_everyone_perm"`
	CaptchaSessions             map[string]CaptchaSession `json:"captcha_sessions"`
}

type SafetySettings struct {
	GuildID   string                `gorm:"primaryKey" json:"guild_id"`
	Enabled   bool                  `gorm:"default:false" json:"enabled"`
	AntiRaid  SafetyARaidSettings   `gorm:"embedded;embeddedPrefix:antiraid_" json:"antiraid"`
	AntiSpam  SafetyASpamSettings   `gorm:"embedded;embeddedPrefix:antispam_" json:"antispam"`
	AntiNuke  SafetyANukeSettings   `gorm:"embedded;embeddedPrefix:antinuke_" json:"antinuke"`
	Captcha   SafetyCaptchaSettings `gorm:"embedded;embeddedPrefix:captcha_" json:"captcha"`
	SaveState SafetySaveGuildState  `gorm:"serializer:json" json:"-"`
}

type SafetyModule struct {
	Data SafetySettings
}

type AntiSpamCacheItem struct {
	LastMessageContent string
	LastMessageTime    time.Time
}

func init() {
	core.Register(&SafetyModule{})
}

func (m *SafetyModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "SafetyModule",
		Icon: "shield",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_SafetyModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_SafetyModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_SafetyModule")
			subs := make(map[string]core.SubmoduleMeta)
			for k, v := range meta.Submodules {
				subs[k] = core.SubmoduleMeta{
					Label:       v.Label,
					Description: v.Description,
				}
			}
			return subs
		},
		Loggable: true,
	}
}

func (m *SafetyModule) Priority() int   { return 1 }
func (m *SafetyModule) IsEnabled() bool { return m.Data.Enabled }
func (m *SafetyModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionManageChannels,
		discord.PermissionManageGuild,
		discord.PermissionManageMessages,
		discord.PermissionBanMembers,
		discord.PermissionKickMembers,
		discord.PermissionModerateMembers,
	}
}

func (m *SafetyModule) Schema() interface{}  { return &SafetySettings{} }
func (m *SafetyModule) DataPtr() interface{} { return &m.Data }
func (m *SafetyModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = SafetySettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, SafetySettings{GuildID: guildID}).Error
}

func (m *SafetyModule) UISchema(locale discord.Locale) core.UISchema {
	meta := locales.GetMeta(locale, "module_SafetyModule")

	getOpt := func(sub, optName string) (string, string) {
		if s, ok := meta.Submodules[sub]; ok {
			if o, ok := s.Options[optName]; ok {
				return o.Label, o.Description
			}
		}
		return "", ""
	}

	getEnum := func(sub, optName, enumKey string) string {
		if s, ok := meta.Submodules[sub]; ok {
			if o, ok := s.Options[optName]; ok {
				if e, ok := o.Options[enumKey]; ok {
					return e.Label
				}
			}
		}
		return ""
	}

	arAltL, arAltD := getOpt("antiraid", "alt_detector")
	arBotL, arBotD := getOpt("antiraid", "anti_bot")
	arMassL, arMassD := getOpt("antiraid", "anti_massjoin_level")

	asLevelL, asLevelD := getOpt("antispam", "anti_spam")
	asRoleL, asRoleD := getOpt("antispam", "quarentine_role")
	asPhishL, asPhishD := getOpt("antispam", "anti_phishing")
	asInviteL, asInviteD := getOpt("antispam", "anti_invite")
	asMentionL, asMentionD := getOpt("antispam", "anti_mention")
	asEmojiL, asEmojiD := getOpt("antispam", "anti_mass_emoji")
	asZalgoL, asZalgoD := getOpt("antispam", "anti_zalgo")
	asIgnoredL, asIgnoredD := getOpt("antispam", "ignored_channels")

	anKickL, anKickD := getOpt("antinuke", "anti_mass_kick")
	anChanL, anChanD := getOpt("antinuke", "anti_mass_channel_delete")
	anRoleL, anRoleD := getOpt("antinuke", "anti_mass_role_delete")
	anUrlL, anUrlD := getOpt("antinuke", "anti_vanity_url_edit")
	anPermL, anPermD := getOpt("antinuke", "anti_danger_permission")

	cEnL, cEnD := getOpt("captcha", "enabled")
	cChanL, cChanD := getOpt("captcha", "channel")
	cRoleL, cRoleD := getOpt("captcha", "vrole")
	cSusL, cSusD := getOpt("captcha", "show_to_sus")

	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "antiraid",
				Label:       meta.Submodules["antiraid"].Label,
				Description: meta.Submodules["antiraid"].Description,
				Components: []core.UIComponent{
					{
						Name:        "alt_detector",
						Label:       arAltL,
						Description: arAltD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_bot",
						Label:       arBotL,
						Description: arBotD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_massjoin_level",
						Label:       arMassL,
						Description: arMassD,
						Type:        core.ComponentTypeSelect,
						Options: []core.UISelectOption{
							{Label: getEnum("antiraid", "anti_massjoin_level", "0"), Value: "0"},
							{Label: getEnum("antiraid", "anti_massjoin_level", "1"), Value: "1"},
							{Label: getEnum("antiraid", "anti_massjoin_level", "2"), Value: "2"},
							{Label: getEnum("antiraid", "anti_massjoin_level", "3"), Value: "3"},
						},
					},
				},
			},
			{
				Name:        "antispam",
				Label:       meta.Submodules["antispam"].Label,
				Description: meta.Submodules["antispam"].Description,
				Components: []core.UIComponent{
					{
						Name:        "quarentine_role",
						Label:       asRoleL,
						Description: asRoleD,
						Type:        core.ComponentTypeRole,
					},
					{
						Name:        "anti_spam",
						Label:       asLevelL,
						Description: asLevelD,
						Type:        core.ComponentTypeSelect,
						Options: []core.UISelectOption{
							{Label: getEnum("antispam", "anti_spam", "0"), Value: "0"},
							{Label: getEnum("antispam", "anti_spam", "1"), Value: "1"},
							{Label: getEnum("antispam", "anti_spam", "2"), Value: "2"},
							{Label: getEnum("antispam", "anti_spam", "3"), Value: "3"},
						},
					},
					{
						Name:        "anti_phishing",
						Label:       asPhishL,
						Description: asPhishD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_invite",
						Label:       asInviteL,
						Description: asInviteD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_mention",
						Label:       asMentionL,
						Description: asMentionD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_mass_emoji",
						Label:       asEmojiL,
						Description: asEmojiD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_zalgo",
						Label:       asZalgoL,
						Description: asZalgoD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:         "ignored_channels",
						Label:        asIgnoredL,
						Description:  asIgnoredD,
						Type:         core.ComponentTypeChannel,
						Multiple:     true,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
				},
			},
			{
				Name:        "antinuke",
				Label:       meta.Submodules["antinuke"].Label,
				Description: meta.Submodules["antinuke"].Description,
				Components: []core.UIComponent{
					{
						Name:        "anti_mass_kick",
						Label:       anKickL,
						Description: anKickD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_mass_channel_delete",
						Label:       anChanL,
						Description: anChanD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_mass_role_delete",
						Label:       anRoleL,
						Description: anRoleD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_vanity_url_edit",
						Label:       anUrlL,
						Description: anUrlD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:        "anti_danger_permission",
						Label:       anPermL,
						Description: anPermD,
						Type:        core.ComponentTypeBoolean,
					},
				},
			},
			{
				Name:        "captcha",
				Label:       meta.Submodules["captcha"].Label,
				Description: meta.Submodules["captcha"].Description,
				Components: []core.UIComponent{
					{
						Name:        "enabled",
						Label:       cEnL,
						Description: cEnD,
						Type:        core.ComponentTypeBoolean,
					},
					{
						Name:         "channel",
						Label:        cChanL,
						Description:  cChanD,
						Type:         core.ComponentTypeChannel,
						ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
					},
					{
						Name:        "vrole",
						Label:       cRoleL,
						Description: cRoleD,
						Type:        core.ComponentTypeRole,
					},
					{
						Name:        "show_to_sus",
						Label:       cSusL,
						Description: cSusD,
						Type:        core.ComponentTypeBoolean,
					},
				},
			},
		},
	}
}
