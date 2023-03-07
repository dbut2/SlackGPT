package main

import (
	"log"
	"os"

	"github.com/dbut2/slackgpt/internal/socket"
)

func main() {
	openAIToken := os.Getenv("OPENAI_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackBotID := os.Getenv("SLACK_BOT_ID")

	config := socket.Config{
		OpenAIToken:   openAIToken,
		SlackAppToken: slackAppToken,
		SlackBotToken: slackBotToken,
		SlackBotID:    slackBotID,
	}

	s := socket.New(config)

	err := s.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
