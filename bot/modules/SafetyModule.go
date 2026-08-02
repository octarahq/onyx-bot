package modules

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/omit"
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
	AltDetector       bool                    `json:"alt_detector"` //done
	AntiMassJoinLevel SafetyAntiMassJoinLevel `json:"anti_massjoin_level"`
	AntiBot           bool                    `json:"anti_bot"` //done
}

type SafetyASpamSettings struct { //done
	QuarentineRole  string              `json:"quarentine_role"`
	AntiSpamLevel   SafetyAntiSpamLevel `json:"anti_spam"` //not done
	AntiPhishing    bool                `json:"anti_phishing"`
	BlockInviteLink bool                `json:"anti_invite"`
	AntiMention     bool                `json:"anti_mention"`
	AntiMassEmoji   bool                `json:"anti_mass_emoji"`
	AntiZalgo       bool                `json:"anti_zalgo"`
	IgnoredChannels string              `json:"ignored_channels"`
}

type SafetyANukeSettings struct { //done
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

func (m *SafetyModule) HandleGuildMemberJoin(b *core.Bot, e *events.GuildMemberJoin) bool {
	if m.Data.Enabled {
		if m.Data.AntiRaid.AntiBot {
			if e.Member.User.Bot {
				e.Client().Rest.RemoveMember(e.GuildID, e.Member.User.ID)
				b.Logger.SendSafetyRaidBotLogs(e.Member)
				return true
			}
		}

		if m.Data.AntiRaid.AltDetector {
			if time.Since(e.Member.CreatedAt()) < 7*24*time.Hour {
				e.Client().Rest.RemoveMember(e.GuildID, e.Member.User.ID)
				b.Logger.SendSafetyRaidAltLogs(e.Member)
				return true
			}
		}

		if m.Data.Captcha.Enabled {
			channel, err := e.Client().Rest.CreateDMChannel(e.Member.User.ID)
			if err != nil {
				cid, parseErr := snowflake.Parse(m.Data.Captcha.Channel)
				if parseErr != nil {
					return false
				}
				_ = cid
			} else {
				cid := channel.ID()

				guild, exist := e.Client().Caches.Guild(e.GuildID)
				if !exist {
					return false
				}

				msg, _, _ := utils.CaptchaBuildMessage("", e.Member, guild)

				e.Client().Rest.CreateMessage(cid, msg)
				return false
			}

			cid, err := snowflake.Parse(m.Data.Captcha.Channel)
			if err != nil {
				return false
			}

			guild, exist := e.Client().Caches.Guild(e.GuildID)
			if !exist {
				return false
			}

			msg, _, _ := utils.CaptchaBuildMessage("ouit", e.Member, guild)

			_, err = e.Client().Rest.CreateMessage(cid, msg)
		}
	}

	return false
}

func (m *SafetyModule) HandleComponentInteractionCreate(b *core.Bot, e *events.ComponentInteractionCreate) bool {
	if strings.HasPrefix(e.Data.CustomID(), "module-safety-") && strings.Contains(e.Data.CustomID(), "-captcha-resolve-") {
		parts := strings.Split(e.Data.CustomID(), "-")
		if len(parts) < 6 {
			return false
		}
		guildIdStr := parts[2]
		sessionid := parts[5]
		isSelect := rand.Intn(2) == 1
		modal := discord.NewModalCreate(fmt.Sprintf("module-safety-%s-captcha-solution-%s", guildIdStr, sessionid), "What is your response ?")

		letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		b := make([]rune, 5)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		resFieldCId := fmt.Sprintf("field-%s", string(b))

		possiblePlaceholder := []string{
			"Enter your captcha answer here",
			"Type the captcha solution",
			"Input your captcha response",
			"Put your answer in this field",
			"Submit your captcha code",
		}

		if isSelect {
			var opts []discord.StringSelectMenuOption
			for i := range 5 {
				opts = append(opts, discord.NewStringSelectMenuOption(fmt.Sprintf("Number %d", i), strconv.Itoa(i)))
			}
			modal = modal.AddLabel(
				"Captcha answer",
				discord.NewStringSelectMenu(resFieldCId, possiblePlaceholder[rand.Intn(len(possiblePlaceholder))], opts...),
			)
		} else {
			modal = modal.AddLabel(
				"Captcha answer",
				discord.NewShortTextInput(resFieldCId).WithPlaceholder(possiblePlaceholder[rand.Intn(len(possiblePlaceholder))]),
			)
		}

		err := e.Modal(modal)
		if err != nil {
			fmt.Println(err)
		}
	}

	return false
}

func (m *SafetyModule) HandleGuildUpdate(b *core.Bot, e *events.GuildUpdate) bool {
	if m.Data.Enabled {
		if m.Data.AntiNuke.AntiVanityUrlEdit {
			oldCode := ""
			if e.OldGuild.VanityURLCode != nil {
				oldCode = *e.OldGuild.VanityURLCode
			}
			newCode := ""
			if e.Guild.VanityURLCode != nil {
				newCode = *e.Guild.VanityURLCode
			}

			if oldCode != newCode {
				var userID snowflake.ID

				auditLogs, err := e.Client().Rest.GetAuditLog(e.GuildID, 0, discord.AuditLogEventGuildUpdate, 0, 0, 1)
				if err == nil && len(auditLogs.AuditLogEntries) > 0 {
					entry := auditLogs.AuditLogEntries[0]
					if entry.UserID != 0 && entry.UserID != e.Guild.OwnerID && entry.UserID != e.Client().ID() {
						userID = entry.UserID
					}
				}

				if userID != 0 {
					type UpdateVanity struct {
						Code string `json:"code"`
					}
					ep := rest.NewEndpoint(http.MethodPatch, "/guilds/{guild.id}/vanity-url")
					e.Client().Rest.Do(ep.Compile(nil, e.GuildID), UpdateVanity{Code: oldCode}, nil)

					member, err := e.Client().Rest.GetMember(e.GuildID, userID)
					if err == nil && member != nil {
						_, dperms := utils.CheckDangerousPermissions(e.Client(), *member)
						utils.RemoveMemberPerms(e.Client(), *member, dperms)
					}

					locale := discord.LocaleFrench
					if e.Guild.PreferredLocale != "" {
						locale = discord.Locale(e.Guild.PreferredLocale)
					}
					trad := locales.GetModule_SafetyModule(locale)

					ownerChannel, err := e.Client().Rest.CreateDMChannel(e.Guild.OwnerID)
					if err == nil {
						msg := discord.NewMessageCreateV2(
							discord.NewContainer(
								discord.NewSection(
									discord.NewTextDisplayf(trad.Vanity_nuke_alert, userID, e.Guild.Name),
								).WithAccessory(discord.NewThumbnail(*e.Guild.IconURL())),
							).WithAccentColor(utils.ParseStrColor("#e74c3c")),
						)
						e.Client().Rest.CreateMessage(ownerChannel.ID(), msg)
					}

					return true
				}
			}
		}
	}

	return false
}

func (m *SafetyModule) HandleGuildMemberUpdate(b *core.Bot, e *events.GuildMemberUpdate) bool {
	if m.Data.Enabled && m.Data.AntiNuke.AntiDangerousPermissions {
		oldRoles := make(map[snowflake.ID]bool)
		for _, roleID := range e.OldMember.RoleIDs {
			oldRoles[roleID] = true
		}

		var rolesToRemove []snowflake.ID

		for _, roleID := range e.Member.RoleIDs {
			if !oldRoles[roleID] {
				if role, ok := e.Client().Caches.Role(e.GuildID, roleID); ok {
					if role.Permissions.Has(discord.PermissionAdministrator) || role.Permissions.Has(discord.PermissionManageGuild) {
						rolesToRemove = append(rolesToRemove, roleID)
					}
				}
			}
		}

		if len(rolesToRemove) > 0 {
			var authorID snowflake.ID
			auditLogs, err := e.Client().Rest.GetAuditLog(e.GuildID, 0, discord.AuditLogEventMemberRoleUpdate, 0, 0, 1)
			if err == nil && len(auditLogs.AuditLogEntries) > 0 {
				entry := auditLogs.AuditLogEntries[0]
				if entry.TargetID != nil && *entry.TargetID == e.Member.User.ID {
					guild, exist := b.Client.Caches.Guild(e.GuildID)
					if !exist {
						return false
					}
					if entry.UserID != 0 && entry.UserID != guild.OwnerID && entry.UserID != e.Client().ID() {
						authorID = entry.UserID
					}
				}
			}

			for _, roleID := range rolesToRemove {
				e.Client().Rest.RemoveMemberRole(e.GuildID, e.Member.User.ID, roleID)
			}

			if authorID != 0 {
				authorMember, err := e.Client().Rest.GetMember(e.GuildID, authorID)
				if err == nil && authorMember != nil {
					_, dperms := utils.CheckDangerousPermissions(e.Client(), *authorMember)
					utils.RemoveMemberPerms(e.Client(), *authorMember, dperms)
				}
			}
			return true
		}
	}

	return false
}

func (m *SafetyModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if m.Data.Enabled {
		ignore := false
		for _, cid := range strings.Split(m.Data.AntiSpam.IgnoredChannels, ",") {
			scid, err := snowflake.Parse(cid)
			if err == nil {
				if e.ChannelID == scid {
					ignore = true
					break
				}
			}
		}

		if ignore {
			return false
		}

		if m.Data.AntiSpam.AntiPhishing {
			if m.handlePhishing(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.BlockInviteLink {
			if m.handleBlockInvite(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.AntiZalgo {
			if m.handleZalgo(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.AntiMention {
			if m.handleMentionSpam(b, e.Client(), e.Message) {
				return true
			}
		}
	}
	return false
}

func (m *SafetyModule) HandleMessageUpdate(b *core.Bot, e *events.MessageUpdate) bool {
	if m.Data.Enabled {
		ignore := false
		for _, cid := range strings.Split(m.Data.AntiSpam.IgnoredChannels, ",") {
			scid, err := snowflake.Parse(cid)
			if err == nil {
				if e.ChannelID == scid {
					ignore = true
					break
				}
			}
		}

		if ignore {
			return false
		}
		if m.Data.AntiSpam.AntiPhishing {
			if m.handlePhishing(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.BlockInviteLink {
			if m.handleBlockInvite(b, e.Client(), e.Message) {
				return true
			}
		}

		if m.Data.AntiSpam.AntiZalgo {
			if m.handleZalgo(b, e.Client(), e.Message) {
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
						Type:        core.ComponentTypeChannel, //modifier la desc pour dire que cest le backup si jamais le bot peut pas le mp
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

func (m *SafetyModule) giveQuarentineRole(client *bot.Client, member discord.Member) {
	if m.Data.AntiSpam.QuarentineRole != "" {
		gid, _ := snowflake.Parse(m.Data.GuildID)
		rid, err := snowflake.Parse(m.Data.AntiSpam.QuarentineRole)
		if err != nil {
			return
		}

		client.Rest.AddMemberRole(gid, member.User.ID, rid)
	}
}

func (m *SafetyModule) handleZalgo(b *core.Bot, client *bot.Client, message discord.Message) bool {
	content := message.Content
	if len(content) == 0 {
		return false
	}

	var diacriticCount int
	var totalRunes int

	for _, r := range content {
		totalRunes++

		if unicode.Is(unicode.Mn, r) {
			diacriticCount++
		}
	}

	if totalRunes == 0 {
		return false
	}

	ratio := float64(diacriticCount) / float64(totalRunes)

	isZalgo := false
	if ratio > 0.30 || diacriticCount > 15 {
		isZalgo = true
	}

	if isZalgo {
		client.Rest.DeleteMessage(message.ChannelID, message.ID)
		m.giveQuarentineRole(client, *message.Member)
		code := b.Logger.SendSafetyZalgoLogs(ratio, message)
		locale := discord.LocaleEnglishUS
		if message.GuildID != nil {
			if guild, ok := client.Caches.Guild(*message.GuildID); ok {
				locale = discord.Locale(guild.PreferredLocale)
			}
		}
		trad := locales.GetModule_SafetyModule(locale)

		title := trad.Zalgo_censored_title
		if title == "" {
			title = "Ton message a été censuré car il contient du texte corrompu (Zalgo)."
		}
		desc := trad.Zalgo_censored_description
		if desc == "" {
			desc = "-# S'il s'agit d'une erreur contactez le support. Code : %s"
		}

		sendCensoredMessage(client, message, string(code), title, desc)
		return true
	}

	return false
}

func (m *SafetyModule) handleBlockInvite(b *core.Bot, client *bot.Client, message discord.Message) bool {
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
		m.giveQuarentineRole(client, *message.Member)
		code := b.Logger.SendSafetyBlockedInviteLogs(urls, message)
		locale := discord.LocaleEnglishUS
		if message.GuildID != nil {
			if guild, ok := client.Caches.Guild(*message.GuildID); ok {
				locale = discord.Locale(guild.PreferredLocale)
			}
		}
		trad := locales.GetModule_SafetyModule(locale)

		title := trad.Blocked_invite_censored_title
		if title == "" {
			title = "Ton message a été censuré car il contient une invitation."
		}
		desc := trad.Blocked_invite_censored_description
		if desc == "" {
			desc = "-# S'il s'agit d'une erreur contactez le support. Code : %s"
		}

		sendCensoredMessage(client, message, string(code), title, desc)
		return true
	}

	return false
}

func (m *SafetyModule) handlePhishing(b *core.Bot, client *bot.Client, message discord.Message) bool {
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
		m.giveQuarentineRole(client, *message.Member)
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

func (m *SafetyModule) handleMentionSpam(b *core.Bot, client *bot.Client, message discord.Message) bool {
	if message.GuildID == nil {
		return false
	}

	guildID := *message.GuildID
	userID := message.Author.ID
	mentionCount := len(message.Mentions)

	cacheKey := "mention_spam:" + guildID.String() + ":" + userID.String()

	utils.Cache.Mu.Lock()
	var userCounts []int
	if val, ok := utils.Cache.Items[cacheKey]; ok {
		if val.Expiration == 0 || time.Now().UnixNano() <= val.Expiration {
			userCounts = val.Value.([]int)
		}
	}

	userCounts = append(userCounts, mentionCount)
	if len(userCounts) > 3 {
		userCounts = userCounts[len(userCounts)-3:]
	}

	totalMentions := 0
	for _, count := range userCounts {
		totalMentions += count
	}

	shouldTimeout := totalMentions >= 3
	if shouldTimeout {
		delete(utils.Cache.Items, cacheKey)
	} else {
		utils.Cache.Items[cacheKey] = utils.CacheItem{
			Value:      userCounts,
			Expiration: time.Now().Add(5 * time.Minute).UnixNano(),
		}
	}
	utils.Cache.Mu.Unlock()

	if shouldTimeout {
		timeoutDuration := 6 * time.Hour
		until := time.Now().Add(timeoutDuration)

		go func() {
			client.Rest.UpdateMember(guildID, userID, discord.MemberUpdate{
				CommunicationDisabledUntil: omit.New(&until),
			})
		}()

		client.Rest.DeleteMessage(message.ChannelID, message.ID)
		m.giveQuarentineRole(client, *message.Member)
		code := b.Logger.SendSafetyMentionSpamLogs(message)

		locale := discord.LocaleEnglishUS
		if guild, ok := client.Caches.Guild(guildID); ok {
			locale = discord.Locale(guild.PreferredLocale)
		}
		trad := locales.GetModule_SafetyModule(locale)

		title := trad.Mention_spam_censored_title
		if title == "" {
			title = "Tu as été timeout pour spam de mentions."
		}
		desc := trad.Mention_spam_censored_description
		if desc == "" {
			desc = "-# S'il s'agit d'une erreur contactez le support. Code : %s"
		}

		sendCensoredMessage(client, message, string(code), title, desc)
		return true
	}

	return false
}

func (m *SafetyModule) handleEmojiSpam(b *core.Bot, client *bot.Client, message discord.Message) bool {
	if message.GuildID == nil {
		return false
	}

	guildID := *message.GuildID
	userID := message.Author.ID

	cacheKey := "emoji_spam:" + guildID.String() + ":" + userID.String()

	includeEmojis, _ := utils.IncludeEmojis(message.Content)
	if !includeEmojis {
		utils.Cache.Mu.Lock()
		delete(utils.Cache.Items, cacheKey)
		utils.Cache.Mu.Unlock()

		return false
	}

	utils.Cache.Mu.Lock()
	var userCounts []int
	if val, ok := utils.Cache.Items[cacheKey]; ok {
		if val.Expiration == 0 || time.Now().UnixNano() <= val.Expiration {
			userCounts = val.Value.([]int)
		}
	}

	userCounts = append(userCounts, 1)
	if len(userCounts) > 10 {
		userCounts = userCounts[len(userCounts)-3:]
	}

	totalEmojis := 0
	for _, count := range userCounts {
		totalEmojis += count
	}

	shouldTimeout := totalEmojis >= 10
	if shouldTimeout {
		delete(utils.Cache.Items, cacheKey)
	} else {
		utils.Cache.Items[cacheKey] = utils.CacheItem{
			Value:      userCounts,
			Expiration: time.Now().Add(5 * time.Minute).UnixNano(),
		}
	}
	utils.Cache.Mu.Unlock()

	if shouldTimeout {
		timeoutDuration := 6 * time.Hour
		until := time.Now().Add(timeoutDuration)

		go func() {
			client.Rest.UpdateMember(guildID, userID, discord.MemberUpdate{
				CommunicationDisabledUntil: omit.New(&until),
			})
		}()

		client.Rest.DeleteMessage(message.ChannelID, message.ID)
		m.giveQuarentineRole(client, *message.Member)
		code := b.Logger.SendSafetyEmojisSpamLogs(message)

		locale := discord.LocaleEnglishUS
		if guild, ok := client.Caches.Guild(guildID); ok {
			locale = discord.Locale(guild.PreferredLocale)
		}
		trad := locales.GetModule_SafetyModule(locale)

		title := trad.Emojis_spam_censored_title
		if title == "" {
			title = "Tu as été timeout pour spam d'émojis."
		}
		desc := trad.Emojis_spam_censored_description
		if desc == "" {
			desc = "-# S'il s'agit d'une erreur contactez le support. Code : %s"
		}

		sendCensoredMessage(client, message, string(code), title, desc)
		return true
	}

	return false
}

func (m *SafetyModule) handleMassAction(client *bot.Client, guildID snowflake.ID, userID snowflake.ID, count int, threshold int, cacheKey string, messageType string, arg1 string, arg2 string) {
	utils.Cache.Mu.Lock()
	var userCounts []time.Time
	if val, ok := utils.Cache.Items[cacheKey]; ok {
		if val.Expiration == 0 || time.Now().UnixNano() <= val.Expiration {
			if v, ok2 := val.Value.([]time.Time); ok2 {
				userCounts = v
			}
		}
	}

	now := time.Now()
	var newCounts []time.Time
	for _, t := range userCounts {
		if now.Sub(t) <= time.Minute {
			newCounts = append(newCounts, t)
		}
	}

	for i := 0; i < count; i++ {
		newCounts = append(newCounts, now)
	}

	shouldPunish := len(newCounts) >= threshold
	if shouldPunish {
		delete(utils.Cache.Items, cacheKey)
	} else {
		utils.Cache.Items[cacheKey] = utils.CacheItem{
			Value:      newCounts,
			Expiration: time.Now().Add(time.Minute).UnixNano(),
		}
	}
	utils.Cache.Mu.Unlock()

	if shouldPunish {
		member, err := client.Rest.GetMember(guildID, userID)
		successStr := ""
		guild, gOk := client.Caches.Guild(guildID)
		if !gOk {
			return
		}

		locale := discord.LocaleFrench
		if guild.PreferredLocale != "" {
			locale = discord.Locale(guild.PreferredLocale)
		}
		trad := locales.GetModule_SafetyModule(locale)

		if err == nil && member != nil {
			_, dperms := utils.CheckDangerousPermissions(client, *member)
			utils.RemoveMemberPerms(client, *member, dperms)

			updatedMember, uErr := client.Rest.GetMember(guildID, userID)
			if uErr == nil {
				_, newDperms := utils.CheckDangerousPermissions(client, *updatedMember)
				if len(newDperms) > 0 {
					successStr = trad.Nuke_fail
					if successStr == "" {
						successStr = "Je n'ai pas pu lui retirer ses permissions (vérifiez mon rôle)."
					}
				} else {
					successStr = trad.Nuke_success
					if successStr == "" {
						successStr = "Ses permissions dangereuses ont été retirées avec succès."
					}
				}
			}
		} else {
			successStr = trad.Nuke_fail
			if successStr == "" {
				successStr = "Je n'ai pas pu lui retirer ses permissions (vérifiez mon rôle)."
			}
		}

		ownerChannel, err := client.Rest.CreateDMChannel(guild.OwnerID)
		if err == nil {
			var text string
			switch messageType {
			case "kick":
				text = trad.Nuke_masskick_alert
				if text != "" {
					text = fmt.Sprintf(text, userID.String(), guild.Name, successStr)
				} else {
					text = fmt.Sprintf("⚠️ **Alerte Sécurité - Anti-Nuke**\nL'utilisateur <@%s> a tenté de kick en masse des membres sur votre serveur **%s**.\n%s", userID.String(), guild.Name, successStr)
				}
			case "delete":
				text = trad.Nuke_massdelete_alert
				if text != "" {
					text = fmt.Sprintf(text, userID.String(), arg1, guild.Name, successStr)
				} else {
					text = fmt.Sprintf("⚠️ **Alerte Sécurité - Anti-Nuke**\nL'utilisateur <@%s> a tenté de supprimer en masse des %s sur votre serveur **%s**.\n%s", userID.String(), arg1, guild.Name, successStr)
				}
			case "edit":
				text = trad.Nuke_massedit_alert
				if text != "" {
					text = fmt.Sprintf(text, userID.String(), arg1, guild.Name, successStr)
				} else {
					text = fmt.Sprintf("⚠️ **Alerte Sécurité - Anti-Nuke**\nL'utilisateur <@%s> a tenté de modifier en masse des %s sur votre serveur **%s**.\n%s", userID.String(), arg1, guild.Name, successStr)
				}
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(text),
					).WithAccessory(discord.NewThumbnail(*guild.IconURL())),
				).WithAccentColor(utils.ParseStrColor("#e74c3c")),
			)
			client.Rest.CreateMessage(ownerChannel.ID(), msg)
		}
	}
}

func (m *SafetyModule) HandleGuildAuditLogEntryCreate(b *core.Bot, e *events.GuildAuditLogEntryCreate) bool {
	if !m.Data.Enabled {
		return false
	}

	guildID := e.GuildID
	if e.AuditLogEntry.UserID == 0 {
		return false
	}
	userID := e.AuditLogEntry.UserID

	if userID == 0 || userID == b.Client.ID() {
		return false
	}
	guild, ok := b.Client.Caches.Guild(guildID)
	if ok && userID == guild.OwnerID {
		return false
	}

	switch e.AuditLogEntry.ActionType {
	case discord.AuditLogEventMemberKick:
		if m.Data.AntiNuke.AntiMassKick {
			cacheKey := "masskick:" + guildID.String() + ":" + userID.String()
			m.handleMassAction(b.Client, guildID, userID, 1, 4, cacheKey, "kick", "", "")
		}

	case discord.AuditLogEventChannelDelete:
		if m.Data.AntiNuke.AntiMassChannelD {
			cacheKey := "masschanneldelete:" + guildID.String() + ":" + userID.String()
			m.handleMassAction(b.Client, guildID, userID, 1, 2, cacheKey, "delete", "salons", "")
		}

	case discord.AuditLogEventRoleDelete:
		if m.Data.AntiNuke.AntiMassRoleD {
			cacheKey := "massroledelete:" + guildID.String() + ":" + userID.String()
			m.handleMassAction(b.Client, guildID, userID, 1, 2, cacheKey, "delete", "rôles", "")
		}

	case discord.AuditLogEventChannelUpdate:
		if m.Data.AntiNuke.AntiMassChannelD {
			ignore := true
			for _, change := range e.AuditLogEntry.Changes {
				if change.Key != discord.AuditLogChangeKeyPosition {
					ignore = false
					break
				}
			}
			if !ignore {
				cacheKey := "masschanneledit:" + guildID.String() + ":" + userID.String()
				m.handleMassAction(b.Client, guildID, userID, 1, 5, cacheKey, "edit", "salons", "")
			}
		}

	case discord.AuditLogEventRoleUpdate:
		if m.Data.AntiNuke.AntiMassRoleD {
			ignore := true
			for _, change := range e.AuditLogEntry.Changes {
				if change.Key != discord.AuditLogChangeKeyPosition {
					ignore = false
					break
				}
			}
			if !ignore {
				cacheKey := "massroleedit:" + guildID.String() + ":" + userID.String()
				m.handleMassAction(b.Client, guildID, userID, 1, 5, cacheKey, "edit", "rôles", "")
			}
		}
	}

	return false
}
