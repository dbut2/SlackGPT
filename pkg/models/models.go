package models

import (
	"time"
)

type Request struct {
	Prompt        string
	User          string
	Timestamp     time.Time
	SlackChannel  string
	SlackThreadTS string
	SlackMsgTS    string
}

type Response struct {
	Completion    string
	Timestamp     time.Time
	SlackChannel  string
	SlackThreadTS string
}
