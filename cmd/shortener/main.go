package main

import (
	"database/sql"
	"log"
	"net/http"

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
	cfg := config.NewConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	var dbConn *sql.DB
	var err error

	if cfg.RepoType == config.RepoTypePostgres {
		dbConn, err = db.NewDB(db.Config{DataSource: cfg.DB.DataSourceName})
		if err != nil {
			logger.Log.Fatal("failed to initialize database", zap.Error(err))
		}
	}

	repo, err := link.NewLinkRepository(cfg, dbConn)
	if err != nil {
		logger.Log.Fatal("repository initialization failed", zap.Error(err))
	}

	rt := router.NewRouter()

	linkService := service.NewLinkService(repo)
	webLinkHandler := web.NewLinkHandler(linkService, cfg.BaseURL)
	apiLinkHandler := api.NewLinkHandler(linkService, cfg.BaseURL)

	healthService := service.NewHealthService(dbConn)
	healthHandler := health.NewHealthHandler(healthService)

	router.RegisterHealthRouters(rt, healthHandler)
	router.RegisterAPIRouters(rt, apiLinkHandler)
	router.RegisterWebRouters(rt, webLinkHandler)

	logger.Log.Info("starting HTTP server", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, rt); err != nil {
		logger.Log.Fatal("HTTP server failed", zap.Error(err))
	}
}
