package suggestion

import (
	"onyx/bot/core"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func (m *SuggestionModule) Command() *discord.SlashCommandCreate {
	return &discord.SlashCommandCreate{
		Name:        "suggestion",
		Description: "Module de suggestions",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "submit",
				Description: "Submit a suggestion",
			},
		},
	}
}

func (m *SuggestionModule) HandleCommand(b *core.Bot, event *events.ApplicationCommandInteractionCreate) bool {
	if !m.IsEnabled() {
		return false
	}

	slash := event.SlashCommandInteractionData()

	if slash.SubCommandName == nil {
		return false
	}

	switch *slash.SubCommandName {
	case "submit":
		modal := discord.NewModalCreate(
			"module-suggestion-all-submit", "Submit a suggestion",
		).AddLabel(
			"suggestion title", discord.NewTextInput("submit-title", discord.TextInputStyleShort).WithMaxLength(100).WithPlaceholder("A new feature").WithRequired(true),
		).AddLabel(
			"suggestion description", discord.NewTextInput("submit-description", discord.TextInputStyleParagraph).WithMaxLength(100).WithPlaceholder("A feature who can...").WithRequired(true),
		)

		if m.Data.Content.AllowImages {
			modal = modal.AddLabel(
				"suggestion images", discord.NewFileUpload("submit-images").WithMinValues(0).WithMaxValues(10),
			)
		}

		event.Modal(modal)
	}

	return false
}

func (m *SuggestionModule) HandleModal(b *core.Bot, event *events.ModalSubmitInteractionCreate, action string, args []string) bool {
	if event.Data.CustomID == "module-suggestion-all-submit" {
		title := event.Data.Text("submit-title")
		desc := event.Data.Text("submit-description")

		cid, err := snowflake.Parse(m.Data.Main.Channel)
		if err != nil {
			return false
		}
		channel, exist := event.Client().Caches.Channel(cid)
		if !exist {
			return false
		}

		container := discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplayf("## %s", title),
				discord.NewTextDisplayf("%s", desc),
			).WithAccessory(discord.NewThumbnail(event.Member().EffectiveAvatarURL())),
		)

		attachments := event.Data.Attachments("submit-images")

		if m.Data.Content.AllowImages {
			if len(attachments) > 0 {
				var items []discord.MediaGalleryItem
				for _, a := range attachments {
					if a.ContentType != nil && strings.Contains(*a.ContentType, "image") {
						items = append(items, discord.MediaGalleryItem{
							Media: discord.UnfurledMediaItem{
								URL: a.URL,
							},
						})
					}
				}
				if len(items) > 0 {
					container = container.AddComponents(
						discord.NewMediaGallery(items...),
					)
				}
			}
		}

		msg := discord.NewMessageCreateV2(
			container.AddComponents(
				discord.NewTextDisplayf("-# Suggestion de <@%s>", event.Member().User.ID.String()),
			),
		)

		switch channel.Type() {
		case discord.ChannelTypeGuildText:
			event.Client().Rest.CreateMessage(channel.ID(), msg)
		case discord.ChannelTypeGuildForum:
			postCreate := discord.ThreadChannelPostCreate{
				Name:    "Suggestion de " + event.Member().User.Username,
				Message: msg,
			}
			event.Client().Rest.CreatePostInThreadChannel(channel.ID(), postCreate)
		default:
			return false
		}
	}

	return false
}
