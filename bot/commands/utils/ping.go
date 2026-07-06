package utils

import (
	"fmt"
	"time"

	"onyx/bot/core"
	"onyx/bot/handlers"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func init() {
	handlers.RegisterCommand(handlers.Command{
		Name:        "ping",
		Description: "Check bot latency",
		Category:    "Utils",
		Create: discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Check bot latency",
		},
		Execute: func(b *core.Bot, event *events.ApplicationCommandInteractionCreate) {
			trad := locales.GetPing(event.Locale())
			guild, exist := event.Client().Caches.GuildCache().Get(*event.GuildID())

			gatewayLatency := event.Client().Gateway.Latency().Milliseconds()

			startREST := time.Now()
			_, _ = event.Client().Rest.GetGateway()
			restLatency := time.Since(startREST).Milliseconds()

			avatarURL := event.User().EffectiveAvatarURL()
			if exist {
				avatarURL = *guild.IconURL()
			}

			msg := discord.NewMessageCreateV2(
				discord.NewContainer(
					discord.NewSection(
						discord.NewTextDisplay(trad.Title),
						discord.NewTextDisplay(fmt.Sprintf(trad.Gateway_latency, gatewayLatency)),
						discord.NewTextDisplay(fmt.Sprintf(trad.Rest_latency, restLatency)),
					).WithAccessory(discord.NewThumbnail(avatarURL)),
				),
			)

			if err := event.CreateMessage(msg); err != nil {
				fmt.Println("Error sending ping message:", err)
			}
		},
	})
}
