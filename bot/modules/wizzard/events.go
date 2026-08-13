package wizzard

import (
	"strings"

	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *WizzardModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if !m.Data.Enabled || e.Message.Author.Bot {
		return false
	}

	prefix := m.Data.Main.Prefix
	if prefix == "" {
		prefix = "!"
	}

	if strings.HasPrefix(e.Message.Content, prefix) {
		content := strings.TrimPrefix(e.Message.Content, prefix)
		args := strings.Fields(content)
		if len(args) == 0 {
			return false
		}

		spellName := strings.ToLower(args[0])

		guild, ok := e.Client().Caches.Guild(*e.GuildID)
		if !ok {
			return false
		}

		locale := discord.Locale(guild.PreferredLocale)
		trad := locales.GetModule_WizzardModule(locale)
		sorts := InitWizzardSpells(trad)

		tradEn := locales.GetModule_WizzardModule(discord.LocaleEnglishUS)
		sortsEn := InitWizzardSpells(tradEn)

		for i, sort := range sorts {
			sortEn := sortsEn[i]
			cleanName := strings.ToLower(strings.ReplaceAll(sort.Name, " ", ""))
			cleanNameEn := strings.ToLower(strings.ReplaceAll(sortEn.Name, " ", ""))

			if spellName == sort.Key || spellName == cleanName || spellName == cleanNameEn {

				var hasPerms bool
				if e.Message.Member != nil {
					memberPerms := b.Client.Caches.MemberPermissions(*e.Message.Member)
					hasPerms = memberPerms.Has(sort.Permissions...)
				}

				me, _ := e.Client().Caches.Member(*e.GuildID, e.Client().ID())
				bHasPerms := b.Client.Caches.MemberPermissions(me).Has(sort.Permissions...)

				if (!hasPerms || !bHasPerms) && len(sort.Permissions) > 0 {
					msg := createSpellMessage("#e74c3c", "Échec du sort", m.Data.Translations.NoPermission)
					b.SendMessage(e.ChannelID.String(), msg)
					return true
				}

				for _, req := range sort.RequiredArgs {
					if req == "mentionUser" && len(e.Message.Mentions) == 0 {
						msg := createSpellMessage("#e4ab17", "Arguments manquants", m.Data.Translations.MissingArgs)
						b.SendMessage(e.ChannelID.String(), msg)
						return true
					}
					if req == "mentionChannel" && len(e.Message.MentionChannels) == 0 {
						msg := createSpellMessage("#e4ab17", "Arguments manquants", m.Data.Translations.MissingArgs)
						b.SendMessage(e.ChannelID.String(), msg)
						return true
					}
				}

				sort.Execute(b, e, args[1:], trad)
				return true
			}
		}
	}

	return false
}
