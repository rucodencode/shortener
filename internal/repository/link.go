package repository

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/rucodencode/shortener/internal/model"
)

type LinkRepository struct {
	mu              sync.Mutex
	links           map[string]model.Link
	codesByURL      map[string]string
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

	codesByURL := make(map[string]string)
	for code, value := range links {
		codesByURL[value.OriginalURL] = code
	}

	return &LinkRepository{
		links:           links,
		codesByURL:      codesByURL,
		fileStoragePath: fileStoragePath,
	}, nil
}

func (r *LinkRepository) Save(link model.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	r.links[link.Code] = link
	r.codesByURL[link.OriginalURL] = link.Code

	if err = json.NewEncoder(file).Encode(r.links); err != nil {
		delete(r.links, link.Code)
		delete(r.codesByURL, link.OriginalURL)
		return err
	}

	return nil
}

func (r *LinkRepository) FindByCode(code string) (model.Link, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, ok := r.links[code]
	return link, ok
}

func (r *LinkRepository) FindByOriginalURL(originalURL string) (model.Link, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	code, ok := r.codesByURL[originalURL]
	if !ok {
		return model.Link{}, false
	}

	link, ok := r.links[code]
	return link, ok
}
