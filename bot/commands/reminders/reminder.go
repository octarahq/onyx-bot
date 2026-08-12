package reminders

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"
	"onyx/bot/utils"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Reminder struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	GuildID   string    `gorm:"index" json:"guild_id"`
	ChannelID string    `gorm:"not null" json:"channel_id"`
	Content   string    `gorm:"not null" json:"content"`
	RemindAt  time.Time `gorm:"index;not null" json:"remind_at"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `gorm:"default:false;index" json:"completed"`
}

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "reminder",
		Description: "Manager yours reminders",
		Category:    "Reminder",
		Schema:      &Reminder{},
		Create: discord.SlashCommandCreate{
			Name:             "reminder",
			Description:      "Manager yours reminders",
			IntegrationTypes: []discord.ApplicationIntegrationType{discord.ApplicationIntegrationTypeGuildInstall, discord.ApplicationIntegrationTypeUserInstall},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "create",
					Description: "Create a reminder",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "content",
							Description: "The content that I will remind your",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "time",
							Description: "When will you want me to remind you",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "list",
					Description: "The list of your active reminders",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "delete",
					Description: "Delete a reminder",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "reminder",
							Description:  "The reminder to delete",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "edit",
					Description: "Edit a reminder",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "reminder",
							Description:  "The reminder to edit",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
			},
		},
		ExecuteAutocomplete: func(b *core.Bot, event *events.AutocompleteInteractionCreate) {
			if event.Data.SubCommandName == nil {
				return
			}

			switch *event.Data.SubCommandName {
			case "delete", "edit":
				query := strings.ToLower(event.Data.String("reminder"))
				userID := event.User().ID.String()

				var userReminders []Reminder
				dbQuery := b.DB.GormDB.Where("user_id = ? AND completed = ?", userID, false)
				if query != "" {
					dbQuery = dbQuery.Where("LOWER(content) LIKE ?", "%"+query+"%")
				}

				if err := dbQuery.Order("created_at DESC").Limit(25).Find(&userReminders).Error; err != nil {
					return
				}

				choices := make([]discord.AutocompleteChoice, 0, len(userReminders))
				for _, r := range userReminders {
					displayContent := r.Content
					if len(displayContent) > 90 {
						displayContent = displayContent[:87] + "..."
					}
					choices = append(choices, discord.AutocompleteChoiceString{
						Name:  displayContent,
						Value: strconv.Itoa(int(r.ID)),
					})
				}

				_ = event.Respond(discord.InteractionResponseTypeAutocompleteResult, discord.AutocompleteResult{
					Choices: choices,
				})
			}
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			slash := event.SlashCommandInteractionData()
			if slash.SubCommandName == nil {
				return
			}

			trad := locales.GetReminder(event.Locale())

			switch *slash.SubCommandName {
			case "create":
				var count int64
				if err := b.DB.GormDB.Model(&Reminder{}).Where("user_id = ? AND completed = ?", event.User().ID.String(), false).Count(&count).Error; err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_db))
					return
				}

				if count >= 25 {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_limit_reached))
					return
				}

				content := slash.String("content")
				duration := slash.String("time")
				t, err := utils.ParseDurationToTime(duration)
				if err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContentf(trad.Error_generic, err))
					return
				}

				now := time.Now()
				oneMinLater := now.Add(1 * time.Minute)
				oneMonthLater := now.AddDate(0, 1, 0)

				if t.Before(oneMinLater) || t.After(oneMonthLater) {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_invalid_time))
					return
				}

				var guildID string
				if event.GuildID() != nil {
					guildID = event.GuildID().String()
				}

				reminder := Reminder{
					UserID:    event.User().ID.String(),
					GuildID:   guildID,
					ChannelID: event.Channel().ID().String(),
					Content:   content,
					RemindAt:  t,
					CreatedAt: now,
				}

				if err := b.DB.GormDB.Create(&reminder).Error; err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_db))
					return
				}

				_ = event.CreateMessage(discord.NewMessageCreateV2().
					WithComponents(
						discord.NewContainer(
							discord.NewSection(
								discord.NewTextDisplay(trad.Create_success_title),
								discord.NewTextDisplayf(trad.Create_success_desc, content, utils.GenerateTimestamp(int(t.Unix()), utils.TimestampRelativeTime)),
							).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
						),
					),
				)
				return
			case "delete":
				id := slash.String("reminder")

				err := b.DB.GormDB.Where("id = ? AND user_id = ?", id, event.User().ID.String()).Delete(&Reminder{}).Error
				if err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContentf(trad.Error_generic, err))
					return
				}

				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Delete_success))
				return

			case "edit":
				id := slash.String("reminder")

				var reminder Reminder
				if err := b.DB.GormDB.Where("id = ? AND user_id = ? AND completed = ?", id, event.User().ID.String(), false).First(&reminder).Error; err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_not_found))
					return
				}

				modal := discord.NewModalCreate(
					fmt.Sprintf("reminder-%s-edit-%d", event.User().ID.String(), reminder.ID),
					trad.Edit_modal_title,
				).
					AddLabel(trad.Edit_modal_content_label, discord.NewShortTextInput("content").WithValue(reminder.Content).WithRequired(true)).
					AddLabel(trad.Edit_modal_time_label, discord.NewShortTextInput("time").WithRequired(false))

				_ = event.Respond(discord.InteractionResponseTypeModal, modal)
				return
			case "list":
				container, err := buildReminderListContainer(b, event.Locale(), event.User().ID.String(), 1)
				if err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContentf(trad.Error_generic, err))
					return
				}

				_ = event.CreateMessage(discord.NewMessageCreateV2(container))
				return
			}
		},
		ExecuteModal: func(b *core.Bot, event *events.ModalSubmitInteractionCreate) {
			customID := event.Data.CustomID
			parts := strings.Split(customID, "-")
			if len(parts) < 4 || parts[2] != "edit" {
				return
			}

			trad := locales.GetReminder(event.Locale())

			reminderID, err := strconv.Atoi(parts[3])
			if err != nil {
				return
			}

			content := event.Data.Text("content")
			duration := strings.TrimSpace(event.Data.Text("time"))

			updates := map[string]interface{}{
				"content": content,
			}

			var targetTime time.Time

			if duration != "" {
				t, err := utils.ParseDurationToTime(duration)
				if err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContentf(trad.Error_generic, err))
					return
				}

				now := time.Now()
				oneMinLater := now.Add(1 * time.Minute)
				oneMonthLater := now.AddDate(0, 1, 0)

				if t.Before(oneMinLater) || t.After(oneMonthLater) {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_invalid_time))
					return
				}

				updates["remind_at"] = t
				targetTime = t
			} else {
				var existing Reminder
				if err := b.DB.GormDB.Where("id = ? AND user_id = ?", reminderID, event.User().ID.String()).First(&existing).Error; err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_not_found))
					return
				}
				targetTime = existing.RemindAt
			}

			if err := b.DB.GormDB.Model(&Reminder{}).Where("id = ? AND user_id = ?", reminderID, event.User().ID.String()).Updates(updates).Error; err != nil {
				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(trad.Error_db))
				return
			}

			_ = event.CreateMessage(discord.NewMessageCreateV2().
				WithComponents(
					discord.NewContainer(
						discord.NewSection(
							discord.NewTextDisplay(trad.Edit_success_title),
							discord.NewTextDisplayf(trad.Edit_success_desc, content, utils.GenerateTimestamp(int(targetTime.Unix()), utils.TimestampRelativeTime)),
						).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
					),
				),
			)
		},
		ExecuteButton: func(b *core.Bot, event *events.ComponentInteractionCreate) {
			customID := event.Data.CustomID()
			parts := strings.Split(customID, "-")
			if len(parts) < 4 || parts[2] != "page" {
				return
			}

			page, err := strconv.Atoi(parts[3])
			if err != nil {
				return
			}

			container, err := buildReminderListContainer(b, event.Locale(), event.User().ID.String(), page)
			if err != nil {
				return
			}

			_ = event.UpdateMessage(discord.NewMessageUpdateV2(container))
		},
	})
}

func buildReminderListContainer(b *core.Bot, locale discord.Locale, userID string, page int) (discord.ContainerComponent, error) {
	trad := locales.GetReminder(locale)

	var totalCount int64
	if err := b.DB.GormDB.Model(&Reminder{}).Where("user_id = ? AND completed = ?", userID, false).Count(&totalCount).Error; err != nil {
		return discord.ContainerComponent{}, err
	}

	pageSize := 5
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * pageSize
	var userReminders []Reminder
	if err := b.DB.GormDB.Where("user_id = ? AND completed = ?", userID, false).
		Order("remind_at ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&userReminders).Error; err != nil {
		return discord.ContainerComponent{}, err
	}

	var components []discord.ContainerSubComponent
	components = append(components, discord.NewTextDisplay(trad.List_title))

	if len(userReminders) == 0 {
		components = append(components, discord.NewTextDisplay(trad.List_empty))
	} else {
		for _, r := range userReminders {
			relTime := utils.GenerateTimestamp(int(r.RemindAt.Unix()), utils.TimestampRelativeTime)
			displayText := fmt.Sprintf("%d - %s\n> %s", r.ID, relTime, r.Content)
			components = append(components, discord.NewTextDisplay(displayText))
		}
	}

	prevBtn := discord.NewSecondaryButton("<", fmt.Sprintf("reminder-%s-page-%d", userID, page-1))
	if page <= 1 {
		prevBtn.Disabled = true
	}

	pageBtn := discord.NewSecondaryButton(fmt.Sprintf("%d/%d", page, totalPages), fmt.Sprintf("reminder-%s-noop", userID))
	pageBtn.Disabled = true

	nextBtn := discord.NewSecondaryButton(">", fmt.Sprintf("reminder-%s-page-%d", userID, page+1))
	if page >= totalPages {
		nextBtn.Disabled = true
	}

	components = append(components, discord.NewActionRow(prevBtn, pageBtn, nextBtn))

	return discord.NewContainer(components...), nil
}
