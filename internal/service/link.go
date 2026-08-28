package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"

	"github.com/rucodencode/shortener/internal/model"
	"github.com/rucodencode/shortener/internal/repository"
)

type LinkService struct {
	repo *repository.LinkRepository
}

func NewLinkService(repo *repository.LinkRepository) *LinkService {
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
		return model.Link{}, errors.New("code collision")
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

	return model.Link{}, errors.New("not found")
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("unsupported URL scheme")
	}

	if u.Host == "" {
		return errors.New("URL must contain host")
	}

	return nil
}

func generateCode(originalURL string) string {
	hash := md5.Sum([]byte(originalURL))
	code := fmt.Sprintf("%x", hash)[:8]

	return code
}
