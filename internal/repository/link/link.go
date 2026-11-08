package link

import (
	"context"
	"errors"

	"github.com/rucodencode/shortener/internal/model"
)

var (
	ErrNotFound = errors.New("link not found")
)

type LinkRepository interface {
	Save(ctx context.Context, link *model.Link) error
	FindByCode(ctx context.Context, code string) (*model.Link, error)
	FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error)
	SaveBatch(ctx context.Context, links []*model.Link) error
	FindByUserID(ctx context.Context, userID string) ([]*model.Link, error)
}
