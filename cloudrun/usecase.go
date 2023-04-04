package main

import (
	"context"
	"fmt"
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
	// APIリクエストに使用したトークン数を記録する
	Add(ctx context.Context, sid EventSourceID, usage APIUsage) error
	// その日の使用トークン数の合計を返す
	FetchTotalTokens(ctx context.Context, sid EventSourceID) (int, error)
}

type OpenAI interface {
	Request(ctx context.Context, prompt string) (string, *APIUsage, error)
	RequestWithHistory(ctx context.Context, prompt string, history []Chat) (string, *APIUsage, error)
}

type Chat struct {
	Input string
	Reply string
}

type ChatRepository interface {
	// チャット(入力、返答)を永続化する
	Save(ctx context.Context, sid EventSourceID, chat Chat) error
	// 直近max件のチャット履歴を取得する
	// 取得するリストは時系列順に並んでいる
	FetchHistory(ctx context.Context, sid EventSourceID, max int) ([]Chat, error)
	// 指定したEventSourceIDの全チャット履歴をアーカイブする
	Archive(ctx context.Context, sid EventSourceID) error
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
	ctx := context.Background()

	start := time.Now()
	defer slog.Debug("execution time", "duration", time.Since(start))

	reply, usage, err := a.openai.Request(ctx, input)
	if err != nil {
		return "OpenAI APIから返事が来なかったにゃ"
	}

	if err := a.chatRepo.Save(ctx, sid, Chat{Input: input, Reply: reply}); err != nil {
		slog.Warn("An error occurred while saving chat", "err", err)
	}
	if err := a.usageRepo.Add(ctx, sid, *usage); err != nil {
		slog.Warn("An error occurred while saving api usage", "err", err)
	}

	return reply
}

const (
	maxHistory   = 3
	maxInputSize = 200
)

func (a *ApplicationService) ReplyWithHistory(input string, sid EventSourceID) string {
	ctx := context.Background()

	start := time.Now()
	defer slog.Debug("execution time", "duration", time.Since(start))

	if input == "リセット" {
		if err := a.chatRepo.Archive(ctx, sid); err != nil {
			slog.Error("An error occured while archiving chat history", "err", err)
			return "なんかバグったにゃ"
		}
		slog.Info("Archived chat history", "EventSourceID", sid)
		return "チャット履歴をリセットしましたにゃ、今までの話は全部忘れちゃったにゃー"
	}

	if input == "トークン" {
		sum, err := a.usageRepo.FetchTotalTokens(ctx, sid)
		if err != nil {
			slog.Error("An error occured while fetching the sum of total tokens", "err", err)
			return "なんかバグったにゃ"
		}
		return fmt.Sprintf("今日の使用トークン数は\n%d だにゃ", sum)
	}

	// 文字数が一定の長さを上回る場合は弾く
	if utf8.RuneCountInString(input) > maxInputSize {
		return "ごめんなさいにゃ、飼い主の懐事情でそんなに長い文章には答えられないにゃ"
	}

	history, err := a.chatRepo.FetchHistory(ctx, sid, maxHistory)
	if err != nil {
		slog.Error("An error occured while fecthing chat history", "err", err)
		return "なんかバグったにゃ"
	}

	reply, usage, err := a.openai.RequestWithHistory(ctx, input, history)
	if err != nil {
		return "OpenAI APIから返事が来なかったにゃ"
	}

	if err := a.chatRepo.Save(ctx, sid, Chat{Input: input, Reply: reply}); err != nil {
		slog.Warn("An error occurred while saving chat", "err", err)
	}
	if err := a.usageRepo.Add(ctx, sid, *usage); err != nil {
		slog.Warn("An error occurred while saving api usage", "err", err)
	}

	return reply
}

func (a *ApplicationService) Unfollow(sid EventSourceID) {
	ctx := context.Background()
	if err := a.chatRepo.Archive(ctx, sid); err != nil {
		slog.Error("An error occured while archiving chat history", "err", err)
		return
	}
	slog.Info("Archived chat history", "EventSourceID", sid)
}
