package modules

import (
	"onyx/bot/core"
	"onyx/bot/locales"
	"onyx/bot/utils"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
	"gorm.io/gorm"
)

type WizzardMainSettings struct {
	Prefix string `json:"prefix"`
}

type WizzardTranslations struct {
	NoPermission string `json:"no_permission"`
	MissingArgs  string `json:"missing_args"`
	SpellSuccess string `json:"spell_success"`
	SpellFailed  string `json:"spell_failed"`
}

type WizzardSettings struct {
	GuildID      string              `gorm:"primaryKey" json:"guild_id"`
	Enabled      bool                `gorm:"default:false" json:"enabled"`
	Main         WizzardMainSettings `gorm:"embedded;embeddedPrefix:main_" json:"main"`
	Translations WizzardTranslations `gorm:"embedded;embeddedPrefix:translations_" json:"translations"`
}

type WizzardModule struct {
	Data WizzardSettings
}

type Sort struct {
	Key          string
	Name         string
	Execute      func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad)
	Permissions  []discord.Permissions
	Description  string
	RequiredArgs []string
}

func init() {
	Register(&WizzardModule{})
}

func (m *WizzardModule) Metadata() core.Metadata {
	return core.Metadata{
		Name: "WizzardModule",
		Icon: "wand_shine",
		Label: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_WizzardModule").Label
		},
		Description: func(locale discord.Locale) string {
			return locales.GetMeta(locale, "module_WizzardModule").Description
		},
		Submodules: func(locale discord.Locale) map[string]core.SubmoduleMeta {
			meta := locales.GetMeta(locale, "module_WizzardModule")
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

func (m *WizzardModule) Priority() int   { return 1 }
func (m *WizzardModule) IsEnabled() bool { return m.Data.Enabled }
func (m *WizzardModule) Permissions() []discord.Permissions {
	return []discord.Permissions{
		discord.PermissionSendMessages,
	}
}

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

func createSpellMessage(color string, title string, description string) discord.MessageCreate {
	builder := discord.NewContainer(
		discord.NewTextDisplayf("## %s", title),
	).WithAccentColor(utils.ParseStrColor(color))

	if description != "" {
		builder = builder.AddComponents(discord.NewTextDisplay(description))
	}

	return discord.NewMessageCreateV2(
		builder,
	)
}

func formatSpellError(failMessage string, err error) string {
	if err == nil {
		return failMessage
	}
	var errStr string
	if strings.Contains(err.Error(), "50013") {
		errStr = "Je n'ai pas la permission d'effectuer cette action (hiérarchie des rôles ou permission manquante)."
	} else {
		errStr = err.Error()
	}
	return failMessage + "\n\n*Erreur : " + errStr + "*"
}

func (m *WizzardModule) Schema() interface{}  { return &WizzardSettings{} }
func (m *WizzardModule) DataPtr() interface{} { return &m.Data }
func (m *WizzardModule) LoadData(db *gorm.DB, guildID string) error {
	m.Data = WizzardSettings{GuildID: guildID}
	return db.FirstOrCreate(&m.Data, WizzardSettings{GuildID: guildID}).Error
}

func (m *WizzardModule) UISchema(locale discord.Locale) core.UISchema {
	return core.UISchema{
		SubModules: []core.UISubModule{
			{
				Name:        "main",
				Label:       "Paramètres Principaux",
				Description: "Configuration générale des sorts.",
				Components: []core.UIComponent{
					{
						Name:        "prefix",
						Label:       "Préfixe des sorts",
						Description: "Le préfixe à utiliser avant le nom du sort (ex: !avadakedavra)",
						Placeholder: "!",
						Type:        core.ComponentTypeString,
						Required:    true,
						Max:         func() *int { v := 5; return &v }(),
						Min:         func() *int { v := 1; return &v }(),
					},
				},
			},
			{
				Name:        "translations",
				Label:       "Traductions",
				Description: "Personnaliser les messages de réponse.",
				Components: []core.UIComponent{
					{
						Name:        "no_permission",
						Label:       "Message - Pas de permission",
						Description: "Envoyé quand l'utilisateur ou le bot n'a pas la permission.",
						Placeholder: "Vous n'avez pas la permission de lancer ce sort.",
						Type:        core.ComponentTypeString,
						Required:    false,
					},
					{
						Name:        "missing_args",
						Label:       "Message - Arguments manquants",
						Description: "Envoyé quand il manque des arguments (ex: mention).",
						Placeholder: "Il manque des arguments pour ce sort.",
						Type:        core.ComponentTypeString,
						Required:    false,
					},
					{
						Name:        "spell_success",
						Label:       "Message - Sort réussi",
						Description: "Envoyé quand un sort est réussi.",
						Placeholder: "Le sort a été lancé avec succès !",
						Type:        core.ComponentTypeString,
						Required:    false,
					},
					{
						Name:        "spell_failed",
						Label:       "Message - Sort échoué",
						Description: "Envoyé quand un sort échoue.",
						Placeholder: "Le sort a échoué...",
						Type:        core.ComponentTypeString,
						Required:    false,
					},
				},
			},
		},
	}
}

func InitWizzardSpells(t locales.ModuleWizzardModuleTrad) []Sort {
	sp := t.Spells
	var sorts []Sort

	sorts = append(sorts, Sort{
		Key:          "aberto",
		Name:         sp.Aberto.Name,
		Description:  sp.Aberto.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Aberto.Name, sp.Aberto.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "accio",
		Name:         sp.Accio.Name,
		Description:  sp.Accio.Description,
		Permissions:  []discord.Permissions{discord.PermissionMoveMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Accio.Name, sp.Accio.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				if vs, ok := e.Client().Caches.VoiceState(*e.GuildID, e.Message.Author.ID); ok && vs.ChannelID != nil {
					channelID := *vs.ChannelID
					update := discord.MemberUpdate{ChannelID: &channelID}
					_, err := b.Client.Rest.UpdateMember(*e.GuildID, user.ID, update)
					if err == nil {
						msg = createSpellMessage("#3498db", sp.Accio.Name, "🧲 "+user.Tag()+" a été attiré vers vous !")
					} else {
						msg = createSpellMessage("#e74c3c", sp.Accio.Name, formatSpellError(sp.Accio.Fail, err))
					}
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "aguamenti",
		Name:         sp.Aguamenti.Name,
		Description:  sp.Aguamenti.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Aguamenti.Name, sp.Aguamenti.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "alohomora",
		Name:         sp.Alohomora.Name,
		Description:  sp.Alohomora.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Alohomora.Name, sp.Alohomora.Fail)
			_, err := b.Client.Rest.UpdateChannel(e.ChannelID, discord.GuildTextChannelUpdate{})
			if err == nil {
				msg = createSpellMessage("#f1c40f", sp.Alohomora.Name, "🔓 Ce salon a été déverrouillé !")
			} else {
				msg = createSpellMessage("#e74c3c", sp.Alohomora.Name, formatSpellError(sp.Alohomora.Fail, err))
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "amplificatum",
		Name:         sp.Amplificatum.Name,
		Description:  sp.Amplificatum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Amplificatum.Name, sp.Amplificatum.Fail)
			if len(args) > 0 {
				text := strings.Join(args, " ")
				msg = createSpellMessage("#2ecc71", sp.Amplificatum.Name, "# "+text)
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "anapneo",
		Name:         sp.Anapneo.Name,
		Description:  sp.Anapneo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Anapneo.Name, sp.Anapneo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "aparecium",
		Name:         sp.Aparecium.Name,
		Description:  sp.Aparecium.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Aparecium.Name, sp.Aparecium.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "apparevestigium",
		Name:         sp.Apparevestigium.Name,
		Description:  sp.Apparevestigium.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Apparevestigium.Name, sp.Apparevestigium.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "araniaexumai",
		Name:         sp.Araniaexumai.Name,
		Description:  sp.Araniaexumai.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Araniaexumai.Name, sp.Araniaexumai.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "arrestomomentum",
		Name:         sp.Arrestomomentum.Name,
		Description:  sp.Arrestomomentum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Arrestomomentum.Name, sp.Arrestomomentum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "ascendio",
		Name:         sp.Ascendio.Name,
		Description:  sp.Ascendio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Ascendio.Name, sp.Ascendio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "assurdiato",
		Name:         sp.Assurdiato.Name,
		Description:  sp.Assurdiato.Description,
		Permissions:  []discord.Permissions{discord.PermissionDeafenMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Assurdiato.Name, sp.Assurdiato.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				deaf := true
				update := discord.MemberUpdate{Deaf: &deaf}
				_, err := b.Client.Rest.UpdateMember(*e.GuildID, user.ID, update)
				if err == nil {
					msg = createSpellMessage("#9b59b6", sp.Assurdiato.Name, "🙉 "+user.Tag()+" a été assourdi !")
				} else {
					msg = createSpellMessage("#e74c3c", sp.Assurdiato.Name, formatSpellError(sp.Assurdiato.Fail, err))
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "avadakedavra",
		Name:         sp.Avadakedavra.Name,
		Description:  sp.Avadakedavra.Description,
		Permissions:  []discord.Permissions{discord.PermissionBanMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Avadakedavra.Name, sp.Avadakedavra.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				err := b.Client.Rest.AddBan(*e.GuildID, user.ID, 0)
				if err == nil {
					msg = createSpellMessage("#e74c3c", sp.Avadakedavra.Name, "☠️ "+user.Tag()+" a été frappé par le sortilège de la mort !")
				} else {
					msg = createSpellMessage("#e74c3c", sp.Avadakedavra.Name, formatSpellError(sp.Avadakedavra.Fail, err))
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "avensegium",
		Name:         sp.Avensegium.Name,
		Description:  sp.Avensegium.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Avensegium.Name, sp.Avensegium.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "avis",
		Name:         sp.Avis.Name,
		Description:  sp.Avis.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Avis.Name, sp.Avis.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "bloclang",
		Name:         sp.Bloclang.Name,
		Description:  sp.Bloclang.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Bloclang.Name, sp.Bloclang.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "bombardamaxima",
		Name:         sp.Bombardamaxima.Name,
		Description:  sp.Bombardamaxima.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Bombardamaxima.Name, sp.Bombardamaxima.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "brachialigo",
		Name:         sp.Brachialigo.Name,
		Description:  sp.Brachialigo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Brachialigo.Name, sp.Brachialigo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "brachiumemendo",
		Name:         sp.Brachiumemendo.Name,
		Description:  sp.Brachiumemendo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Brachiumemendo.Name, sp.Brachiumemendo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "caveinimicum",
		Name:         sp.Caveinimicum.Name,
		Description:  sp.Caveinimicum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Caveinimicum.Name, sp.Caveinimicum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "circumrota",
		Name:         sp.Circumrota.Name,
		Description:  sp.Circumrota.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Circumrota.Name, sp.Circumrota.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "capaciousextremis",
		Name:         sp.Capaciousextremis.Name,
		Description:  sp.Capaciousextremis.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Capaciousextremis.Name, sp.Capaciousextremis.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "cistemaperio",
		Name:         sp.Cistemaperio.Name,
		Description:  sp.Cistemaperio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Cistemaperio.Name, sp.Cistemaperio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "collaporta",
		Name:         sp.Collaporta.Name,
		Description:  sp.Collaporta.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Collaporta.Name, sp.Collaporta.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "confringo",
		Name:         sp.Confringo.Name,
		Description:  sp.Confringo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Confringo.Name, sp.Confringo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "confundo",
		Name:         sp.Confundo.Name,
		Description:  sp.Confundo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Confundo.Name, sp.Confundo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "slugvomitingcharm",
		Name:         sp.Slugvomitingcharm.Name,
		Description:  sp.Slugvomitingcharm.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Slugvomitingcharm.Name, sp.Slugvomitingcharm.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "defodio",
		Name:         sp.Defodio.Name,
		Description:  sp.Defodio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Defodio.Name, sp.Defodio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "deprimo",
		Name:         sp.Deprimo.Name,
		Description:  sp.Deprimo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Deprimo.Name, sp.Deprimo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "dentesaugmento",
		Name:         sp.Dentesaugmento.Name,
		Description:  sp.Dentesaugmento.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Dentesaugmento.Name, sp.Dentesaugmento.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "destructum",
		Name:         sp.Destructum.Name,
		Description:  sp.Destructum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Destructum.Name, sp.Destructum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "diffindo",
		Name:         sp.Diffindo.Name,
		Description:  sp.Diffindo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Diffindo.Name, sp.Diffindo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "dissendium",
		Name:         sp.Dissendium.Name,
		Description:  sp.Dissendium.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Dissendium.Name, sp.Dissendium.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "duro",
		Name:         sp.Duro.Name,
		Description:  sp.Duro.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Duro.Name, sp.Duro.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "emancipare",
		Name:         sp.Emancipare.Name,
		Description:  sp.Emancipare.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Emancipare.Name, sp.Emancipare.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "endoloris",
		Name:         sp.Endoloris.Name,
		Description:  sp.Endoloris.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Endoloris.Name, sp.Endoloris.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "enervatum",
		Name:         sp.Enervatum.Name,
		Description:  sp.Enervatum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Enervatum.Name, sp.Enervatum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "episkey",
		Name:         sp.Episkey.Name,
		Description:  sp.Episkey.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Episkey.Name, sp.Episkey.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "erigo",
		Name:         sp.Erigo.Name,
		Description:  sp.Erigo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Erigo.Name, sp.Erigo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "evanesco",
		Name:         sp.Evanesco.Name,
		Description:  sp.Evanesco.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Evanesco.Name, sp.Evanesco.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "evertestatum",
		Name:         sp.Evertestatum.Name,
		Description:  sp.Evertestatum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Evertestatum.Name, sp.Evertestatum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "expectopatronum",
		Name:         sp.Expectopatronum.Name,
		Description:  sp.Expectopatronum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Expectopatronum.Name, sp.Expectopatronum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "expelliarmus",
		Name:         sp.Expelliarmus.Name,
		Description:  sp.Expelliarmus.Description,
		Permissions:  []discord.Permissions{discord.PermissionModerateMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Expelliarmus.Name, sp.Expelliarmus.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				until := time.Now().Add(5 * time.Minute)
				update := discord.MemberUpdate{CommunicationDisabledUntil: omit.New(&until)}
				_, err := b.Client.Rest.UpdateMember(*e.GuildID, user.ID, update)
				if err == nil {
					msg = createSpellMessage("#e67e22", sp.Expelliarmus.Name, "🛡️ "+user.Tag()+" a été désarmé (timeout de 5 minutes) !")
				} else {
					msg = createSpellMessage("#e74c3c", sp.Expelliarmus.Name, formatSpellError(sp.Expelliarmus.Fail, err))
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "expulso",
		Name:         sp.Expulso.Name,
		Description:  sp.Expulso.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Expulso.Name, sp.Expulso.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "fenestra",
		Name:         sp.Fenestra.Name,
		Description:  sp.Fenestra.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Fenestra.Name, sp.Fenestra.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "failamalle",
		Name:         sp.Failamalle.Name,
		Description:  sp.Failamalle.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Failamalle.Name, sp.Failamalle.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "ferula",
		Name:         sp.Ferula.Name,
		Description:  sp.Ferula.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Ferula.Name, sp.Ferula.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "fiantoduri",
		Name:         sp.Fiantoduri.Name,
		Description:  sp.Fiantoduri.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Fiantoduri.Name, sp.Fiantoduri.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "fidelitas",
		Name:         sp.Fidelitas.Name,
		Description:  sp.Fidelitas.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Fidelitas.Name, sp.Fidelitas.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "finiteincantatem",
		Name:         sp.Finiteincantatem.Name,
		Description:  sp.Finiteincantatem.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Finiteincantatem.Name, sp.Finiteincantatem.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "flambios",
		Name:         sp.Flambios.Name,
		Description:  sp.Flambios.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Flambios.Name, sp.Flambios.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "fulgari",
		Name:         sp.Fulgari.Name,
		Description:  sp.Fulgari.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Fulgari.Name, sp.Fulgari.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "furunculus",
		Name:         sp.Furunculus.Name,
		Description:  sp.Furunculus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Furunculus.Name, sp.Furunculus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "gemino",
		Name:         sp.Gemino.Name,
		Description:  sp.Gemino.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Gemino.Name, sp.Gemino.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "glisseo",
		Name:         sp.Glisseo.Name,
		Description:  sp.Glisseo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Glisseo.Name, sp.Glisseo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "harmonianecterepassus",
		Name:         sp.Harmonianecterepassus.Name,
		Description:  sp.Harmonianecterepassus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Harmonianecterepassus.Name, sp.Harmonianecterepassus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "hominumrevelio",
		Name:         sp.Hominumrevelio.Name,
		Description:  sp.Hominumrevelio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Hominumrevelio.Name, sp.Hominumrevelio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "homomorphus",
		Name:         sp.Homomorphus.Name,
		Description:  sp.Homomorphus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Homomorphus.Name, sp.Homomorphus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "impedimenta",
		Name:         sp.Impedimenta.Name,
		Description:  sp.Impedimenta.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Impedimenta.Name, sp.Impedimenta.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "impero",
		Name:         sp.Impero.Name,
		Description:  sp.Impero.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Impero.Name, sp.Impero.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "immobulus",
		Name:         sp.Immobulus.Name,
		Description:  sp.Immobulus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Immobulus.Name, sp.Immobulus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "impervius",
		Name:         sp.Impervius.Name,
		Description:  sp.Impervius.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Impervius.Name, sp.Impervius.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "incarcerem",
		Name:         sp.Incarcerem.Name,
		Description:  sp.Incarcerem.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Incarcerem.Name, sp.Incarcerem.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "incendio",
		Name:         sp.Incendio.Name,
		Description:  sp.Incendio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Incendio.Name, sp.Incendio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "lacarnuminflammari",
		Name:         sp.Lacarnuminflammari.Name,
		Description:  sp.Lacarnuminflammari.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Lacarnuminflammari.Name, sp.Lacarnuminflammari.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "legilimens",
		Name:         sp.Legilimens.Name,
		Description:  sp.Legilimens.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Legilimens.Name, sp.Legilimens.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "levicorpus",
		Name:         sp.Levicorpus.Name,
		Description:  sp.Levicorpus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Levicorpus.Name, sp.Levicorpus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "liberacorpus",
		Name:         sp.Liberacorpus.Name,
		Description:  sp.Liberacorpus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Liberacorpus.Name, sp.Liberacorpus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "locomotorbarda",
		Name:         sp.Locomotorbarda.Name,
		Description:  sp.Locomotorbarda.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Locomotorbarda.Name, sp.Locomotorbarda.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "locomotormortis",
		Name:         sp.Locomotormortis.Name,
		Description:  sp.Locomotormortis.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Locomotormortis.Name, sp.Locomotormortis.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "locomotorwibbly",
		Name:         sp.Locomotorwibbly.Name,
		Description:  sp.Locomotorwibbly.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Locomotorwibbly.Name, sp.Locomotorwibbly.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "meteorribilisrecanto",
		Name:         sp.Meteorribilisrecanto.Name,
		Description:  sp.Meteorribilisrecanto.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Meteorribilisrecanto.Name, sp.Meteorribilisrecanto.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "mobiliarbus",
		Name:         sp.Mobiliarbus.Name,
		Description:  sp.Mobiliarbus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Mobiliarbus.Name, sp.Mobiliarbus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "mobilicorpus",
		Name:         sp.Mobilicorpus.Name,
		Description:  sp.Mobilicorpus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Mobilicorpus.Name, sp.Mobilicorpus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "molliare",
		Name:         sp.Molliare.Name,
		Description:  sp.Molliare.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Molliare.Name, sp.Molliare.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "morsmordre",
		Name:         sp.Morsmordre.Name,
		Description:  sp.Morsmordre.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Morsmordre.Name, sp.Morsmordre.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "mutinlutinmalinpesti",
		Name:         sp.Mutinlutinmalinpesti.Name,
		Description:  sp.Mutinlutinmalinpesti.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Mutinlutinmalinpesti.Name, sp.Mutinlutinmalinpesti.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "nebulus",
		Name:         sp.Nebulus.Name,
		Description:  sp.Nebulus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Nebulus.Name, sp.Nebulus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "nox",
		Name:         sp.Nox.Name,
		Description:  sp.Nox.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Nox.Name, sp.Nox.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "obscuro",
		Name:         sp.Obscuro.Name,
		Description:  sp.Obscuro.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Obscuro.Name, sp.Obscuro.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "oppugno",
		Name:         sp.Oppugno.Name,
		Description:  sp.Oppugno.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Oppugno.Name, sp.Oppugno.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "orchideus",
		Name:         sp.Orchideus.Name,
		Description:  sp.Orchideus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Orchideus.Name, sp.Orchideus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "oscausi",
		Name:         sp.Oscausi.Name,
		Description:  sp.Oscausi.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Oscausi.Name, sp.Oscausi.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "oubliettes",
		Name:         sp.Oubliettes.Name,
		Description:  sp.Oubliettes.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Oubliettes.Name, sp.Oubliettes.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "partistemporus",
		Name:         sp.Partistemporus.Name,
		Description:  sp.Partistemporus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Partistemporus.Name, sp.Partistemporus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "periculum",
		Name:         sp.Periculum.Name,
		Description:  sp.Periculum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Periculum.Name, sp.Periculum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "petrificustotalus",
		Name:         sp.Petrificustotalus.Name,
		Description:  sp.Petrificustotalus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Petrificustotalus.Name, sp.Petrificustotalus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "piertotumlocomotor",
		Name:         sp.Piertotumlocomotor.Name,
		Description:  sp.Piertotumlocomotor.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Piertotumlocomotor.Name, sp.Piertotumlocomotor.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "portus",
		Name:         sp.Portus.Name,
		Description:  sp.Portus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Portus.Name, sp.Portus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "priorincanto",
		Name:         sp.Priorincanto.Name,
		Description:  sp.Priorincanto.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Priorincanto.Name, sp.Priorincanto.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "prioriincantatum",
		Name:         sp.Prioriincantatum.Name,
		Description:  sp.Prioriincantatum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Prioriincantatum.Name, sp.Prioriincantatum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "protego",
		Name:         sp.Protego.Name,
		Description:  sp.Protego.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Protego.Name, sp.Protego.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "protegodiabolica",
		Name:         sp.Protegodiabolica.Name,
		Description:  sp.Protegodiabolica.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Protegodiabolica.Name, sp.Protegodiabolica.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "protegototalum",
		Name:         sp.Protegototalum.Name,
		Description:  sp.Protegototalum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Protegototalum.Name, sp.Protegototalum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "protegohorribilis",
		Name:         sp.Protegohorribilis.Name,
		Description:  sp.Protegohorribilis.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Protegohorribilis.Name, sp.Protegohorribilis.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "protegomaxima",
		Name:         sp.Protegomaxima.Name,
		Description:  sp.Protegomaxima.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Protegomaxima.Name, sp.Protegomaxima.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "recurvite",
		Name:         sp.Recurvite.Name,
		Description:  sp.Recurvite.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Recurvite.Name, sp.Recurvite.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "reducto",
		Name:         sp.Reducto.Name,
		Description:  sp.Reducto.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Reducto.Name, sp.Reducto.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "reparo",
		Name:         sp.Reparo.Name,
		Description:  sp.Reparo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Reparo.Name, sp.Reparo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "repellomoldum",
		Name:         sp.Repellomoldum.Name,
		Description:  sp.Repellomoldum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Repellomoldum.Name, sp.Repellomoldum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "repelloinimicium",
		Name:         sp.Repelloinimicium.Name,
		Description:  sp.Repelloinimicium.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Repelloinimicium.Name, sp.Repelloinimicium.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "reverte",
		Name:         sp.Reverte.Name,
		Description:  sp.Reverte.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Reverte.Name, sp.Reverte.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "revigor",
		Name:         sp.Revigor.Name,
		Description:  sp.Revigor.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Revigor.Name, sp.Revigor.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "revelio",
		Name:         sp.Revelio.Name,
		Description:  sp.Revelio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Revelio.Name, sp.Revelio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "rictusempra",
		Name:         sp.Rictusempra.Name,
		Description:  sp.Rictusempra.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Rictusempra.Name, sp.Rictusempra.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "riddikulus",
		Name:         sp.Riddikulus.Name,
		Description:  sp.Riddikulus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Riddikulus.Name, sp.Riddikulus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "salveomaleficia",
		Name:         sp.Salveomaleficia.Name,
		Description:  sp.Salveomaleficia.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Salveomaleficia.Name, sp.Salveomaleficia.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "sectumsempra",
		Name:         sp.Sectumsempra.Name,
		Description:  sp.Sectumsempra.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Sectumsempra.Name, sp.Sectumsempra.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "serpensortia",
		Name:         sp.Serpensortia.Name,
		Description:  sp.Serpensortia.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Serpensortia.Name, sp.Serpensortia.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "silencio",
		Name:         sp.Silencio.Name,
		Description:  sp.Silencio.Description,
		Permissions:  []discord.Permissions{discord.PermissionMuteMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Silencio.Name, sp.Silencio.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				mute := true
				update := discord.MemberUpdate{Mute: &mute}
				_, err := b.Client.Rest.UpdateMember(*e.GuildID, user.ID, update)
				if err == nil {
					msg = createSpellMessage("#9b59b6", sp.Silencio.Name, "🤫 "+user.Tag()+" a été rendu muet !")
				} else {
					msg = createSpellMessage("#e74c3c", sp.Silencio.Name, formatSpellError(sp.Silencio.Fail, err))
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "sonorus",
		Name:         sp.Sonorus.Name,
		Description:  sp.Sonorus.Description,
		Permissions:  []discord.Permissions{discord.PermissionMuteMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Sonorus.Name, sp.Sonorus.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				mute := false
				update := discord.MemberUpdate{Mute: &mute}
				_, err := b.Client.Rest.UpdateMember(*e.GuildID, user.ID, update)
				if err == nil {
					msg = createSpellMessage("#f1c40f", sp.Sonorus.Name, "🔊 "+user.Tag()+" peut à nouveau parler !")
				} else {
					msg = createSpellMessage("#e74c3c", sp.Sonorus.Name, formatSpellError(sp.Sonorus.Fail, err))
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "surdinam",
		Name:         sp.Surdinam.Name,
		Description:  sp.Surdinam.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Surdinam.Name, sp.Surdinam.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "surgito",
		Name:         sp.Surgito.Name,
		Description:  sp.Surgito.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Surgito.Name, sp.Surgito.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "specialisrevelio",
		Name:         sp.Specialisrevelio.Name,
		Description:  sp.Specialisrevelio.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Specialisrevelio.Name, sp.Specialisrevelio.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "speropatronum",
		Name:         sp.Speropatronum.Name,
		Description:  sp.Speropatronum.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Speropatronum.Name, sp.Speropatronum.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "stupefix",
		Name:         sp.Stupefix.Name,
		Description:  sp.Stupefix.Description,
		Permissions:  []discord.Permissions{discord.PermissionKickMembers},
		RequiredArgs: []string{"mentionUser"},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Stupefix.Name, sp.Stupefix.Fail)
			if len(e.Message.Mentions) > 0 {
				user := e.Message.Mentions[0]
				err := b.Client.Rest.RemoveMember(*e.GuildID, user.ID)
				if err == nil {
					msg = createSpellMessage("#e74c3c", sp.Stupefix.Name, "💥 "+user.Tag()+" a été expulsé par Stupéfix !")
				} else {
					msg = createSpellMessage("#e74c3c", sp.Stupefix.Name, formatSpellError(sp.Stupefix.Fail, err))
				}
			}
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "tarentallegra",
		Name:         sp.Tarentallegra.Name,
		Description:  sp.Tarentallegra.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Tarentallegra.Name, sp.Tarentallegra.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "tergeo",
		Name:         sp.Tergeo.Name,
		Description:  sp.Tergeo.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Tergeo.Name, sp.Tergeo.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "ventus",
		Name:         sp.Ventus.Name,
		Description:  sp.Ventus.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Ventus.Name, sp.Ventus.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "veraverto",
		Name:         sp.Veraverto.Name,
		Description:  sp.Veraverto.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Veraverto.Name, sp.Veraverto.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "viperaevanesca",
		Name:         sp.Viperaevanesca.Name,
		Description:  sp.Viperaevanesca.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Viperaevanesca.Name, sp.Viperaevanesca.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "volateascendere",
		Name:         sp.Volateascendere.Name,
		Description:  sp.Volateascendere.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Volateascendere.Name, sp.Volateascendere.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "vulnerasanentur",
		Name:         sp.Vulnerasanentur.Name,
		Description:  sp.Vulnerasanentur.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Vulnerasanentur.Name, sp.Vulnerasanentur.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "waddiwasi",
		Name:         sp.Waddiwasi.Name,
		Description:  sp.Waddiwasi.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Waddiwasi.Name, sp.Waddiwasi.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	sorts = append(sorts, Sort{
		Key:          "wingardiumleviosa",
		Name:         sp.Wingardiumleviosa.Name,
		Description:  sp.Wingardiumleviosa.Description,
		Permissions:  []discord.Permissions{},
		RequiredArgs: []string{},
		Execute: func(b *core.Bot, e *events.MessageCreate, args []string, t locales.ModuleWizzardModuleTrad) {
			msg := createSpellMessage("#e74c3c", sp.Wingardiumleviosa.Name, sp.Wingardiumleviosa.Fail)
			b.SendMessage(e.ChannelID.String(), msg)
		},
	})

	return sorts
}
