package modules

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"

	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"gorm.io/gorm"
)

type SafetyAntiMassJoinLevel int

const (
	SafetyAntiMassJoinLevelNone   SafetyAntiMassJoinLevel = 0 // ne rien faire
	SafetyAntiMassJoinLevelSoft   SafetyAntiMassJoinLevel = 1 // blockerles permissions denvoie de message pour @everyone
	SafetyAntiMassJoinLevelMedium SafetyAntiMassJoinLevel = 2 // activation du mode de verification maximal
	SafetyAntiMassJoinLevelHight  SafetyAntiMassJoinLevel = 3 // mets en quarentaine (bloque toutes les permissions) puis kick a la fin du raid
)

type SafetyAntiSpamLevel int

const (
	SafetyAntiSpamLevelNone   SafetyAntiSpamLevel = 0 // ne rien faire
	SafetyAntiSpamLevelSoft   SafetyAntiSpamLevel = 1 // bloquer si un contenu dupplique a été envoye en moins de 5 sec (seulement liens + petits message)
	SafetyAntiSpamLevelMedium SafetyAntiSpamLevel = 2 // bloque le spam de message 2 en moins de 2 secondes
	SafetyAntiSpamLevelHight  SafetyAntiSpamLevel = 3 // bloque les contenu duplique si 2 message daffile se suivent
)

type SafetyARaidSettings struct {
	AltDetector       bool                    `json:"alt_detector"`
	AntiMassJoinLevel SafetyAntiMassJoinLevel `json:"anti_massjoin_level"`
	AntiBot           bool                    `json:"anti_bot"`
}

type SafetyASpamSettings struct {
	QuarentineRole  string              `json:"quarentine_role"`
	AntiSpamLevel   SafetyAntiSpamLevel `json:"anti_spam"`
	AntiPhishing    bool                `json:"anti_phishing"` // blacklist + levenshtein
	BlockInviteLink bool                `json:"anti_invite"`
	AntiMention     bool                `json:"anti_mention"`
	AntiMassEmoji   bool                `json:"anti_mass_emoji"`
	AntiZalgo       bool                `json:"anti_zalgo"`
	IgnoredChannels string              `json:"ignored_channels"` // liste des salons/categories ignores
}

type SafetyANukeSettings struct {
	AntiMassKick             bool `json:"anti_mass_kick"`
	AntiMassChannelD         bool `json:"anti_mass_channel_delete"` // dedans compte aussi le mass channel edit
	AntiMassRoleD            bool `json:"anti_mass_role_delete"`    // dedans compte aussi le mass role edit
	AntiVanityUrlEdit        bool `json:"anti_vanity_url_edit"`
	AntiDangerousPermissions bool `json:"anti_danger_permission"`
}

type SafetyCaptchaSettings struct {
	Enabled       bool   `json:"enabled"`
	Channel       string `json:"channel"`
	VerifiedRole  string `json:"vrole"`
	ShowToSusUser bool   `json:"show_to_sus"` // dans un cas ou un anti spam ou autre est active si cette personne sest deja faite verifie precedement elle devra faire le captcha pour ne pas etre kick
}

type SafetySaveGuildState struct { // servira a garder lancien etat du serveur (nest pas affiche sur le dash)
	AntiMassJoinOldVerifLevel discord.VerificationLevel
}

type SafetySettings struct {
	GuildID  string                `gorm:"primaryKey" json:"guild_id"`
	Enabled  bool                  `gorm:"default:false" json:"enabled"`
	AntiRaid SafetyARaidSettings   `gorm:"embedded;embeddedPrefix:antiraid_" json:"antiraid"`
	AntiSpam SafetyASpamSettings   `gorm:"embedded;embeddedPrefix:antispam_" json:"antispam"`
	AntiNuke SafetyANukeSettings   `gorm:"embedded;embeddedPrefix:antinuke_" json:"antinuke"`
	Captcha  SafetyCaptchaSettings `gorm:"embedded;embeddedPrefix:captcha_" json:"captcha"`
}

type SafetyModule struct {
	Data SafetySettings
}

func init() {
	Register(&SafetyModule{})
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
	}
}

func (m *SafetyModule) Priority() int   { return 1 }
func (m *SafetyModule) IsEnabled() bool { return m.Data.Enabled }
func (m *SafetyModule) Permissions() []discord.Permissions {
	return []discord.Permissions{}
}

