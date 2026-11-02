package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/rucodencode/shortener/internal/model"
	linkrepo "github.com/rucodencode/shortener/internal/repository/link"
)

var (
	ErrCodeGeneration = errors.New("failed to generate unique code")
	ErrNotFound       = errors.New("link not found")
	ErrInvalidURL     = errors.New("invalid URL")
	ErrAlreadyExists  = errors.New("link already exists")
)

type LinkService struct {
	mu   sync.Mutex
	repo linkrepo.LinkRepository
}

func NewLinkService(repo linkrepo.LinkRepository) *LinkService {
	return &LinkService{
		repo: repo,
	}
}

func (s *LinkService) Create(ctx context.Context, originalURL string) (*model.Link, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	link, err := s.repo.FindByOriginalURL(ctx, originalURL)
	if err == nil {
		return link, ErrAlreadyExists
	}

	if !errors.Is(err, linkrepo.ErrNotFound) {
		return nil, fmt.Errorf("create link: %w", err)
	}

	code, err := s.findFreeCode(ctx, originalURL)
	if err != nil {
		return nil, fmt.Errorf("find free code: %w", err)
	}

	link = &model.Link{
		Code:        code,
		OriginalURL: originalURL,
	}

	if err := s.repo.Save(ctx, link); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	return link, nil
}

func (s *LinkService) CreateBatch(ctx context.Context, originalURLs []string) ([]*model.Link, error) {
	for _, originalURL := range originalURLs {
		if err := validateURL(originalURL); err != nil {
			return nil, fmt.Errorf("create batch link: %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*model.Link, 0)
	newLinks := make([]*model.Link, 0)

	for _, originalURL := range originalURLs {
		link, err := s.repo.FindByOriginalURL(ctx, originalURL)
		if err == nil {
			result = append(result, link)
			continue
		}

		if !errors.Is(err, linkrepo.ErrNotFound) {
			return nil, fmt.Errorf("create batch link: %w", err)
		}

		code, err := s.findFreeCode(ctx, originalURL)
		if err != nil {
			return nil, fmt.Errorf("find free code: %w", err)
		}

		link = &model.Link{
			Code:        code,
			OriginalURL: originalURL,
		}

		result = append(result, link)
		newLinks = append(newLinks, link)
	}

	if len(newLinks) > 0 {
		err := s.repo.SaveBatch(ctx, newLinks)
		if err != nil {
			return nil, fmt.Errorf("create batch links: %w", err)
		}
	}

	return result, nil
}

func (s *LinkService) GetByCode(ctx context.Context, code string) (*model.Link, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if err == nil {
		return link, nil
	}

	if errors.Is(err, linkrepo.ErrNotFound) {
		return nil, fmt.Errorf("get link by code: %w", ErrNotFound)
	}

	return nil, fmt.Errorf("get link by code: %w", err)
}

func (s *LinkService) findFreeCode(ctx context.Context, originalURL string) (string, error) {
	const maxAttempts = 100

	for attempt := 0; attempt < maxAttempts; attempt++ {
		source := originalURL

		if attempt > 0 {
			source = fmt.Sprintf("%s:%d", originalURL, attempt)
		}

		code := generateCode(source)

		_, err := s.repo.FindByCode(ctx, code)
		if err == nil {
			continue
		}

		if !errors.Is(err, linkrepo.ErrNotFound) {
			return "", err
		}

		return code, nil
	}

	return "", ErrCodeGeneration
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
