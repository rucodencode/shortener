package link

import (
	"context"

	"github.com/rucodencode/shortener/internal/model"
)

type LinkRepository interface {
	Save(ctx context.Context, link *model.Link) error
	FindByCode(ctx context.Context, code string) (*model.Link, bool)
	FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, bool)
}
