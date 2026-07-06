package utils

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"
	"onyx/bot/utils"
	"sort"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "translate",
		Description: "Translate a text.",
		Category:    "Utils",
		Create: discord.SlashCommandCreate{
			Name:        "translate",
			Description: "Translate a text.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "text",
					Description: "The text to translate",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "lang",
					Description: "The targeted lang",
					Required:    true,
					Choices:     translateLangChoices(),
				},
			},
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			cmd := event.SlashCommandInteractionData()
			locale := locales.GetTranslate(event.Locale())

			text, _ := cmd.OptString("text")
			lang, _ := cmd.OptString("lang")

			if len(text) > 2000 {
				text = fmt.Sprintf("%s...", text[0:1996])
			}

			t := utils.Translate(utils.TranslateParams{
				Query:  text,
				Source: "auto",
				Target: lang,
			})

			translatedText := t.TranslatedText
			if len(translatedText) > 2000 {
				translatedText = fmt.Sprintf("%s...", translatedText[0:1996])
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay(translatedText),
					discord.NewTextDisplayf(locale.Footer, utils.TranslateLangs[lang].Flag, utils.TranslateLangs[lang].Name),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			id := strings.Split(event.Data.CustomID(), "-")[2]
			switch id {
			case "ephemeral":
				lang := discordLocaleToTranslateLang(event.Locale())

				text := event.Message.Components[0].(discord.ContainerComponent).Components[0].(discord.TextDisplayComponent).Content

				if len(text) > 2000 {
					text = fmt.Sprintf("%s...", text[0:1996])
				}

				t := utils.Translate(utils.TranslateParams{
					Query:  text,
					Source: "auto",
					Target: lang,
				})

				trad := t.TranslatedText
				if len(trad) > 2000 {
					trad = fmt.Sprintf("%s...", trad[0:1996])
				}

				msg := discord.NewMessageCreateV2(
					discord.NewContainer(
						discord.NewTextDisplay(trad),
						discord.NewTextDisplayf("-# %s %s Translation", utils.TranslateLangs[lang].Flag, utils.TranslateLangs[lang].Name),
					),
				).WithFlags(discord.MessageFlagEphemeral)

				if err := event.CreateMessage(msg); err != nil {

				}
			}
		},
	})
}

func discordLocaleToTranslateLang(locale discord.Locale) string {
	switch locale {
	case discord.LocaleChineseCN, discord.LocaleChineseTW:
		return "zh"
	case discord.LocaleEnglishUS, discord.LocaleEnglishGB:
		return "en"
	case discord.LocaleFrench:
		return "fr"
	case discord.LocaleGerman:
		return "de"
	case discord.LocaleItalian:
		return "it"
	case discord.LocaleJapanese:
		return "ja"
	case discord.LocalePortugueseBR:
		return "pt"
	case discord.LocaleRussian:
		return "ru"
	case discord.LocaleSpanishES, discord.LocaleSpanishLATAM:
		return "es"
	default:
		return "en"
	}
}

func translateLangChoices() []discord.ApplicationCommandOptionChoiceString {
	keys := make([]string, 0, len(utils.TranslateLangs))
	for value := range utils.TranslateLangs {
		keys = append(keys, value)
	}
	sort.Strings(keys)

	choices := make([]discord.ApplicationCommandOptionChoiceString, 0, len(keys))
	for _, value := range keys {
		choices = append(choices, discord.ApplicationCommandOptionChoiceString{
			Name:  string(utils.TranslateLangs[value].Name),
			Value: value,
		})
	}

	return choices
}
