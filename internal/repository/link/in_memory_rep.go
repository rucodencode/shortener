package link

import (
	"context"
	"sync"

	"github.com/rucodencode/shortener/internal/model"
)

type InMemoryRepository struct {
	mu         sync.RWMutex
	links      map[string]*model.Link
	codesByURL map[string]string
}

func NewInMemoryRepository() (LinkRepository, error) {
	return &InMemoryRepository{
		mu:         sync.RWMutex{},
		links:      make(map[string]*model.Link),
		codesByURL: make(map[string]string),
	}, nil
}

func (r *InMemoryRepository) Save(ctx context.Context, link *model.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.links[link.Code] = link
	r.codesByURL[link.OriginalURL] = link.Code

	return nil
}

func (r *InMemoryRepository) FindByCode(ctx context.Context, code string) (*model.Link, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.links[code]
	return link, ok
}

func (r *InMemoryRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	code, ok := r.codesByURL[originalURL]
	if !ok {
		return nil, false
	}

	link, ok := r.links[code]
	return link, ok
}
