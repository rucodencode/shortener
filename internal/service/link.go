package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/rucodencode/shortener/internal/model"
	"github.com/rucodencode/shortener/internal/repository/link"
)

var (
	ErrCodeGeneration = errors.New("failed to generate unique code")
	ErrNotFound       = errors.New("not found")
	ErrInvalidURL     = errors.New("invalid url")
)

type LinkService struct {
	mu   sync.Mutex
	repo link.LinkRepository
}

func NewLinkService(repo link.LinkRepository) *LinkService {
	return &LinkService{
		repo: repo,
	}
}

func (s *LinkService) Create(ctx context.Context, originalURL string) (*model.Link, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	link, found := s.repo.FindByOriginalURL(ctx, originalURL)
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

		existing, found := s.repo.FindByCode(ctx, code)
		if found {
			if existing.OriginalURL == originalURL {
				return existing, nil
			}

			continue
		}

		link := &model.Link{
			Code:        code,
			OriginalURL: originalURL,
		}

		if err := s.repo.Save(ctx, link); err != nil {
			return nil, err
		}

		return link, nil
	}

	return nil, ErrCodeGeneration
}

func (s *LinkService) GetByCode(ctx context.Context, code string) (*model.Link, error) {
	link, found := s.repo.FindByCode(ctx, code)
	if found {
		return link, nil
	}

	return nil, ErrNotFound
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
