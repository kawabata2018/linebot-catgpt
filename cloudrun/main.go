package main

import (
	"os"
	"os/signal"
	"syscall"

	"log/slog"
)

func main() {
	slog.SetDefault(NewGCPLogger())

	openaiAdaper, err := NewOpenAIAdapter()
	if err != nil {
		slog.Error("Failed to create OpenAI adapter")
		return
	}
	firestoreRepo, err := NewFirestoreRepository()
	if err != nil {
		slog.Error("Failed to create Firestore repository")
		return
	}
	app := NewApplicationService(openaiAdaper, firestoreRepo, firestoreRepo)

	linebot, err := NewLinebot()
	if err != nil {
		slog.Error("Failed to create Linebot")
		return
	}
	httpHandler := linebot.CreateHandler(app.ReplyWithHistory, app.Unfollow)

	server, err := NewServer(httpHandler)
	if err != nil {
		slog.Error("Failed to create server")
		return
	}

	server.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)

	<-quit
	slog.Debug("SIGNAL received, then shutting down...")
	firestoreRepo.Close()
	if err := server.GracefulShutdown(); err != nil {
		slog.Error("Failed to gracefully shutdown server", "error", err)
	}
}
