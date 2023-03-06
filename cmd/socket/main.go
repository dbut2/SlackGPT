package main

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/dbut2/slackgpt/internal/socket"
	"github.com/dbut2/slackgpt/pkg/events"
	"github.com/dbut2/slackgpt/pkg/openai"
	"github.com/dbut2/slackgpt/pkg/prompt"
	"github.com/dbut2/slackgpt/pkg/slackclient"
)

func main() {
	openAIToken := os.Getenv("OPENAI_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	slackBotID := os.Getenv("SLACK_BOT_ID")
	model := os.Getenv("MODEL")

	sc := slack.New(slackBotToken, slack.OptionDebug(true), slack.OptionAppLevelToken(slackAppToken), slack.OptionLog(log.New(os.Stdout, "sc: ", log.Lshortfile|log.LstdFlags)))
	sm := socketmode.New(sc, socketmode.OptionDebug(true), socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)))

	scc := slackclient.New(sc)
	enhancer := prompt.NewEnhancer(scc, slackBotID)
	sender := openai.New(openAIToken, enhancer, scc, model)
	eventHandler := events.New(sender)

	s := socket.New(sm, eventHandler)

	err := s.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
