package openai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sashabaranov/go-openai"

	"github.com/dbut2/slackgpt/pkg/models"
	"github.com/dbut2/slackgpt/pkg/prompt"
	"github.com/dbut2/slackgpt/pkg/slackgpt"
)

type Client struct {
	openai    *openai.Client
	mg        prompt.MessageGetter
	botID     string
	responder slackgpt.Responder
	model     string
	separator string
}

func New(token string, mg prompt.MessageGetter, responder slackgpt.Responder, opts ...ClientOption) *Client {
	client := openai.NewClient(token)

	c := &Client{
		openai:    client,
		mg:        mg,
		responder: responder,
		model:     openai.GPT3Dot5Turbo,
		separator: "\n\n---\n\n",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ClientOption func(client *Client)

func WithBotID(botID string) ClientOption {
	return func(c *Client) {
		c.botID = botID
	}
}

func WithModel(model string) ClientOption {
	return func(c *Client) {
		if model == "" {
			return
		}
		c.model = model
	}
}

func WithSeparator(separator string) ClientOption {
	return func(c *Client) {
		c.separator = separator
	}
}

func (c *Client) Send(ctx context.Context, req models.Request) error {
	var resp string
	var err error
	switch c.model {
	case
		openai.GPT3Dot5Turbo,
		openai.GPT3Dot5Turbo0301:
		resp, err = c.requestChat(ctx, req)
	case
		openai.GPT3TextDavinci003,
		openai.GPT3TextDavinci002,
		openai.GPT3TextDavinci001,
		openai.GPT3TextCurie001,
		openai.GPT3TextBabbage001,
		openai.GPT3TextAda001:
		resp, err = c.requestCompletion(ctx, req)
	default:
		err = errors.New("unsupported model")
	}
	if err != nil {
		return err
	}

	return c.responder.Respond(context.Background(), models.Response{
		Completion:    resp,
		Timestamp:     time.Now(),
		SlackChannel:  req.SlackChannel,
		SlackThreadTS: req.SlackThreadTS,
	})
}

type apiFunc[U, V any] func(context.Context, U) (V, error)

func (c *Client) requestCompletion(ctx context.Context, req models.Request) (string, error) {
	msgs, err := c.mg.GetMessages(req)
	if err != nil {
		return "", err
	}

	ss := c.formatMessages(msgs)
	post := c.formatMessage(prompt.Message{Name: c.botID})

	p := shrinkMsgs((4000-250)*4, ss, post)

	r := openai.CompletionRequest{
		Model:     c.model,
		Prompt:    p,
		MaxTokens: 250,
		Stop:      []string{c.separator},
		User:      req.User,
	}

	resp, err := request(ctx, c.openai.CreateCompletion, r)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) < 1 {
		return "", errors.New("no completions returned")
	}
	if resp.Choices[0].Text == "" {
		return "", errors.New("empty completion text")
	}
	return resp.Choices[0].Text, nil
}

func (c *Client) formatMessages(msgs []prompt.Message) []string {
	s := make([]string, len(msgs))
	var err error
	for i, msg := range msgs {
		s[i] = c.formatMessage(msg) + c.separator
		if err != nil {
			return nil
		}
	}
	return s
}

func (c *Client) formatMessage(msg prompt.Message) string {
	return fmt.Sprintf("<@%s>: %s", msg.Name, msg.Message)
}

func (c *Client) requestChat(ctx context.Context, req models.Request) (string, error) {
	msgs, err := c.mg.GetMessages(req)
	if err != nil {
		return "", err
	}

	sms := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are SlackGPT, a Slack bot built by <@UU3TUL99S>. Answer as concisely as possible.",
		},
	}
	ms := mapMessages(msgs)

	r := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		Messages:  shrinkOpenAIMsgs((4000-250)*4, sms, ms),
		MaxTokens: 250,
		User:      req.User,
	}

	resp, err := request(ctx, c.openai.CreateChatCompletion, r)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) < 1 {
		return "", errors.New("no completions returned")
	}
	if resp.Choices[0].Message.Content == "" {
		return "", errors.New("empty completion text")
	}
	return resp.Choices[0].Message.Content, nil
}

func request[U, V any](ctx context.Context, f apiFunc[U, V], r U) (V, error) {
	resp, err := f(ctx, r)

	bo := backoff.ExponentialBackOff{
		InitialInterval: time.Second,
		Multiplier:      1.5,
	}
	for err != nil {
		if reqErr, ok := err.(*openai.RequestError); ok {
			if reqErr.StatusCode == http.StatusTooManyRequests {
				nbo := bo.NextBackOff()
				log.Printf("OpenAI rate limit: %s", nbo.String())
				time.Sleep(nbo)
				resp, err = f(ctx, r)
			} else {
				return *new(V), err
			}
		} else {
			return *new(V), err
		}
	}

	return resp, nil
}

var roleMap = map[prompt.Role]string{
	prompt.SystemRole:    openai.ChatMessageRoleSystem,
	prompt.AssistantRole: openai.ChatMessageRoleAssistant,
	prompt.UserRole:      openai.ChatMessageRoleUser,
}

func mapMessages(msgs []prompt.Message) []openai.ChatCompletionMessage {
	var ms []openai.ChatCompletionMessage
	for _, msg := range msgs {
		ms = append(ms, openai.ChatCompletionMessage{
			Role:    roleMap[msg.Role],
			Content: msg.Message,
			Name:    msg.Name,
		})
	}
	return ms
}

func shrinkOpenAIMsgs(max int, sms []openai.ChatCompletionMessage, ms []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	for msgsLength(ms)+msgsLength(sms) > max {
		ms = ms[1:]
	}
	return append(sms, ms...)
}

func msgsLength(msgs []openai.ChatCompletionMessage) int {
	return getLength(msgs, func(msg openai.ChatCompletionMessage) int {
		return len(msg.Role) + len(msg.Content) + len(msg.Name)
	})
}

func shrinkMsgs(max int, s []string, post string) string {
	for strSliceLength(s)+len(post) > max {
		s = s[1:]
	}
	return strings.Join(s, "") + post
}

func strSliceLength(ss []string) int {
	return getLength(ss, func(s string) int {
		return len(s)
	})
}

func getLength[T any](items []T, lengthOf func(T) int) int {
	c := 0
	for _, v := range items {
		c += lengthOf(v)
	}
	return c
}
