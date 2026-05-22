package service

import (
	"context"
	"time"

	"github.com/yerlan/dota2/internal/domain/pollstate/model"
)

type RepoDbI interface {
	Get(ctx context.Context, accountID int64) (*model.Main, error)
	Upsert(ctx context.Context, m *model.Main) error
}

type Service struct {
	repo RepoDbI
}

func New(repo RepoDbI) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, accountID int64) (*model.Main, error) {
	return s.repo.Get(ctx, accountID)
}

func (s *Service) Mark(ctx context.Context, accountID, lastMatchID int64) error {
	now := time.Now()
	return s.repo.Upsert(ctx, &model.Main{
		AccountID:    accountID,
		LastMatchID:  &lastMatchID,
		LastPolledAt: &now,
	})
}

func (s *Service) Touch(ctx context.Context, accountID int64) error {
	now := time.Now()
	return s.repo.Upsert(ctx, &model.Main{
		AccountID:    accountID,
		LastPolledAt: &now,
	})
}
