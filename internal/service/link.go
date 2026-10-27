package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"

	"github.com/rucodencode/shortener/internal/model"
)

var (
	ErrCodeCollision = errors.New("code collision")
	ErrNotFound      = errors.New("not found")
	ErrInvalidURL    = errors.New("invalid url")
)

type LinkRepository interface {
	Save(link model.Link)
	FindByCode(code string) (model.Link, bool)
	FindByOriginalURL(originalURL string) (model.Link, bool)
}

type LinkService struct {
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

	link, found := s.repo.FindByOriginalURL(originalURL)
	if found {
		return link, nil
	}

	code := generateCode(originalURL)
	_, found = s.repo.FindByCode(code)
	if found {
		return model.Link{}, ErrCodeCollision
	}

	link = model.Link{Code: code, OriginalURL: originalURL}
	s.repo.Save(link)

	return link, nil
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
