package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"onyx/bot/core"
	"onyx/bot/handlers"

	_ "onyx/bot/commands"
	_ "onyx/bot/events"
	"onyx/bot/locales"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	if err := locales.Load("locales"); err != nil {
		fmt.Printf("Warning: failed to load locales: %v\n", err)
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("Warning: DISCORD_TOKEN is not set")
	}

	adminIDsStr := os.Getenv("ADMIN_IDS")
	var adminIDs []string
	if adminIDsStr != "" {
		adminIDs = strings.Split(adminIDsStr, ",")
		for i := range adminIDs {
			adminIDs[i] = strings.TrimSpace(adminIDs[i])
		}
	}

	coreBot := &core.Bot{
		AdminIDs: adminIDs,
	}

	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentMessageContent,
			),
		),
		bot.WithRestConfigOpts(
			rest.WithDefaultAllowedMentions(discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			}),
		),
	)
	if err != nil {
		panic(err)
	}

	coreBot.Client = client

	handlers.SetupEvents(coreBot)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client.Close(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.OpenGateway(ctx); err != nil {
		fmt.Printf("Failed to open gateway (is token valid?): %v\n", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	fmt.Println("Shutting down bot...")
}