func (m *SafetyModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if m.Data.Enabled {
		if m.Data.AntiSpam.AntiPhishing {
			if handlePhishing(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.BlockInviteLink {
			if handleBlockInvite(b, e.Client(), e.Message) {
				return true
			}
		}
	}
	return false
}

func (m *SafetyModule) HandleMessageUpdate(b *core.Bot, e *events.MessageUpdate) bool {
	if m.Data.Enabled {
		if m.Data.AntiSpam.AntiPhishing {
			if handlePhishing(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.BlockInviteLink {
			if handleBlockInvite(b, e.Client(), e.Message) {
				return true
			}
		}
	}
	return false
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
						Name:        "quarentine_role",
						Label:       asRoleL,
						Description: asRoleD,
						Type:        core.ComponentTypeRole,
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
						Name:        "ignored_channels",
						Label:       asIgnoredL,
						Description: asIgnoredD,
						Type:        core.ComponentTypeChannel,
						Multiple:    true,
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
						Name:        "channel",
						Label:       cChanL,
						Description: cChanD,
						Type:        core.ComponentTypeChannel,
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

func handleBlockInvite(b *core.Bot, client *bot.Client, message discord.Message) bool {
	urls := utils.ExtractURLs(message.Content)
	var includeInvite bool

	for _, url := range urls {
		host := strings.ToLower(url.Hostname())

		switch {
		case host == "discord.gg" || strings.HasSuffix(host, ".discord.gg"):
			includeInvite = true

		case host == "discord.com" || strings.HasSuffix(host, ".discord.com"):
			if strings.HasPrefix(url.Path, "/invite") || strings.HasPrefix(url.Path, "/application-directory") {
				includeInvite = true
			} else if strings.HasPrefix(url.Path, "/oauth2/authorize") {
				includeInvite = true
			}
		}

		if includeInvite {
			break
		}
	}

	if includeInvite {
		client.Rest.DeleteMessage(message.ChannelID, message.ID)
		code := b.Logger.SendSafetyBlockedInviteLogs(urls, message)
		locale := discord.LocaleEnglishUS
		if message.GuildID != nil {
			if guild, ok := client.Caches.Guild(*message.GuildID); ok {
				locale = discord.Locale(guild.PreferredLocale)
			}
		}
		trad := locales.GetModule_SafetyModule(locale)

		title := trad.BlockedInvite_censored_title
		if title == "" {
			title = "Ton message a été censuré car il contient une invitation."
		}
		desc := trad.BlockedInvite_censored_description
		if desc == "" {
			desc = "-# S'il s'agit d'une erreur contactez le support. Code : %s"
		}

		sendCensoredMessage(client, message, string(code), title, desc)
		return true
	}

	return false
}

func handlePhishing(b *core.Bot, client *bot.Client, message discord.Message) bool {
	content := message.Content
	urls := utils.ExtractURLs(content)

	isPhishing := false

	for _, u := range urls {
		hostname := strings.ToLower(u.Hostname())
		if len(hostname) == 0 {
			continue
		}

		firstChar := rune(hostname[0])
		fileName := string(firstChar) + ".csv"
		filePath := filepath.Join("data", "security", "domainlist", fileName)

		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		file.Close()
		if err != nil {
			continue
		}

		isExactMatch := false
		matchedPhishing := false

		for i, record := range records {
			if i == 0 {
				continue
			}
			if len(record) < 2 {
				continue
			}
			legitDomain := record[1]
			if hostname == legitDomain {
				isExactMatch = true
				matchedPhishing = false
				break
			}
			if !matchedPhishing {
				dist := levenshtein.DistanceForStrings([]rune(hostname), []rune(legitDomain), levenshtein.DefaultOptions)
				if dist <= 2 {
					matchedPhishing = true
				}
			}
		}

		if matchedPhishing && !isExactMatch {
			isPhishing = true
			break
		}
	}

	mds := utils.ExtractMDURLs(content)
	for _, md := range mds {
		name, link := md.Name, md.URL

		nameDomain := utils.ExtractDomain(name)

		if nameDomain == "" {
			continue
		}

		linkDomain := strings.ToLower(link.Hostname())
		linkDomain = strings.TrimPrefix(linkDomain, "www.")

		if nameDomain != linkDomain {
			isPhishing = true
		}
	}

	if isPhishing {
		client.Rest.DeleteMessage(message.ChannelID, message.ID)
		code := b.Logger.SendSafetyPhishingLogs(urls, message)
		locale := discord.LocaleEnglishUS
		if message.GuildID != nil {
			if guild, ok := client.Caches.Guild(*message.GuildID); ok {
				locale = discord.Locale(guild.PreferredLocale)
			}
		}
		trad := locales.GetModule_SafetyModule(locale)

		title := trad.Phishing_censored_title
		if title == "" {
			title = "Ton lien a été censuré pour suspicion de phishing."
		}
		desc := trad.Phishing_censored_description
		if desc == "" {
			desc = "-# S'il s'agit d'une erreur contactez le support. Code : %s"
		}

		sendCensoredMessage(client, message, string(code), title, desc)
		return true
	}

	return false
}

func sendCensoredMessage(client *bot.Client, message discord.Message, code string, title string, desc string) {
	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplayf("# <@%s>", message.Author.ID.String()),
				discord.NewTextDisplay(title),
				discord.NewTextDisplayf(desc, code),
			).WithAccessory(discord.NewThumbnail(message.Author.EffectiveAvatarURL())),
			discord.NewActionRow(
				discord.NewLinkButton("Support", "https://onyx.octara.xyz"),
			),
		),
	).WithAllowedMentions(&discord.AllowedMentions{
		Users: []snowflake.ID{message.Author.ID},
	})

	client.Rest.CreateMessage(message.ChannelID, msg)
}
