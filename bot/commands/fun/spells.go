package fun

import (
	"fmt"
	"strings"

	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"
	"onyx/bot/modules/wizzard"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "spells",
		Description: "Get spells informations",
		Category:    "Utils",
		Create: discord.SlashCommandCreate{
			Name:        "spells",
			Description: "Get spells information",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "spell",
					Description:  "The spell",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			searchOpt, _ := event.Data.Option("spell")
			search := strings.ToLower(fmt.Sprint(searchOpt))
			spells := wizzard.InitWizzardSpells(locales.GetModule_WizzardModule(event.Locale()))
			choices := make([]discord.AutocompleteChoice, 0, 25)

			for _, s := range spells {
				name := strings.ToLower(s.Name)
				if search != "" && !strings.Contains(name, search) {
					continue
				}

				choices = append(choices, discord.AutocompleteChoiceString{
					Name:  s.Name,
					Value: s.Key,
				})
				if len(choices) == 25 {
					break
				}
			}

			_ = event.Respond(discord.InteractionResponseTypeAutocompleteResult, discord.AutocompleteResult{
				Choices: choices,
			})
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			cmd := event.SlashCommandInteractionData()
			s, _ := cmd.Option("spell")
			spells := wizzard.InitWizzardSpells(locales.GetModule_WizzardModule(event.Locale()))

			var spell wizzard.Sort
			for _, sort := range spells {
				if sort.Key == s.String() {
					spell = sort
					break
				}
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewTextDisplay("# " + spell.Name),
					discord.NewTextDisplay("> *" + spell.Description + "*"),
				),
			)

			if err := event.CreateMessage(msg); err != nil {

			}
		},
	})
}
