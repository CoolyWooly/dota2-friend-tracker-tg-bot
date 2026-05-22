package telegram

import (
	"context"
	"log/slog"
	"runtime/debug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	tgservice "github.com/yerlan/dota2/internal/service/telegram"
)

type BotUseCaseI interface {
	HandleUpdate(ctx context.Context, upd tgbotapi.Update)
}

type Listener struct {
	bot *tgservice.Bot
	uc  BotUseCaseI
}

func New(bot *tgservice.Bot, uc BotUseCaseI) *Listener {
	return &Listener{bot: bot, uc: uc}
}

func (l *Listener) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := l.bot.API().GetUpdatesChan(u)
	slog.Info("telegram: listening for updates")

	for {
		select {
		case <-ctx.Done():
			l.bot.StopReceivingUpdates()
			slog.Info("telegram: stopped")
			return
		case upd, ok := <-updates:
			if !ok {
				slog.Warn("telegram: updates channel closed")
				return
			}
			go l.handle(ctx, upd)
		}
	}
}

func (l *Listener) handle(ctx context.Context, upd tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("telegram: panic in handler", "panic", r, "stack", string(debug.Stack()))
		}
	}()
	l.uc.HandleUpdate(ctx, upd)
}
