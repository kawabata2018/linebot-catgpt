package main

import (
	"time"
	"unicode/utf8"

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

type APIUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type APIUsageRepository interface {
	Add(sid EventSourceID, usage APIUsage) error
}

type OpenAI interface {
	Request(prompt string) (string, *APIUsage, error)
	RequestWithHistory(prompt string, history []Chat) (string, *APIUsage, error)
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
	openai    OpenAI
	chatRepo  ChatRepository
	usageRepo APIUsageRepository
}

func NewApplicationService(openai OpenAI, chatRepo ChatRepository, usageRepo APIUsageRepository) *ApplicationService {
	return &ApplicationService{
		openai:    openai,
		chatRepo:  chatRepo,
		usageRepo: usageRepo,
	}
}

func (a *ApplicationService) Reply(input string, sid EventSourceID) string {
	start := time.Now()
	defer slog.Debug("execution time", "duration", time.Since(start))

	slog.Info("Print input", "input", input, "EventSourceID", sid)

	reply, usage, err := a.openai.Request(input)
	if err != nil {
		return "OpenAI APIから返事が来なかったにゃ"
	}

	slog.Info("Print reply", "reply", reply, "EventSourceID", sid)

	if err := a.chatRepo.Save(sid, Chat{Input: input, Reply: reply}); err != nil {
		slog.Warn("An error occurred while saving chat", "err", err)
	}
	if err := a.usageRepo.Add(sid, *usage); err != nil {
		slog.Warn("An error occurred while saving api usage", "err", err)
	}

	return reply
}

const (
	maxHistory   = 3
	maxInputSize = 200
)

func (a *ApplicationService) ReplyWithHistory(input string, sid EventSourceID) string {
	start := time.Now()
	defer slog.Debug("execution time", "duration", time.Since(start))

	slog.Info("Print input", "input", input, "EventSourceID", sid)

	// 文字数が一定の長さを上回る場合は弾く
	if utf8.RuneCountInString(input) > maxInputSize {
		return "ごめんなさいにゃ、飼い主の懐事情でそんなに長い文章には答えられないにゃ"
	}

	history, err := a.chatRepo.FetchHistory(sid, maxHistory)
	if err != nil {
		slog.Error("An error occured while fecthing chat history", err)
		return "なんかバグったにゃ"
	}

	reply, usage, err := a.openai.RequestWithHistory(input, history)
	if err != nil {
		return "OpenAI APIから返事が来なかったにゃ"
	}

	slog.Info("Print reply", "reply", reply, "EventSourceID", sid)

	if err := a.chatRepo.Save(sid, Chat{Input: input, Reply: reply}); err != nil {
		slog.Warn("An error occurred while saving chat", "err", err)
	}
	if err := a.usageRepo.Add(sid, *usage); err != nil {
		slog.Warn("An error occurred while saving api usage", "err", err)
	}

	return reply
}
