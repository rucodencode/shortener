package service

import (
	"context"
	"database/sql"
	"errors"
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
		return errors.New("database is not configured")
	}

	return s.db.PingContext(ctx)
}
