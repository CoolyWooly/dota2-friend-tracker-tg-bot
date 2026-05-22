package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	commonpg "github.com/yerlan/dota2/internal/domain/common/repo/pg"
)

type Repo struct {
	*commonpg.Base
}

func New(con *pgxpool.Pool) *Repo {
	return &Repo{Base: commonpg.NewBase(con)}
}

func (r *Repo) Mark(ctx context.Context, accountID, matchID int64) (bool, error) {
	const q = `
		INSERT INTO sent_notifications (account_id, match_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	tag, err := r.Con.Exec(ctx, q, accountID, matchID)
	if err != nil {
		return false, fmt.Errorf("exec: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
