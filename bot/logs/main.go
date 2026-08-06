package logs

import (
	"fmt"
	"os"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
)

type LogChannel string

const (
	SafetyLogsChannel LogChannel = "SUPPORT_SAFETY_LOGS"
)

type Logger struct {
	Client *bot.Client
}

func NewLogger(client *bot.Client) Logger {
	return Logger{
		Client: client,
	}
}

func (l Logger) SendLog(c LogChannel, msg discord.MessageCreate) {
	_ = godotenv.Load()
	cid := os.Getenv(string(c))
	scid, err := snowflake.Parse(cid)
	if err != nil {
		fmt.Printf("[Error] invalid snowflake for channel %s\n", c)
		return
	}

	if l.Client != nil {
		_, err = l.Client.Rest.CreateMessage(scid, msg)
		if err != nil {
			fmt.Printf("[Error] while sending log: \n%s\n", err.Error())
		}
	} else {
		fmt.Println("[Error] Logger Client is nil")
	}
}

