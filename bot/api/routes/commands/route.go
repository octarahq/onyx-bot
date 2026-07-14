package commands

import (
	"net/http"
	"onyx/bot/api"
	"onyx/bot/core"
	"onyx/bot/locales"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/gin-gonic/gin"
)

type Subcommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Category struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}
type Command struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Subcommands []Subcommand `json:"subcommands"`
	Category    Category     `json:"category"`
}

func init() {
	api.AddRoute(api.Route{
		Method:  http.MethodGet,
		Path:    "/commands",
		Handler: handleCommandList,
	})
}

func handleCommandList(c *gin.Context) {
	var commands []Command
	bot, exists := c.MustGet("bot").(*core.Bot)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot context not found", "error_code": "INTERNAL_ERROR"})
		return
	}

	lang := c.Query("lang")
	headerLang := c.GetHeader("Accept-Language")

	if lang == "" {
		lang = headerLang
		if len(lang) > 2 {
			lang = lang[:2]
		}
	}
	if lang == "" {
		lang = string(discord.LocaleEnglishUS)
	}
	locale := discord.Locale(lang)

	for _, cmd := range bot.Commands {
		var command Command
		meta := locales.GetMeta(locale, cmd.Name)

		command.Name = cmd.Name
		if meta.Name != "" {
			command.Name = meta.Name
		}

		command.Description = cmd.Description
		if meta.Description != "" {
			command.Description = meta.Description
		}

		command.Category = Category{
			Id:    strings.ToLower(cmd.Category),
			Label: cmd.Category,
		}

		if slashCmd, ok := cmd.Create.(discord.SlashCommandCreate); ok {
			for _, opt := range slashCmd.Options {
				if subCmd, ok := opt.(discord.ApplicationCommandOptionSubCommand); ok {
					subName := subCmd.Name
					subDesc := subCmd.Description

					if subMeta, ok := meta.Options[subCmd.Name]; ok {
						if subMeta.Name != "" {
							subName = subMeta.Name
						}
						if subMeta.Description != "" {
							subDesc = subMeta.Description
						}
					}

					command.Subcommands = append(command.Subcommands, Subcommand{
						Name:        subName,
						Description: subDesc,
					})
				}
			}
		}

		commands = append(commands, command)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": commands,
	})
}
