package bot

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	accountmodel "github.com/yerlan/dota2/internal/domain/account/model"
	"github.com/yerlan/dota2/internal/service/opendota"
	"github.com/yerlan/dota2/internal/service/telegram"
)

type AccountSvcI interface {
	Add(ctx context.Context, accountID int64, displayName string) error
	Remove(ctx context.Context, accountID int64) (bool, error)
	List(ctx context.Context) ([]*accountmodel.Main, error)
}

type OpenDotaI interface {
	GetPlayer(ctx context.Context, accountID int64) (*opendota.Player, error)
}

type SenderI interface {
	Send(chatID int64, text string) error
}

type UseCase struct {
	ownerID      int64
	pollInterval time.Duration
	accounts     AccountSvcI
	opendota     OpenDotaI
	sender       SenderI
}

func New(ownerID int64, pollInterval time.Duration, accounts AccountSvcI, od OpenDotaI, sender SenderI) *UseCase {
	return &UseCase{
		ownerID:      ownerID,
		pollInterval: pollInterval,
		accounts:     accounts,
		opendota:     od,
		sender:       sender,
	}
}

func (uc *UseCase) HandleUpdate(ctx context.Context, upd tgbotapi.Update) {
	msg := upd.Message
	if msg == nil || msg.Text == "" {
		return
	}
	chatID := msg.Chat.ID
	if chatID != uc.ownerID {
		slog.Info("bot: ignored non-owner", "from", chatID, "text", msg.Text)
		return
	}

	text := strings.TrimSpace(msg.Text)
	cmd, arg := splitCommand(text)
	switch cmd {
	case "start", "help":
		uc.cmdStart(chatID)
	case "add":
		uc.cmdAdd(ctx, chatID, arg)
	case "remove":
		uc.cmdRemove(ctx, chatID, arg)
	case "list":
		uc.cmdList(ctx, chatID)
	case "test":
		uc.cmdTest(chatID)
	default:
		uc.reply(chatID, "Неизвестная команда. /help — список команд.")
	}
}

func (uc *UseCase) cmdStart(chatID int64) {
	uc.reply(chatID, fmt.Sprintf(`Я слежу за матчами в Dota 2 и присылаю карточки.

<b>Команды:</b>
/add &lt;account_id&gt; [имя] — добавить отслеживаемый аккаунт
/remove &lt;account_id&gt; — убрать
/list — список отслеживаемых
/test — тестовое уведомление

Интервал опроса OpenDota: <b>%s</b>.`, uc.pollInterval))
}

func (uc *UseCase) cmdAdd(ctx context.Context, chatID int64, arg string) {
	parts := strings.SplitN(arg, " ", 2)
	id, ok := parseAccountID(parts[0])
	if !ok {
		uc.reply(chatID, "Использование: /add &lt;account_id&gt; [имя]")
		return
	}
	name := ""
	if len(parts) == 2 {
		name = strings.TrimSpace(parts[1])
	}

	player, err := uc.opendota.GetPlayer(ctx, id)
	if err != nil && !errors.Is(err, opendota.ErrNotFound) {
		slog.Error("add: GetPlayer", "error", err)
		uc.reply(chatID, "OpenDota недоступен.")
		return
	}
	if name == "" && player != nil {
		name = player.Profile.Personaname
	}
	if name == "" {
		name = strconv.FormatInt(id, 10)
	}

	if player == nil || player.Profile.AccountID == 0 {
		uc.reply(chatID, "⚠️ Профиль приватный или не существует. Уведомления будут возможны только при включённом <b>Expose Public Match Data</b>.")
	}

	if err := uc.accounts.Add(ctx, id, name); err != nil {
		slog.Error("add: Add", "error", err)
		uc.reply(chatID, "Не удалось добавить.")
		return
	}
	uc.reply(chatID, fmt.Sprintf("Добавлен: <b>%s</b> (id=%d)", html.EscapeString(name), id))
}

func (uc *UseCase) cmdRemove(ctx context.Context, chatID int64, arg string) {
	id, ok := parseAccountID(arg)
	if !ok {
		uc.reply(chatID, "Использование: /remove &lt;account_id&gt;")
		return
	}
	removed, err := uc.accounts.Remove(ctx, id)
	if err != nil {
		slog.Error("remove", "error", err)
		uc.reply(chatID, "Не удалось убрать.")
		return
	}
	if !removed {
		uc.reply(chatID, "Такого аккаунта в списке нет.")
		return
	}
	uc.reply(chatID, "Удалён.")
}

func (uc *UseCase) cmdList(ctx context.Context, chatID int64) {
	list, err := uc.accounts.List(ctx)
	if err != nil {
		slog.Error("list", "error", err)
		uc.reply(chatID, "Не удалось получить список.")
		return
	}
	if len(list) == 0 {
		uc.reply(chatID, "Список пуст. /add &lt;account_id&gt; [имя]")
		return
	}
	var b strings.Builder
	b.WriteString("<b>Отслеживаемые аккаунты</b>\n")
	for _, a := range list {
		fmt.Fprintf(&b, "• <b>%s</b> — <code>%d</code>\n", html.EscapeString(a.DisplayName), a.AccountID)
	}
	uc.reply(chatID, b.String())
}

func (uc *UseCase) cmdTest(chatID int64) {
	fake := &opendota.RecentMatch{
		MatchID:    1234567890,
		HeroID:     1,
		StartTime:  0,
		Duration:   2415,
		GameMode:   22,
		LobbyType:  7,
		PlayerSlot: 0,
		RadiantWin: true,
		Kills:      12,
		Deaths:     4,
		Assists:    8,
		LastHits:   210,
		Denies:     14,
		GoldPerMin: 612,
		XPPerMin:   715,
	}
	text := telegram.FormatMatchCard(fake, "Anti-Mage", "тестовый аккаунт")
	uc.reply(chatID, "<i>Тестовое уведомление:</i>")
	uc.reply(chatID, text)
}

func (uc *UseCase) reply(chatID int64, text string) {
	if err := uc.sender.Send(chatID, text); err != nil {
		slog.Error("bot: send", "chat", chatID, "error", err)
	}
}

func parseAccountID(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func splitCommand(text string) (cmd, arg string) {
	if !strings.HasPrefix(text, "/") {
		return "", text
	}
	text = strings.TrimPrefix(text, "/")
	idx := strings.IndexAny(text, " ")
	if idx < 0 {
		return strings.ToLower(stripBot(text)), ""
	}
	return strings.ToLower(stripBot(text[:idx])), strings.TrimSpace(text[idx+1:])
}

func stripBot(cmd string) string {
	if i := strings.Index(cmd, "@"); i >= 0 {
		return cmd[:i]
	}
	return cmd
}
