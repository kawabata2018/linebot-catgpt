package main

import (
	"context"
	"time"

	"log/slog"

	"github.com/caarlos0/env/v7"
	"github.com/sashabaranov/go-openai"
)

type openaiConfig struct {
	OpenAIAPIKey string `env:"OPENAI_API_KEY,required"`
}

type OpenAIAdapter struct {
	config openaiConfig
}

func NewOpenAIAdapter() (*OpenAIAdapter, error) {
	cfg := openaiConfig{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env", "error", err)
		return nil, ErrParseConfig
	}
	return &OpenAIAdapter{
		config: cfg,
	}, nil
}

const requestTimeout = 100 * time.Second

func (a *OpenAIAdapter) Request(ctx context.Context, prompt string) (string, *APIUsage, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	client := openai.NewClient(a.config.OpenAIAPIKey)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			MaxTokens:   300,
			Temperature: 0.9,
			Messages: createMessages(
				"あなたはネコ型対話ロボット「CatGPT」にゃ、ネコ風に語尾は「にゃ」にしてくださいにゃん",
				prompt,
				nil,
			),
		},
	)
	if err != nil {
		slog.Error("ChatCompletionError", "error", err)
		return "", nil, ErrOpenAIAPIRequest
	}

	message := resp.Choices[0].Message.Content
	usage := &APIUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	return message, usage, nil
}

func (a *OpenAIAdapter) RequestWithHistory(ctx context.Context, prompt string, history []Chat) (string, *APIUsage, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	client := openai.NewClient(a.config.OpenAIAPIKey)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			MaxTokens:   300,
			Temperature: 0.9,
			Messages: createMessages(
				"あなたはネコ型対話ロボット「CatGPT」にゃ、ネコ風に語尾は「にゃ」にしてくださいにゃん",
				prompt,
				history,
			),
		},
	)
	if err != nil {
		slog.Error("ChatCompletionError", "error", err)
		return "", nil, ErrOpenAIAPIRequest
	}

	message := resp.Choices[0].Message.Content
	usage := &APIUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	return message, usage, nil
}

func createMessages(system, prompt string, history []Chat) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0, len(history)+2)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: system,
	})
	for _, chat := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: chat.Input,
		})
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: chat.Reply,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	slog.Debug("Print messages", "messages", messages)

	return messages
}
