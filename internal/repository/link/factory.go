package link

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rucodencode/shortener/internal/config"
)

func NewLinkRepository(cfg config.Config, pool *pgxpool.Pool) (LinkRepository, error) {
	switch cfg.RepoType {
	case config.RepoTypePostgres:
		return NewPostgresRepository(pool)
	case config.RepoTypeFile:
		return NewFileRepository(cfg.FileStoragePath)
	case config.RepoTypeMemory:
		return NewInMemoryRepository()
	default:
		return NewInMemoryRepository()
	}
}
