package match

import (
	"context"
	"fmt"
	"log/slog"

	accountmodel "github.com/yerlan/dota2/internal/domain/account/model"
	"github.com/yerlan/dota2/internal/service/opendota"
	"github.com/yerlan/dota2/internal/service/telegram"
)

type AccountSvcI interface {
	Get(ctx context.Context, accountID int64) (*accountmodel.Main, error)
}

type NotificationSvcI interface {
	MarkIfAbsent(ctx context.Context, accountID, matchID int64) (bool, error)
}

type HeroNamerI interface {
	Name(ctx context.Context, heroID int) string
}

type OpenDotaI interface {
	GetPlayer(ctx context.Context, accountID int64) (*opendota.Player, error)
}

type SenderI interface {
	Send(chatID int64, text string) error
}

type UseCase struct {
	ownerID   int64
	accounts  AccountSvcI
	notifs    NotificationSvcI
	heroNamer HeroNamerI
	opendota  OpenDotaI
	sender    SenderI
}

func New(ownerID int64, accounts AccountSvcI, notifs NotificationSvcI, heroNamer HeroNamerI, od OpenDotaI, sender SenderI) *UseCase {
	return &UseCase{
		ownerID:   ownerID,
		accounts:  accounts,
		notifs:    notifs,
		heroNamer: heroNamer,
		opendota:  od,
		sender:    sender,
	}
}

func (uc *UseCase) HandleNewMatch(ctx context.Context, accountID int64, m *opendota.RecentMatch) error {
	fresh, err := uc.notifs.MarkIfAbsent(ctx, accountID, m.MatchID)
	if err != nil {
		return fmt.Errorf("MarkIfAbsent: %w", err)
	}
	if !fresh {
		return nil
	}

	var displayName, personaname string
	if acc, err := uc.accounts.Get(ctx, accountID); err != nil {
		slog.Warn("match: get account", "account_id", accountID, "error", err)
	} else if acc != nil {
		displayName = acc.DisplayName
	}
	if player, err := uc.opendota.GetPlayer(ctx, accountID); err != nil {
		slog.Warn("match: get player", "account_id", accountID, "error", err)
	} else if player != nil {
		personaname = player.Profile.Personaname
	}

	label := buildLabel(displayName, personaname, accountID)

	heroName := uc.heroNamer.Name(ctx, m.HeroID)
	text := telegram.FormatMatchCard(m, heroName, label)
	if err := uc.sender.Send(uc.ownerID, text); err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func buildLabel(displayName, personaname string, accountID int64) string {
	switch {
	case displayName != "" && personaname != "" && displayName != personaname:
		return fmt.Sprintf("%s (%s)", displayName, personaname)
	case personaname != "":
		return personaname
	case displayName != "":
		return displayName
	default:
		return fmt.Sprintf("id %d", accountID)
	}
}
