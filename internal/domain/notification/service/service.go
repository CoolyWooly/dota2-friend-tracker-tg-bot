package service

import "context"

type RepoDbI interface {
	Mark(ctx context.Context, accountID, matchID int64) (inserted bool, err error)
}

type Service struct {
	repo RepoDbI
}

func New(repo RepoDbI) *Service {
	return &Service{repo: repo}
}

// MarkIfAbsent атомарно вставляет запись об отправленном уведомлении и
// возвращает true, если уведомление ещё не отправлялось.
func (s *Service) MarkIfAbsent(ctx context.Context, accountID, matchID int64) (bool, error) {
	return s.repo.Mark(ctx, accountID, matchID)
}
