package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"
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

	app := NewApplicationService(openaiAdaper, firestoreRepo)

	linebot, err := NewLinebot(app)
	if err != nil {
		slog.Error("Failed to create Linebot")
		return
	}

	server, err := NewServer(linebot.Handler)
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
	server.GracefulShutdown()
}
