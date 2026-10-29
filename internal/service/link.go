package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/rucodencode/shortener/internal/model"
)

var (
	ErrCodeGeneration = errors.New("failed to generate unique code")
	ErrNotFound       = errors.New("not found")
	ErrInvalidURL     = errors.New("invalid url")
)

type LinkRepository interface {
	Save(link model.Link) error
	FindByCode(code string) (model.Link, bool)
	FindByOriginalURL(originalURL string) (model.Link, bool)
}

type LinkService struct {
	mu   sync.Mutex
	repo LinkRepository
}

func NewLinkService(repo LinkRepository) *LinkService {
	return &LinkService{
		repo: repo,
	}
}

func (s *LinkService) Create(originalURL string) (model.Link, error) {
	if err := validateURL(originalURL); err != nil {
		return model.Link{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	link, found := s.repo.FindByOriginalURL(originalURL)
	if found {
		return link, nil
	}

	const maxAttempts = 100

	for attempt := 0; attempt < maxAttempts; attempt++ {
		source := originalURL

		if attempt > 0 {
			source = fmt.Sprintf("%s:%d", originalURL, attempt)
		}

		code := generateCode(source)

		existing, found := s.repo.FindByCode(code)
		if found {
			if existing.OriginalURL == originalURL {
				return existing, nil
			}

			continue
		}

		link := model.Link{
			Code:        code,
			OriginalURL: originalURL,
		}

		if err := s.repo.Save(link); err != nil {
			return model.Link{}, err
		}

		return link, nil
	}

	return model.Link{}, ErrCodeGeneration
}

func (s *LinkService) GetByCode(code string) (model.Link, error) {
	link, found := s.repo.FindByCode(code)
	if found {
		return link, nil
	}

	return model.Link{}, ErrNotFound
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: %s", ErrInvalidURL, errors.New("unsupported URL scheme"))
	}

	if u.Host == "" {
		return fmt.Errorf("%w: %s", ErrInvalidURL, errors.New("URL must contain host"))
	}

	return nil
}

func generateCode(originalURL string) string {
	hash := md5.Sum([]byte(originalURL))
	code := fmt.Sprintf("%x", hash)[:8]

	return code
}
