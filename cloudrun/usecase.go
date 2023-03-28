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
	RequestWithHistory(prompt string, history []Chat) (string, error)
}

type Chat struct {
	Input string
	Reply string
}

type ChatRepository interface {
	// チャット(入力、返答)を永続化する
	Save(sid EventSourceID, chat Chat) error
	// 直近max件のチャット履歴を取得する
	// 取得するリストは時系列順に並んでいる
	FetchHistory(sid EventSourceID, max int) ([]Chat, error)
}

type ApplicationService struct {
	openai   OpenAI
	chatRepo ChatRepository
}

func NewApplicationService(openai OpenAI, chatRepo ChatRepository) *ApplicationService {
	return &ApplicationService{
		openai:   openai,
		chatRepo: chatRepo,
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

	a.chatRepo.Save(sid, Chat{
		Input: input,
		Reply: reply,
	})
	return reply
}

const (
	maxHistory = 3
)

func (a *ApplicationService) ReplyWithHistory(input string, sid EventSourceID) string {
	start := time.Now()
	defer slog.Debug("execution time", "duration", time.Since(start))

	slog.Info("Print input", "input", input, "EventSourceID", sid)

	history, err := a.chatRepo.FetchHistory(sid, maxHistory)
	if err != nil {
		slog.Error("An error occured while fecthing chat history", err)
		return "なんかバグったにゃ"
	}

	reply, err := a.openai.RequestWithHistory(input, history)
	if err != nil {
		return "OpenAI APIから返事が来なかったにゃ"
	}

	slog.Info("Print reply", "reply", reply, "EventSourceID", sid)

	a.chatRepo.Save(sid, Chat{
		Input: input,
		Reply: reply,
	})
	return reply
}
