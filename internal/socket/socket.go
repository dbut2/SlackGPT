package socket

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/dbut2/slackgpt/pkg/events"
	"github.com/dbut2/slackgpt/pkg/openai"
	"github.com/dbut2/slackgpt/pkg/prompt"
	"github.com/dbut2/slackgpt/pkg/slackclient"
)

type Config struct {
	OpenAIToken   string
	SlackAppToken string
	SlackBotToken string
	Model         string
}

type Socket struct {
	sm           *socketmode.Client
	eventHandler events.Handler
}

func New(config Config) *Socket {
	sc := slack.New(config.SlackBotToken, slack.OptionDebug(true), slack.OptionAppLevelToken(config.SlackAppToken), slack.OptionLog(log.New(os.Stdout, "sc: ", log.Lshortfile|log.LstdFlags)))
	sm := socketmode.New(sc, socketmode.OptionDebug(true), socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)))

	scc := slackclient.New(sc)
	botID, err := scc.GetBotID()
	if err != nil {
		log.Fatal(err.Error())
	}
	enhancer := prompt.NewMessageGetter(scc, botID)

	var opts []openai.ClientOption
	if botID != "" {
		opts = append(opts, openai.WithBotID(botID))
	}
	if config.Model != "" {
		opts = append(opts, openai.WithModel(config.Model))
	}
	sender := openai.New(config.OpenAIToken, enhancer, scc, opts...)

	eventHandler := events.New(sender)

	return &Socket{
		sm:           sm,
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
