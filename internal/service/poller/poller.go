package poller

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"time"

	pollstatemodel "github.com/yerlan/dota2/internal/domain/pollstate/model"
	"github.com/yerlan/dota2/internal/service/opendota"
)

type AccountSvcI interface {
	ListIDs(ctx context.Context) ([]int64, error)
}

type PollStateSvcI interface {
	Get(ctx context.Context, accountID int64) (*pollstatemodel.Main, error)
	Mark(ctx context.Context, accountID, lastMatchID int64) error
	Touch(ctx context.Context, accountID int64) error
}

type OpenDotaI interface {
	GetRecentMatches(ctx context.Context, accountID int64) ([]*opendota.RecentMatch, error)
}

type MatchUseCaseI interface {
	HandleNewMatch(ctx context.Context, accountID int64, m *opendota.RecentMatch) error
}

type Poller struct {
	interval time.Duration

	accounts AccountSvcI
	state    PollStateSvcI
	opendota OpenDotaI
	matchUC  MatchUseCaseI
}

func New(interval time.Duration, accounts AccountSvcI, state PollStateSvcI, od OpenDotaI, matchUC MatchUseCaseI) *Poller {
	return &Poller{
		interval: interval,
		accounts: accounts,
		state:    state,
		opendota: od,
		matchUC:  matchUC,
	}
}

func (p *Poller) Run(ctx context.Context) {
	slog.Info("poller: starting", "interval", p.interval)
	t := time.NewTicker(p.interval)
	defer t.Stop()

	p.safeTick(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("poller: stopping")
			return
		case <-t.C:
			p.safeTick(ctx)
		}
	}
}

func (p *Poller) safeTick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("poller: panic in tick", "panic", r, "stack", string(debug.Stack()))
		}
	}()
	p.tick(ctx)
}

func (p *Poller) tick(ctx context.Context) {
	ids, err := p.accounts.ListIDs(ctx)
	if err != nil {
		slog.Error("poller: list accounts", "error", err)
		return
	}
	slog.Info("poller: tick", "accounts", len(ids))
	if len(ids) == 0 {
		return
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}
		if err := p.pollAccount(ctx, id); err != nil {
			slog.Warn("poller: account failed", "account_id", id, "error", err)
		}
	}
}

func (p *Poller) pollAccount(ctx context.Context, accountID int64) error {
	state, err := p.state.Get(ctx, accountID)
	if err != nil {
		return err
	}

	matches, err := p.opendota.GetRecentMatches(ctx, accountID)
	if err != nil {
		if errors.Is(err, opendota.ErrNotFound) {
			_ = p.state.Touch(ctx, accountID)
			return nil
		}
		return err
	}
	if len(matches) == 0 {
		return p.state.Touch(ctx, accountID)
	}

	newest := matches[0]

	// первый прогон — фиксируем точку и ничего не отправляем
	if state == nil || state.LastMatchID == nil {
		return p.state.Mark(ctx, accountID, newest.MatchID)
	}

	last := *state.LastMatchID
	if newest.MatchID == last {
		return p.state.Touch(ctx, accountID)
	}

	var fresh []*opendota.RecentMatch
	for _, m := range matches {
		if m.MatchID == last {
			break
		}
		fresh = append(fresh, m)
	}
	for i := len(fresh) - 1; i >= 0; i-- {
		if err := p.matchUC.HandleNewMatch(ctx, accountID, fresh[i]); err != nil {
			slog.Error("poller: handle match", "account_id", accountID, "match_id", fresh[i].MatchID, "error", err)
		}
	}

	return p.state.Mark(ctx, accountID, newest.MatchID)
}
