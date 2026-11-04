package link

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	sqlText := "INSERT INTO links(code, original_url) VALUES($1, $2)"
	_, err := r.db.ExecContext(ctx, sqlText, link.Code, link.OriginalURL)
	if err != nil {
		return fmt.Errorf("save link: %w", err)
	}

	return nil
}

func (r *PostgresRepository) SaveBatch(ctx context.Context, links []*model.Link) error {
	if len(links) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(links))
	args := make([]any, 0, len(links)*2)

	for i, link := range links {
		firstParam := i*2 + 1
		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d)", firstParam, firstParam+1),
		)

		args = append(args, link.Code, link.OriginalURL)
	}

	sqlText := "INSERT INTO links(code, original_url) VALUES " +
		strings.Join(placeholders, ", ")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin batch transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, sqlText, args...); err != nil {
		return fmt.Errorf("save batch links: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit batch transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) FindByCode(ctx context.Context, code string) (*model.Link, error) {
	link := &model.Link{}

	sqlText := "SELECT code, original_url FROM links WHERE code = $1"
	row := r.db.QueryRowContext(ctx, sqlText, code)

	err := row.Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("find link by code: %w", err)
	}

	return link, nil
}

func (r *PostgresRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error) {
	link := &model.Link{}

	sqlText := "SELECT code, original_url FROM links WHERE original_url = $1"
	row := r.db.QueryRowContext(ctx, sqlText, originalURL)
	err := row.Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("find link by originalURL: %w", err)
	}

	return link, nil
}
