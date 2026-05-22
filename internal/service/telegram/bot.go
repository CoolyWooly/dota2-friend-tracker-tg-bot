package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func New(token string, debug bool) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}
	api.Debug = debug
	slog.Info("telegram bot authorized", "username", api.Self.UserName)
	return &Bot{api: api}, nil
}

func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

func (b *Bot) Send(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (b *Bot) Reply(chatID int64, replyTo int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	msg.ReplyToMessageID = replyTo
	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (b *Bot) StopReceivingUpdates() {
	b.api.StopReceivingUpdates()
}
