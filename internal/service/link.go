package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sync"

	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/model"
	linkrepo "github.com/rucodencode/shortener/internal/repository/link"
	"go.uber.org/zap"
)

var (
	ErrCodeGeneration = errors.New("failed to generate unique code")
	ErrNotFound       = errors.New("link not found")
	ErrInvalidURL     = errors.New("invalid URL")
	ErrAlreadyExists  = errors.New("link already exists")
)

var shortCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{8}$`)

type deleteLinks struct {
	userID string
	codes  []string
}

type LinkService struct {
	mu              sync.Mutex
	repo            linkrepo.LinkRepository
	deleteLinksChan chan deleteLinks
}

func NewLinkService(repo linkrepo.LinkRepository) *LinkService {
	instance := &LinkService{
		repo:            repo,
		deleteLinksChan: make(chan deleteLinks, 1024),
	}

	go instance.flushDeleteLinks()

	return instance
}

func (s *LinkService) Create(ctx context.Context, userID string, originalURL string) (*model.Link, error) {
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
		UserID:      userID,
		Code:        code,
		OriginalURL: originalURL,
	}

	if err := s.repo.Save(ctx, link); err != nil {
		if errors.Is(err, linkrepo.ErrOriginalURLConflict) || linkrepo.IsCodeConflict(err) {
			existing, lookupErr := s.repo.FindByOriginalURL(ctx, originalURL)
			if lookupErr == nil {
				return existing, ErrAlreadyExists
			}
			if !errors.Is(lookupErr, linkrepo.ErrNotFound) {
				return nil, fmt.Errorf("find URL after write conflict: %w", lookupErr)
			}
		}

		return nil, fmt.Errorf("create link: %w", err)
	}

	return link, nil
}

func (s *LinkService) CreateBatch(ctx context.Context, userID string, originalURLs []string) ([]*model.Link, error) {
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
			UserID:      userID,
			Code:        code,
			OriginalURL: originalURL,
		}

		result = append(result, link)
		newLinks = append(newLinks, link)
	}

	if len(newLinks) > 0 {
		if err := s.repo.SaveBatch(ctx, newLinks); err != nil {
			if errors.Is(err, linkrepo.ErrOriginalURLConflict) {
				return nil, fmt.Errorf("create batch links: %w: %w", ErrAlreadyExists, err)
			}
			return nil, fmt.Errorf("create batch links: %w", err)
		}
	}

	return result, nil
}

func (s *LinkService) DeleteLinksByCodes(ctx context.Context, userID string, codes []string) error {
	data := deleteLinks{
		userID: userID,
		codes:  codes,
	}

	select {
	case s.deleteLinksChan <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *LinkService) flushDeleteLinks() {
skip:
	for link := range s.deleteLinksChan {
		for _, code := range link.codes {
			if !validateCode(code) {
				logger.Log.Debug("link code is not valid: %s", zap.String("code", code))
				continue skip
			}
		}

		err := s.repo.DeleteByCodesAndUserID(context.TODO(), link.userID, link.codes)
		if err != nil {
			logger.Log.Debug("cannot delete links", zap.Error(err))
		}
	}

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

func (s *LinkService) GetListByUserID(ctx context.Context, userID string) ([]*model.Link, error) {
	links, err := s.repo.FindByUserID(ctx, userID)
	if err == nil {
		return links, nil
	}

	if errors.Is(err, linkrepo.ErrNotFound) {
		return nil, fmt.Errorf("get link by userID: %w", ErrNotFound)
	}

	return nil, fmt.Errorf("get link by userID: %w", err)
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

func validateCode(code string) bool {
	return shortCodePattern.MatchString(code)
}

func generateCode(originalURL string) string {
	hash := md5.Sum([]byte(originalURL))
	code := fmt.Sprintf("%x", hash)[:8]

	return code
}
