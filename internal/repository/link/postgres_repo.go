package link

import (
	"context"
	"database/sql"

	"github.com/rucodencode/shortener/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (LinkRepository, error) {
	return &PostgresRepository{
		db: db,
	}, nil
}

func (r *PostgresRepository) Save(ctx context.Context, link *model.Link) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO links(code, original_url) VALUES($1, $2)", link.Code, link.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) FindByCode(ctx context.Context, code string) (*model.Link, bool) {
	link := &model.Link{}

	row := r.db.QueryRowContext(ctx, "SELECT code, original_url FROM links WHERE code = ?", code)
	err := row.Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		return nil, false
	}

	err = row.Err()
	if err != nil {
		return nil, false
	}

	return link, true
}

func (r *PostgresRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, bool) {
	link := &model.Link{}

	row := r.db.QueryRowContext(ctx, "SELECT code, original_url FROM links WHERE original_url = ?", originalURL)
	err := row.Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		return nil, false
	}

	err = row.Err()
	if err != nil {
		return nil, false
	}

	return link, true
}
