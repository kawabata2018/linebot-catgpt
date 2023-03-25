package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v7"
	"golang.org/x/exp/slog"
)

type serverConfig struct {
	Port string `env:"PORT" envDefault:"8080"`
}

type Server struct {
	config serverConfig
}

func NewServer() (*Server, error) {
	cfg := serverConfig{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env", err)
		return nil, ErrParseConfig
	}

	return &Server{
		config: cfg,
	}, nil
}

func (s *Server) Run(handler http.HandlerFunc) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", s.config.Port),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				slog.Error("Failed to listen and serve", err)
			}
			slog.Debug("Server closed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	slog.Debug("SIGNAL received, then shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Failed to gracefully shutdown", err)
	}
}
