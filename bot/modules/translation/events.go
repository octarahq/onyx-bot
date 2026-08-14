package translation

import (
	"fmt"
	"onyx/bot/core"
	"onyx/bot/utils"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *TranslationModule) HandleMessageCreate(b *core.Bot, e *events.MessageCreate) bool {
	if e.Message.Author.Bot {
		return false
	}

	channel, ok := e.Client().Caches.Channel(e.ChannelID)
	if !ok || channel.Type() != discord.ChannelTypeGuildNews {
		return false
	}

	if m.Data.Main.Channels == "" {
		return false
	}
	channelIDs := strings.Split(m.Data.Main.Channels, ",")

	var ch string
	for _, c := range channelIDs {
		c = strings.TrimSpace(c)
		if c == e.ChannelID.String() {
			ch = e.ChannelID.String()
		}
	}

	if ch != e.ChannelID.String() {
		return false
	}

	cacheKey := fmt.Sprintf("translation_ratelimit:%s", e.ChannelID.String())
	if _, limited := utils.Cache.Get(cacheKey); limited {
		return false
	}
	utils.Cache.Set(cacheKey, true, 5*time.Second)

	params := discord.ThreadCreateFromMessage{
		Name:                fmt.Sprintf("Traduction %s", m.Data.Main.Lang),
		AutoArchiveDuration: discord.AutoArchiveDuration1w,
	}

	thread, err := e.Client().Rest.CreateThreadFromMessage(e.ChannelID, e.MessageID, params)
	if err != nil {
		return false
	}

	content := e.Message.Content

	if len(content) > 2000 {
		content = fmt.Sprintf("%s...", content[0:1996])
	}

	t := utils.Translate(utils.TranslateParams{
		Query:  content,
		Source: "auto",
		Target: m.Data.Main.Lang,
	})

	trad := t.TranslatedText
	if len(trad) > 2000 {
		trad = fmt.Sprintf("%s...", trad[0:1996])
	}

	msg := discord.NewMessageCreateV2(
		discord.NewContainer(
			discord.NewTextDisplay(trad),
			discord.NewTextDisplayf("-# %s %s Translation", utils.TranslateLangs[m.Data.Main.Lang].Flag, utils.TranslateLangs[m.Data.Main.Lang].Name),
			discord.NewActionRow(
				discord.NewSecondaryButton("Translate", "translate-all-ephemeral"),
			),
		),
	)

	if _, err := e.Client().Rest.CreateMessage(thread.ID(), msg); err != nil {
		fmt.Printf("Error %s\n", err.Error())
		return false
	}

	return false
}
