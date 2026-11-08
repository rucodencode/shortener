package main

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rucodencode/shortener/internal/config"
	"github.com/rucodencode/shortener/internal/config/db"
	"github.com/rucodencode/shortener/internal/handler/api"
	"github.com/rucodencode/shortener/internal/handler/health"
	"github.com/rucodencode/shortener/internal/handler/web"
	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/repository/link"
	"github.com/rucodencode/shortener/internal/router"
	"github.com/rucodencode/shortener/internal/service"
	"go.uber.org/zap"
)

func main() {
	var err error
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	var pool *pgxpool.Pool

	if cfg.RepoType == config.RepoTypePostgres {
		pool, err = db.NewDB(db.Config{DataSource: cfg.DB.DataSourceName})
		if err != nil {
			logger.Log.Fatal("failed to initialize database", zap.Error(err))
		}
	}

	repo, err := link.NewLinkRepository(cfg, pool)
	if err != nil {
		logger.Log.Fatal("repository initialization failed", zap.Error(err))
	}

	rt := router.NewRouter(cfg)

	linkService := service.NewLinkService(repo)
	webLinkHandler := web.NewLinkHandler(linkService, cfg.BaseURL)
	apiLinkHandler := api.NewLinkHandler(linkService, cfg.BaseURL)

	healthService := service.NewHealthService(pool)
	healthHandler := health.NewHealthHandler(healthService)

	router.RegisterHealthRouters(rt, healthHandler)
	router.RegisterAPIRouters(rt, apiLinkHandler)
	router.RegisterWebRouters(rt, webLinkHandler)

	logger.Log.Info("starting HTTP server", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, rt); err != nil {
		logger.Log.Fatal("HTTP server failed", zap.Error(err))
	}
}
