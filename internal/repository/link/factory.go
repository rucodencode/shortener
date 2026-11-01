package link

import (
	"database/sql"

	"github.com/rucodencode/shortener/internal/config"
)

func NewLinkRepository(cfg config.Config, db *sql.DB) (LinkRepository, error) {
	switch cfg.RepoType {
	case config.RepoTypePostgres:
		return NewPostgresRepository(db)
	case config.RepoTypeFile:
		return NewFileRepository(cfg.FileStoragePath)
	case config.RepoTypeMemory:
		return NewInMemoryRepository()
	default:
		return NewInMemoryRepository()
	}
}
