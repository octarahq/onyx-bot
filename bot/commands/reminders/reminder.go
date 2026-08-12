package reminders

import (
	"strconv"
	"strings"
	"time"

	"onyx/bot/core"
	"onyx/bot/handlers"
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

			switch *slash.SubCommandName {
			case "create":
				var count int64
				if err := b.DB.GormDB.Model(&Reminder{}).Where("user_id = ? AND completed = ?", event.User().ID.String(), false).Count(&count).Error; err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent("Oups, une erreur s'est produite."))
					return
				}

				if count >= 25 {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(":x: Vous ne pouvez pas avoir plus de 25 rappels."))
					return
				}

				content := slash.String("content")
				duration := slash.String("time")
				t, err := utils.ParseDurationToTime(duration)
				if err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContentf(":x: Un problème est survenu : %s", err))
					return
				}

				now := time.Now()
				oneMinLater := now.Add(1 * time.Minute)
				oneMonthLater := now.AddDate(0, 1, 0)

				if t.Before(oneMinLater) || t.After(oneMonthLater) {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent(":x: Le temps doit être compris entre 1min et 1 mois."))
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
					_ = event.CreateMessage(discord.NewMessageCreate().WithContent("Oups, une erreur s'est produite."))
					return
				}

				_ = event.CreateMessage(discord.NewMessageCreateV2().
					WithComponents(
						discord.NewContainer(
							discord.NewSection(
								discord.NewTextDisplay("## C'est noté !"),
								discord.NewTextDisplayf("Je vous rappellerai : `%s`\n> %s", content, utils.GenerateTimestamp(int(t.Unix()), utils.TimestampRelativeTime)),
							).WithAccessory(discord.NewThumbnail(event.User().EffectiveAvatarURL())),
						),
					),
				)
				return
			case "delete":
				id := slash.String("reminder")

				err := b.DB.GormDB.Where("id = ? AND user_id = ?", id, event.User().ID.String()).Delete(&Reminder{}).Error
				if err != nil {
					_ = event.CreateMessage(discord.NewMessageCreate().WithContentf(":x: Un problème est survenu : %s", err))
					return
				}

				_ = event.CreateMessage(discord.NewMessageCreate().WithContent(":white_check_mark: Ce rappel a été supprimé"))
				return

			}
		},
	})
}
