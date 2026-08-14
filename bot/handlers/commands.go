package handlers

import (
	"fmt"
	"os"
	"strings"

	"onyx/bot/constants"
	"onyx/bot/core"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Command = core.Command

var Commands = make(map[string]Command)

func RegisterCommand(cmd Command) {
	Commands[cmd.Name] = cmd
}

func SyncCommands(b *core.Bot) error {
	for _, mod := range b.Modules {
		if modCmd, ok := mod.(core.ModuleCommand); ok {
			cmdCreate := modCmd.Command()
			if cmdCreate != nil && cmdCreate.Name != "" {
				mc := modCmd
				modRef := mod
				RegisterCommand(Command{
					Name:        cmdCreate.Name,
					Description: cmdCreate.Description,
					Category:    "Modules",
					Create:      *cmdCreate,
					Execute: func(bot *core.Bot, event *events.ApplicationCommandInteractionCreate) {
						var guildIDStr string
						if event.GuildID() != nil {
							guildIDStr = event.GuildID().String()
							if dbAware, ok := modRef.(core.DatabaseAware); ok {
								_ = dbAware.LoadData(bot.DB.GormDB, guildIDStr)
							}
						}
						g, hasGuild := event.Guild()

						if !modRef.IsEnabled() {
							siteURL := os.Getenv("SITE_URL")
							if siteURL == "" {
								siteURL = "https://onyx.octara.xyz"
							}
							configURL := fmt.Sprintf("%s/dashboard/guilds/%s/%s", strings.TrimSuffix(siteURL, "/"), guildIDStr, modRef.Metadata().Name)

							var iconURL string
							if hasGuild && g.IconURL() != nil {
								iconURL = *g.IconURL()
							} else {
								iconURL = constants.DISCORD_DEFAULT_AVATAR_URL
							}

							_ = event.CreateMessage(discord.NewMessageCreateV2(
								discord.NewContainer(
									discord.NewSection(
										discord.NewTextDisplay("## Module désactivé"),
										discord.NewTextDisplayf("> Le module **%s** est actuellement désactivé sur ce serveur. Vous pouvez l'activer et le configurer depuis le dashboard.", modRef.Metadata().Label(event.Locale())),
									).WithAccessory(discord.NewThumbnail(iconURL)),
									discord.NewActionRow(
										discord.NewLinkButton("Configurer", configURL),
									),
								),
							))
							return
						}

						mc.HandleCommand(bot, event)
					},
				})
			}
		}
	}

	b.Commands = Commands

	var cmds []discord.ApplicationCommandCreate
	for name, cmd := range Commands {
		create := cmd.Create

		if slashCmd, ok := create.(discord.SlashCommandCreate); ok {
			if slashCmd.NameLocalizations == nil {
				slashCmd.NameLocalizations = make(map[discord.Locale]string)
			}
			if slashCmd.DescriptionLocalizations == nil {
				slashCmd.DescriptionLocalizations = make(map[discord.Locale]string)
			}

			for locale, metas := range locales.Metas {
				if meta, exists := metas[name]; exists {
					if meta.Name != "" {
						slashCmd.NameLocalizations[locale] = meta.Name
					}
					if meta.Description != "" {
						slashCmd.DescriptionLocalizations[locale] = meta.Description
					}
					if len(meta.Options) > 0 && len(slashCmd.Options) > 0 {
						slashCmd.Options = localizeOptions(slashCmd.Options, locale, meta)
					}
				}
			}
			cmds = append(cmds, slashCmd)
		} else {
			cmds = append(cmds, create)
		}
	}

	_, err := b.Client.Rest.SetGlobalCommands(b.Client.ApplicationID, cmds)
	return err
}
