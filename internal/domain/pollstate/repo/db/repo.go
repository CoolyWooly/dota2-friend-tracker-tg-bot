package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonpg "github.com/yerlan/dota2/internal/domain/common/repo/pg"
	"github.com/yerlan/dota2/internal/domain/pollstate/model"
)

type Repo struct {
	*commonpg.Base
}

func New(con *pgxpool.Pool) *Repo {
	return &Repo{Base: commonpg.NewBase(con)}
}

func (r *Repo) Get(ctx context.Context, accountID int64) (*model.Main, error) {
	var m model.Main
	err := r.Con.QueryRow(ctx, `
		SELECT account_id, last_match_id, last_polled_at
		FROM poll_state WHERE account_id = $1`, accountID,
	).Scan(&m.AccountID, &m.LastMatchID, &m.LastPolledAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query: %w", err)
	}
	return &m, nil
}

func (r *Repo) Upsert(ctx context.Context, m *model.Main) error {
	const q = `
		INSERT INTO poll_state (account_id, last_match_id, last_polled_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id) DO UPDATE SET
		    last_match_id  = COALESCE(EXCLUDED.last_match_id, poll_state.last_match_id),
		    last_polled_at = COALESCE(EXCLUDED.last_polled_at, poll_state.last_polled_at)
	`
	if _, err := r.Con.Exec(ctx, q, m.AccountID, m.LastMatchID, m.LastPolledAt); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
