package main

import (
	"net/http"

	"github.com/caarlos0/env/v7"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/exp/slog"
)

type linebotConfig struct {
	ChannelSecret      string `env:"LINE_CHANNEL_SECRET,required"`
	ChannelAccessToken string `env:"LINE_CHANNEL_ACCESS_TOKEN,required"`
}

type Linebot struct {
	config linebotConfig
	client *linebot.Client
	app    *ApplicationService
}

func NewLinebot(app *ApplicationService) (*Linebot, error) {
	cfg := linebotConfig{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env", err)
		return nil, ErrParseConfig
	}
	client, err := linebot.New(cfg.ChannelSecret, cfg.ChannelAccessToken)
	if err != nil {
		slog.Error("Failed to create linebot", err)
		return nil, err
	}

	return &Linebot{
		config: cfg,
		client: client,
		app:    app,
	}, nil
}

func (l *Linebot) Handler(w http.ResponseWriter, req *http.Request) {
	// リクエストを受信するのにわざわざ *linebot.Client を使っているのは、署名を検証するため（それ以外には使っていない）
	events, err := l.client.ParseRequest(req)
	if err != nil {
		slog.Error("ParseRequestError", err)
		return
	}
	if len(events) == 0 {
		slog.Error("no events")
		return
	}

	event := events[0]
	slog.Debug("Print event", "event", event)

	if event.Type == linebot.EventTypeUnfollow {
		slog.Info("Unfollowed by", "EventSourceID", getEventSourceID(event))
	}

	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		reply := l.app.ReplyWithHistory(message.Text, getEventSourceID(event))
		l.replyTextMessage(event.ReplyToken, reply)
	}
}

func (l *Linebot) replyTextMessage(token, message string) {
	if _, err := l.client.ReplyMessage(token, linebot.NewTextMessage(message)).Do(); err != nil {
		slog.Error("ReplyMessageError", err)
		return
	}
}

func getEventSourceID(e *linebot.Event) EventSourceID {
	source := e.Source
	switch source.Type {
	case linebot.EventSourceTypeUser:
		return EventSourceID(source.UserID)
	case linebot.EventSourceTypeGroup:
		return EventSourceID(source.GroupID)
	case linebot.EventSourceTypeRoom:
		return EventSourceID(source.RoomID)
	}
	return EventSourceID(source.UserID)
}
