package service

import (
	"context"
	"fmt"

	"github.com/yerlan/dota2/internal/domain/account/model"
)

type RepoDbI interface {
	Add(ctx context.Context, m *model.Main) error
	Remove(ctx context.Context, accountID int64) (bool, error)
	Get(ctx context.Context, accountID int64) (*model.Main, error)
	List(ctx context.Context) ([]*model.Main, error)
	ListIDs(ctx context.Context) ([]int64, error)
}

type Service struct {
	repo RepoDbI
}

func New(repo RepoDbI) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, accountID int64, displayName string) error {
	if err := s.repo.Add(ctx, &model.Main{AccountID: accountID, DisplayName: displayName}); err != nil {
		return fmt.Errorf("repo.Add: %w", err)
	}
	return nil
}

func (s *Service) Remove(ctx context.Context, accountID int64) (bool, error) {
	return s.repo.Remove(ctx, accountID)
}

func (s *Service) Get(ctx context.Context, accountID int64) (*model.Main, error) {
	return s.repo.Get(ctx, accountID)
}

func (s *Service) List(ctx context.Context) ([]*model.Main, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListIDs(ctx context.Context) ([]int64, error) {
	return s.repo.ListIDs(ctx)
}
