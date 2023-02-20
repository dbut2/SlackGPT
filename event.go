package event

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"google.golang.org/protobuf/proto"

	"github.com/dbut2/slackgpt/proto/pkg"
)

func Event(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}

	sv, err := slack.NewSecretsVerifier(r.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}

	_, err = sv.Write(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}

	err = sv.Ensure()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}

	e, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err.Error())
	}

	switch e.Type {
	case slackevents.URLVerification:
		c := &slackevents.EventsAPIURLVerificationEvent{}
		err = json.Unmarshal(body, c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err.Error())
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("content-type", "text/plain")
		_, err = w.Write([]byte(c.Challenge))
		if err != nil {
			log.Fatal(err.Error())
		}
	case slackevents.CallbackEvent:
		ie := e.InnerEvent
		switch ev := ie.Data.(type) {
		case *slackevents.AppMentionEvent:
			err = HandleAppMentionEvent(ev)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Fatal(err.Error())
			}
		case *slackevents.MessageEvent:
			err = HandleMessageEvent(ev)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Fatal(err.Error())
			}
		}
	}
}

func HandleAppMentionEvent(a *slackevents.AppMentionEvent) error {
	return sendRequest(a.Channel, a.ThreadTimeStamp, a.Text)
}

func HandleMessageEvent(a *slackevents.MessageEvent) error {
	if a.BotID != "" {
		return nil
	}
	return sendRequest(a.Channel, a.ThreadTimeStamp, a.Text)
}

func sendRequest(channel string, thread_ts string, prompt string) error {
	req := &pkg.Request{
		Prompt:          prompt,
		Channel:         channel,
		ThreadTimestamp: thread_ts,
	}
	b, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	msg := &pubsub.Message{
		Data: b,
	}

	psc, err := pubsub.NewClient(context.Background(), os.Getenv("PROJECT_ID"))
	if err != nil {
		return err
	}

	res := psc.Topic(os.Getenv("PUBSUB_TOPIC")).Publish(context.Background(), msg)
	_, err = res.Get(context.Background())
	if err != nil {
		return err
	}

	return nil
}
