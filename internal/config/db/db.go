package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	DataSource string
}

func NewDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DataSource)
	if err != nil {
		return nil, err
	}

	return db, nil
}
