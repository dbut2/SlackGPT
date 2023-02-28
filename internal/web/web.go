package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"cloud.google.com/go/pubsub"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/dbut2/slackgpt/pkg/events"
	ps "github.com/dbut2/slackgpt/pkg/pubsub"
	"github.com/dbut2/slackgpt/pkg/verifier"
)

type Config struct {
	SlackSigningSecret string
	SlackBotToken      string
	PubsubProjectID    string
	PubsubTopic        string
}

type Web struct {
	verifier     *verifier.Verifier
	slack        *slack.Client
	eventHandler events.Handler
}

func New(config Config) (*Web, error) {
	verifier := verifier.New(config.SlackSigningSecret)

	sc := slack.New(config.SlackBotToken)

	psc, err := pubsub.NewClient(context.Background(), config.PubsubProjectID)
	if err != nil {
		return nil, err
	}
	topic := psc.Topic(config.PubsubTopic)
	sender := ps.New(topic)

	return &Web{
		verifier:     verifier,
		slack:        sc,
		eventHandler: events.New(sc, sender),
	}, nil
}

func (h *Web) Handle(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = h.verifier.Verify(r.Header, body)
	if err != nil {
		return err
	}

	e, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		return err
	}

	switch e.Type {
	case slackevents.URLVerification:
		challenge, err := h.handleURLVerification(body)
		if err != nil {
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("content-type", "text/plain")
		_, err = w.Write([]byte(challenge))
		if err != nil {
			return err
		}
	case slackevents.CallbackEvent:
		err = h.handleCallbackEvent(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Web) handleURLVerification(body []byte) (string, error) {
	c := &slackevents.EventsAPIURLVerificationEvent{}
	err := json.Unmarshal(body, c)
	if err != nil {
		return "", err
	}

	return c.Challenge, nil
}

func (h *Web) handleCallbackEvent(e slackevents.EventsAPIEvent) error {
	ie := e.InnerEvent
	switch ev := ie.Data.(type) {
	case *slackevents.AppMentionEvent:
		err := h.eventHandler.HandleAppMentionEvent(ev)
		if err != nil {
			return err
		}
	case *slackevents.MessageEvent:
		err := h.eventHandler.HandleMessageEvent(ev)
		if err != nil {
			return err
		}
	}

	return nil
}
