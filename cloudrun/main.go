package main

import (
	"golang.org/x/exp/slog"
)

func main() {
	slog.SetDefault(NewGCPLogger())

	openaiAdaper, err := NewOpenAIAdapter()
	if err != nil {
		slog.Error("Failed to create OpenAI adapter")
		return
	}
	app := NewApplicationService(openaiAdaper)

	linebot, err := NewLinebot(app)
	if err != nil {
		slog.Error("Failed to create Linebot")
		return
	}

	server, err := NewServer()
	if err != nil {
		slog.Error("Failed to create server")
		return
	}

	server.Run(linebot.Handler)
}
