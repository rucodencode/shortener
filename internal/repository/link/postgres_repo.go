package link

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rucodencode/shortener/internal/model"
)

var ErrOriginalURLConflict = errors.New("original URL already exists")

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) (LinkRepository, error) {
	return &PostgresRepository{
		pool: pool,
	}, nil
}

func isUniqueViolationOn(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == pgerrcode.UniqueViolation &&
		pgErr.ConstraintName == constraint
}

func classifyWriteError(err error) error {
	if isUniqueViolationOn(err, "links_original_url_key") {
		return fmt.Errorf("%w: %w", ErrOriginalURLConflict, err)
	}
	return err
}

func IsCodeConflict(err error) bool {
	return isUniqueViolationOn(err, "links_code_key")
}

func (r *PostgresRepository) Save(ctx context.Context, link *model.Link) error {
	sqlText := "INSERT INTO links(user_id, code, original_url) VALUES($1, $2, $3)"
	_, err := r.pool.Exec(ctx, sqlText, link.UserID, link.Code, link.OriginalURL)
	if err != nil {
		return fmt.Errorf("save link: %w", classifyWriteError(err))
	}

	return nil
}

func (r *PostgresRepository) SaveBatch(ctx context.Context, links []*model.Link) error {
	if len(links) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(links))
	args := make([]any, 0, len(links)*3)

	for i, link := range links {
		firstParam := i*3 + 1
		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d, $%d)", firstParam, firstParam+1, firstParam+2),
		)

		args = append(args, link.UserID, link.Code, link.OriginalURL)
	}

	sqlText := "INSERT INTO links(user_id, code, original_url) VALUES " +
		strings.Join(placeholders, ", ")

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin batch transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sqlText, args...); err != nil {
		return fmt.Errorf("save batch links: %w", classifyWriteError(err))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit batch transaction: %w", classifyWriteError(err))
	}

	return nil
}

func (r *PostgresRepository) FindByCode(ctx context.Context, code string) (*model.Link, error) {
	link := &model.Link{}

	sqlText := "SELECT user_id, code, original_url FROM links WHERE code = $1"
	row := r.pool.QueryRow(ctx, sqlText, code)

	err := row.Scan(&link.UserID, &link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("find link by code: %w", err)
	}

	return link, nil
}

func (r *PostgresRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error) {
	link := &model.Link{}

	sqlText := "SELECT user_id, code, original_url FROM links WHERE original_url = $1"
	row := r.pool.QueryRow(ctx, sqlText, originalURL)
	err := row.Scan(&link.UserID, &link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("find link by originalURL: %w", err)
	}

	return link, nil
}

func (r *PostgresRepository) FindByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	links := []*model.Link{}

	sqlText := "SELECT user_id, code, original_url FROM links WHERE user_id = $1"
	rows, err := r.pool.Query(ctx, sqlText, userID)
	if err != nil {
		return nil, fmt.Errorf("find link by userID: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		link := &model.Link{}
		err := rows.Scan(&link.UserID, &link.Code, &link.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("find link by userID: %w", err)
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("find link by userID: %w", err)
	}

	return links, nil
}
