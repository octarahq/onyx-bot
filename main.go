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
	"onyx/bot/db"
	"onyx/bot/handlers"
	"onyx/bot/logs"

	_ "onyx/bot/commands"
	_ "onyx/bot/events"
	"onyx/bot/locales"
	"onyx/bot/modules"

	"onyx/bot/api"
	_ "onyx/bot/api/routes"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/joho/godotenv"
)

func main() {
	var version = "1.2.0"
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

	db := db.New()

	coreBot := &core.Bot{
		AdminIDs: adminIDs,
		DB:       db,
		Commands: handlers.Commands,
		Events:   handlers.Events,
		Modules:  modules.RegisteredModules,
		Version:  version,
	}

	for _, mod := range coreBot.Modules {
		if dbAware, ok := mod.(core.DatabaseAware); ok {
			schema := dbAware.Schema()
			if slice, isSlice := schema.([]interface{}); isSlice {
				db.GormDB.AutoMigrate(slice...)
			} else {
				db.GormDB.AutoMigrate(schema)
			}
		}
	}

	for _, cmd := range coreBot.Commands {
		if cmd.Schema != nil {
			if slice, isSlice := cmd.Schema.([]interface{}); isSlice {
				db.GormDB.AutoMigrate(slice...)
			} else {
				db.GormDB.AutoMigrate(cmd.Schema)
			}
		}
	}

	var connectedSince = time.Now()

	client, err := disgo.New(token,
		bot.WithEventListenerFunc(func(e *events.Ready) {
			connectedSince = time.Now()
		}),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentMessageContent,
				gateway.IntentGuildMembers,
			),
		),
		bot.WithRestConfigOpts(
			rest.WithDefaultAllowedMentions(discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			}),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagsAll),
		),
		bot.WithMemberChunkingFilter(bot.MemberChunkingFilterAll),
	)
	if err != nil {
		panic(err)
	}

	coreBot.Client = client
	coreBot.ConnectedSince = connectedSince

	logger := logs.NewLogger(client)
	coreBot.Logger = logger

	for _, m := range coreBot.Modules {
		if ml, ok := m.(core.ModuleLogger); ok {
			coreBot.ModuleLogger = ml
			break
		}
	}

	handlers.SetupEvents(coreBot)

	go api.Start(coreBot)

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
