package service

import "context"

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	repo HealthRepository
}

func NewHealthService(repo HealthRepository) *HealthService {
	return &HealthService{
		repo: repo,
	}
}

func (s *HealthService) PingDB(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
