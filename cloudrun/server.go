package main

import (
	"context"
	"net"
	"net/http"
	"time"

	"log/slog"

	"github.com/caarlos0/env/v7"
)

type serverConfig struct {
	Port string `env:"PORT" envDefault:"8080"`
}

type Server struct {
	httpServer *http.Server
}

func NewServer(handler http.HandlerFunc) (*Server, error) {
	cfg := serverConfig{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env", "error", err)
		return nil, ErrParseConfig
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", cfg.Port),
		Handler: mux,
	}

	return &Server{
		httpServer: srv,
	}, nil
}

func (s *Server) Run() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				slog.Error("Failed to listen and serve", "error", err)
			}
			slog.Debug("Server closed")
		}
	}()
}

func (s *Server) GracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
