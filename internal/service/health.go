package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthService struct {
	pool *pgxpool.Pool
}

func NewHealthService(pool *pgxpool.Pool) *HealthService {
	return &HealthService{
		pool: pool,
	}
}

func (s *HealthService) PingDB(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}

	return s.pool.Ping(ctx)
}
