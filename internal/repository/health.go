package repository

import (
	"context"
	"database/sql"
)

type HealthRepository struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) *HealthRepository {
	return &HealthRepository{
		db: db,
	}
}

func (r *HealthRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
