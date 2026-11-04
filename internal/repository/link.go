package repository

import (
	"github.com/rucodencode/shortener/internal/model"
)

type LinkRepository struct {
	links map[string]model.Link
}

func NewLinkRepository() *LinkRepository {
	return &LinkRepository{
		links: make(map[string]model.Link),
	}
}

func (r *LinkRepository) Save(link model.Link) {
	r.links[link.Code] = link
}

func (r *LinkRepository) FindByCode(code string) (model.Link, bool) {
	link, ok := r.links[code]
	return link, ok
}

func (r *LinkRepository) FindByOriginalURL(originalURL string) (model.Link, bool) {
	for _, link := range r.links {
		if link.OriginalURL == originalURL {
			return link, true
		}
	}

	return model.Link{}, false
}
