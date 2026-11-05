package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rucodencode/shortener/migrations"
)

type Config struct {
	DataSource string
}

func NewDB(cfg Config) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DataSource)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	return pool, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dbConn := stdlib.OpenDBFromPool(pool)
	defer dbConn.Close()

	sourceDriver, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("create migrations source: %w", err)
	}

	conn, err := dbConn.Conn(ctx)
	if err != nil {
		sourceDriver.Close()
		return fmt.Errorf("get migration connection: %w", err)
	}

	databaseDriver, err := postgres.WithConnection(ctx, conn, &postgres.Config{})
	if err != nil {
		conn.Close()
		sourceDriver.Close()
		return fmt.Errorf("create migration database driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseDriver)
	if err != nil {
		databaseDriver.Close()
		sourceDriver.Close()
		return fmt.Errorf("create migrator: %w", err)
	}

	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
