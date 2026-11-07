package link

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/rucodencode/shortener/internal/model"
)

type FileRepository struct {
	mu              sync.RWMutex
	links           map[string]*model.Link
	codesByURL      map[string]string
	fileStoragePath string
}

func NewFileRepository(fileStoragePath string) (LinkRepository, error) {
	links, err := loadDataFromDisk(fileStoragePath)
	if err != nil {
		return nil, err
	}

	codesByURL := make(map[string]string)
	for code, value := range links {
		if value == nil {
			continue
		}

		codesByURL[value.OriginalURL] = code
	}

	return &FileRepository{
		mu:              sync.RWMutex{},
		links:           links,
		codesByURL:      codesByURL,
		fileStoragePath: fileStoragePath,
	}, nil
}

func (r *FileRepository) Save(ctx context.Context, link *model.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.links[link.Code] = link
	r.codesByURL[link.OriginalURL] = link.Code

	if err := r.persistToDisk(); err != nil {
		delete(r.links, link.Code)
		delete(r.codesByURL, link.OriginalURL)
		return err
	}

	return nil
}

func (r *FileRepository) FindByCode(ctx context.Context, code string) (*model.Link, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.links[code]
	return link, ok
}

func (r *FileRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.Link, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	code, ok := r.codesByURL[originalURL]
	if !ok {
		return nil, false
	}

	link, ok := r.links[code]
	return link, ok
}

func (r *FileRepository) persistToDisk() error {
	file, err := os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(r.links)
}

func loadDataFromDisk(fileStoragePath string) (map[string]*model.Link, error) {
	links := make(map[string]*model.Link)
	file, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&links); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return links, nil
}
