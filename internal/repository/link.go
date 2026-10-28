package repository

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/rucodencode/shortener/internal/model"
)

type LinkRepository struct {
	links           map[string]model.Link
	fileStoragePath string
}

func NewLinkRepository(fileStoragePath string) (*LinkRepository, error) {
	links := make(map[string]model.Link)
	file, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&links); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return &LinkRepository{
		links:           links,
		fileStoragePath: fileStoragePath,
	}, nil
}

func (r *LinkRepository) Save(link model.Link) error {
	file, err := os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	r.links[link.Code] = link
	if err = json.NewEncoder(file).Encode(r.links); err != nil {
		delete(r.links, link.Code)
		return err
	}

	return nil
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
