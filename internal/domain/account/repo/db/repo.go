package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yerlan/dota2/internal/domain/account/model"
	commonpg "github.com/yerlan/dota2/internal/domain/common/repo/pg"
)

type Repo struct {
	*commonpg.Base
}

func New(con *pgxpool.Pool) *Repo {
	return &Repo{Base: commonpg.NewBase(con)}
}

func (r *Repo) Add(ctx context.Context, m *model.Main) error {
	const q = `
		INSERT INTO tracked_accounts (account_id, display_name) VALUES ($1, $2)
		ON CONFLICT (account_id) DO UPDATE SET display_name = EXCLUDED.display_name
	`
	if _, err := r.Con.Exec(ctx, q, m.AccountID, m.DisplayName); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (r *Repo) Remove(ctx context.Context, accountID int64) (bool, error) {
	tag, err := r.Con.Exec(ctx, `DELETE FROM tracked_accounts WHERE account_id = $1`, accountID)
	if err != nil {
		return false, fmt.Errorf("exec: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repo) Get(ctx context.Context, accountID int64) (*model.Main, error) {
	var m model.Main
	err := r.Con.QueryRow(ctx,
		`SELECT account_id, display_name, created_at FROM tracked_accounts WHERE account_id = $1`,
		accountID,
	).Scan(&m.AccountID, &m.DisplayName, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query: %w", err)
	}
	return &m, nil
}

func (r *Repo) List(ctx context.Context) ([]*model.Main, error) {
	rows, err := r.Con.Query(ctx,
		`SELECT account_id, display_name, created_at FROM tracked_accounts ORDER BY created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var res []*model.Main
	for rows.Next() {
		var m model.Main
		if err := rows.Scan(&m.AccountID, &m.DisplayName, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, &m)
	}
	return res, rows.Err()
}

func (r *Repo) ListIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.Con.Query(ctx, `SELECT account_id FROM tracked_accounts`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
