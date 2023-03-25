package main

import (
	"context"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
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
		slog.Error("Failed to parse env", err)
		return nil, ErrParseConfig
	}
	return &OpenAIAdapter{
		config: cfg,
	}, nil
}

const requestTimeout = 3 * time.Minute

func (a *OpenAIAdapter) Request(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	client := openai.NewClient(a.config.OpenAIAPIKey)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			MaxTokens:   300,
			Temperature: 0.9,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたはネコ型対話ロボットCatGPTにゃ、ネコ風に語尾は「にゃ」にしてくださいにゃん",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		slog.Error("ChatCompletionError", err)
		return "", ErrOpenAIAPIRequest
	}

	message := resp.Choices[0].Message.Content
	usage := resp.Usage

	slog.Debug("token usage", "prompt_tokens", usage.PromptTokens, "completion_tokens", usage.CompletionTokens, "total_tokens", usage.TotalTokens)
	return message, nil
}
