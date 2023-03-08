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
	model := os.Getenv("MODEL")

	config := socket.Config{
		OpenAIToken:   openAIToken,
		SlackAppToken: slackAppToken,
		SlackBotToken: slackBotToken,
		Model:         model,
	}

	s := socket.New(config)

	err := s.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
