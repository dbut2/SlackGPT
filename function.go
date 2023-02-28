package event

import (
	"log"
	"net/http"
	"os"

	"github.com/dbut2/slackgpt/internal/web"
)

func SlackEvent(w http.ResponseWriter, r *http.Request) {
	slackSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	pubsubProjectID := os.Getenv("PROJECT_ID")
	pubsubTopic := os.Getenv("PUBSUB_TOPIC")

	handler, err := web.New(web.Config{
		SlackSigningSecret: slackSigningSecret,
		SlackBotToken:      slackBotToken,
		PubsubProjectID:    pubsubProjectID,
		PubsubTopic:        pubsubTopic,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}

	err = handler.Handle(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}
}
