package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/caarlos0/env/v7"
	"golang.org/x/exp/slog"
)

type firestoreConfig struct {
	ProjectID string `env:"PROJECT_ID,required"`
}

type FirestoreRepository struct {
	client         *firestore.Client
	collectionName string
}

func NewFirestoreRepository() (*FirestoreRepository, error) {
	cfg := firestoreConfig{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env", err)
		return nil, ErrParseConfig
	}
	client, err := firestore.NewClient(context.Background(), cfg.ProjectID)
	if err != nil {
		slog.Error("Failed to create firestore client", err)
		return nil, ErrNewFirestoreClient
	}
	return &FirestoreRepository{
		client:         client,
		collectionName: "catgpt",
	}, nil
}

var (
	jst = time.FixedZone("Asia/Tokyo", 9*60*60)
)

type document struct {
	EventSourceID string
	Input         string
	Reply         string
	Timestamp     time.Time
}

func (f *FirestoreRepository) Save(sid EventSourceID, chat Chat) error {
	ctx := context.Background()

	doc := document{
		EventSourceID: string(sid),
		Input:         chat.Input,
		Reply:         chat.Reply,
		Timestamp:     time.Now().In(jst),
	}

	_, _, err := f.client.Collection(f.collectionName).Add(ctx, doc)
	return err
}

func (f *FirestoreRepository) Close() error {
	slog.Debug("Firestore client closed")
	return f.client.Close()
}
