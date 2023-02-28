package socket

import (
	"log"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/dbut2/slackgpt/pkg/events"
)

type Config struct {
	OpenAIToken   string
	SlackAppToken string
	SlackBotToken string
	SlackBotID    string
	Model         string
}

type Socket struct {
	sm           *socketmode.Client
	eventHandler events.Handler
}

func New(socketClient *socketmode.Client, eventHandler events.Handler) *Socket {
	return &Socket{
		sm:           socketClient,
		eventHandler: eventHandler,
	}
}

func (s *Socket) Run() error {
	go func() {
		for e := range s.sm.Events {
			switch e.Type {
			case socketmode.EventTypeEventsAPI:
				s.sm.Ack(*e.Request)
				ev := e.Data.(slackevents.EventsAPIEvent)
				switch ev.Type {
				case slackevents.CallbackEvent:
					ie := ev.InnerEvent
					switch iv := ie.Data.(type) {
					case *slackevents.AppMentionEvent:
						err := s.eventHandler.HandleAppMentionEvent(iv)
						if err != nil {
							log.Print(err.Error())
						}
					case *slackevents.MessageEvent:
						err := s.eventHandler.HandleMessageEvent(iv)
						if err != nil {
							log.Print(err.Error())
						}
					}
				}
			}
		}
	}()

	return s.sm.Run()
}
