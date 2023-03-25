package main

import (
	"time"

	"golang.org/x/exp/slog"
)

type EventSourceType string

const (
	userEventSource    EventSourceType = "U"
	groupEventSource   EventSourceType = "G"
	roomEventSource    EventSourceType = "R"
	unknownEventSource EventSourceType = ""
)

type EventSourceID string

func (id EventSourceID) Type() EventSourceType {
	switch id[0:1] {
	case "U":
		return userEventSource
	case "G":
		return groupEventSource
	case "R":
		return roomEventSource
	}
	return unknownEventSource
}

type OpenAI interface {
	Request(prompt string) (string, error)
}

type Message struct {
	EventSourceID EventSourceID
	Input         string
	Reply         string
}

type MessageRepository interface {
	Save(m Message) error
}

type ApplicationService struct {
	openai      OpenAI
	messageRepo MessageRepository
}

func NewApplicationService(openai OpenAI, messageRepo MessageRepository) *ApplicationService {
	return &ApplicationService{
		openai:      openai,
		messageRepo: messageRepo,
	}
}

func (a *ApplicationService) Reply(input string, sid EventSourceID) string {
	start := time.Now()
	defer slog.Debug("execution time", "duration", time.Since(start))

	slog.Info("Print input", "input", input, "EventSourceID", sid)

	reply, err := a.openai.Request(input)
	if err != nil {
		return "OpenAI APIから返事が来なかったにゃ"
	}

	slog.Info("Print reply", "reply", reply, "EventSourceID", sid)

	a.messageRepo.Save(Message{
		EventSourceID: sid,
		Input:         input,
		Reply:         reply,
	})
	return reply
}
