package service

import (
	"context"
	"database/sql"
)

type HealthService struct {
	db *sql.DB
}

func NewHealthService(db *sql.DB) *HealthService {
	return &HealthService{
		db: db,
	}
}

func (s *HealthService) PingDB(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	return s.db.PingContext(ctx)
}
