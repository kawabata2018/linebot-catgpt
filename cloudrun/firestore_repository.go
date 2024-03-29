package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/caarlos0/env/v7"
	"google.golang.org/api/iterator"
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

func (f *FirestoreRepository) Save(ctx context.Context, sid EventSourceID, chat Chat) error {
	doc := document{
		EventSourceID: string(sid),
		Input:         chat.Input,
		Reply:         chat.Reply,
		Timestamp:     time.Now().In(jst),
	}

	_, _, err := f.client.Collection(f.collectionName).Add(ctx, doc)
	return err
}

func (f *FirestoreRepository) FetchHistory(ctx context.Context, sid EventSourceID, max int) ([]Chat, error) {
	query := f.client.Collection(f.collectionName).
		Where("EventSourceID", "==", sid).
		OrderBy("Timestamp", firestore.Desc).
		Limit(max)

	iter := query.Documents(ctx)
	defer iter.Stop()

	chats := make([]Chat, 0, max)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var c Chat
		if err := doc.DataTo(&c); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}

	// タイムスタンプ降順になっているので、時系列順に直してあげる
	slices.Reverse(chats)
	return chats, nil
}

type archivedDocument struct {
	EventSourceID string
	Input         string
	Reply         string
	Timestamp     time.Time
	ArchivedAt    time.Time
}

func (f *FirestoreRepository) Archive(ctx context.Context, sid EventSourceID) error {
	sourceColl := f.client.Collection(f.collectionName)
	destColl := f.client.Collection(fmt.Sprintf("%s_archive", f.collectionName))

	query := sourceColl.Where("EventSourceID", "==", sid)
	iter := query.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		// アーカイブ用ドキュメントに詰め替える
		var target document
		if err := doc.DataTo(&target); err != nil {
			return err
		}
		archived := archivedDocument{
			EventSourceID: target.EventSourceID,
			Input:         target.Input,
			Reply:         target.Reply,
			Timestamp:     target.Timestamp,
			ArchivedAt:    time.Now().In(jst),
		}

		// アーカイブ先コレクションに追加
		if _, _, err = destColl.Add(ctx, archived); err != nil {
			return err
		}

		// 対象のドキュメントをコレクションから削除
		if _, err = doc.Ref.Delete(ctx); err != nil {
			return err
		}
	}

	return nil
}

type usageDocument struct {
	EventSourceID    string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Timestamp        time.Time
}

func (f *FirestoreRepository) Add(ctx context.Context, sid EventSourceID, usage APIUsage) error {
	slog.Debug("Print api usage", "usage", usage)

	coll := fmt.Sprintf("%s_usage", f.collectionName)
	doc := usageDocument{
		EventSourceID:    string(sid),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		Timestamp:        time.Now().In(jst),
	}

	_, _, err := f.client.Collection(coll).Add(ctx, doc)
	return err
}

func (f *FirestoreRepository) FetchTotalTokens(ctx context.Context, sid EventSourceID) (int, error) {
	now := time.Now().In(jst)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
	todayEnd := todayStart.Add(24 * time.Hour)

	coll := f.client.Collection(fmt.Sprintf("%s_usage", f.collectionName))
	query := coll.Where("EventSourceID", "==", sid).
		Where("Timestamp", ">=", todayStart).
		Where("Timestamp", "<", todayEnd)

	iter := query.Documents(ctx)
	defer iter.Stop()

	sum := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		var usage usageDocument
		if err := doc.DataTo(&usage); err != nil {
			return 0, err
		}
		sum += usage.TotalTokens
	}

	return sum, nil
}

func (f *FirestoreRepository) Close() error {
	slog.Debug("Firestore client closed")
	return f.client.Close()
}
