package suggestion

import (
	"onyx/bot/core"
	"onyx/bot/locales"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func (m *SuggestionModule) Command() *discord.SlashCommandCreate {
	return &discord.SlashCommandCreate{
		Name:        "suggestion",
		Description: "Suggestions module",
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
		trad := locales.GetModule_SuggestionModule(event.Locale())
		modal := discord.NewModalCreate(
			"module-suggestion-all-submit", trad.Modal_title,
		).AddLabel(
			trad.Modal_title_label, discord.NewTextInput("submit-title", discord.TextInputStyleShort).WithMaxLength(100).WithPlaceholder(trad.Modal_title_placeholder).WithRequired(true),
		).AddLabel(
			trad.Modal_desc_label, discord.NewTextInput("submit-description", discord.TextInputStyleParagraph).WithMaxLength(100).WithPlaceholder(trad.Modal_desc_placeholder).WithRequired(true),
		)

		if m.Data.Content.AllowImages {
			modal = modal.AddLabel(
				trad.Modal_images_label, discord.NewFileUpload("submit-images").WithMinValues(0).WithMaxValues(10),
			)
		}

		event.Modal(modal)
	}

	return false
}

func (m *SuggestionModule) HandleModal(b *core.Bot, event *events.ModalSubmitInteractionCreate, action string, args []string) bool {
	if event.Data.CustomID == "module-suggestion-all-submit" {
		trad := locales.GetModule_SuggestionModule(event.Locale())
		title := event.Data.Text("submit-title")
		if len(title) > 0 {
			title = strings.ToUpper(string(title[0])) + title[1:]
		}
		desc := event.Data.Text("submit-description")

		cid, err := snowflake.Parse(m.Data.Main.Channel)
		if err != nil {
			event.CreateMessage(discord.NewMessageCreateV2(
				discord.NewContainer(discord.NewTextDisplay(trad.Error_channel_not_found)),
			).WithEphemeral(true))
			return false
		}
		channel, exist := event.Client().Caches.Channel(cid)
		if !exist {
			event.CreateMessage(discord.NewMessageCreateV2(
				discord.NewContainer(discord.NewTextDisplay(trad.Error_channel_not_found)),
			).WithEphemeral(true))
			return false
		}

		container := discord.NewContainer(
			discord.NewSection(
				discord.NewTextDisplayf("## %s", title),
				discord.NewTextDisplayf(trad.Suggestion_from, event.Member().User.ID.String()),
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
				discord.NewTextDisplay(trad.Submit_footer),
			),
		)

		var suggestionUrl string
		switch channel.Type() {
		case discord.ChannelTypeGuildText:
			msg, err := event.Client().Rest.CreateMessage(channel.ID(), msg)
			if err != nil {
				event.CreateMessage(discord.NewMessageCreateV2(
					discord.NewContainer(discord.NewTextDisplay(trad.Error_send_failed)),
				).WithEphemeral(true))
				return false
			}
			suggestionUrl = msg.JumpURL()

			if m.Data.Main.AllowDebate {
				event.Client().Rest.CreateThreadFromMessage(msg.ChannelID, msg.ID, discord.ThreadCreateFromMessage{
					Name: trad.Thread_name + event.Member().User.EffectiveName(),
				})
			}
		case discord.ChannelTypeGuildForum:
			postCreate := discord.ThreadChannelPostCreate{
				Name:    trad.Thread_name + event.Member().User.EffectiveName(),
				Message: msg,
			}
			post, err := event.Client().Rest.CreatePostInThreadChannel(channel.ID(), postCreate)
			if err != nil {
				event.CreateMessage(discord.NewMessageCreateV2(
					discord.NewContainer(discord.NewTextDisplay(trad.Error_send_failed)),
				).WithEphemeral(true))
				return false
			}
			suggestionUrl = post.Message.JumpURL()

			if !m.Data.Main.AllowDebate {
				locked := true
				event.Client().Rest.UpdateChannel(post.ID(), discord.GuildThreadUpdate{
					Locked: &locked,
				})
			}

		default:
			event.CreateMessage(discord.NewMessageCreateV2(
				discord.NewContainer(discord.NewTextDisplay(trad.Error_invalid_channel_type)),
			).WithEphemeral(true))
			return false
		}

		guild, exist := event.Guild()
		if !exist {
			return false
		}

		event.CreateMessage(discord.NewMessageCreateV2(
			discord.NewContainer(
				discord.NewSection(
					discord.NewTextDisplayf("## %s", trad.Success_title),
					discord.NewTextDisplayf(trad.Success_desc, suggestionUrl),
				).WithAccessory(discord.NewThumbnail(*guild.IconURL())),
			),
		).WithFlags(discord.MessageFlagEphemeral))
	}

	return false
}
