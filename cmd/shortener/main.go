package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rucodencode/shortener/internal/config"
	"github.com/rucodencode/shortener/internal/config/db"
	"github.com/rucodencode/shortener/internal/handler/api"
	"github.com/rucodencode/shortener/internal/handler/health"
	"github.com/rucodencode/shortener/internal/handler/web"
	"github.com/rucodencode/shortener/internal/logger"
	"github.com/rucodencode/shortener/internal/repository"
	"github.com/rucodencode/shortener/internal/repository/link"
	"github.com/rucodencode/shortener/internal/router"
	"github.com/rucodencode/shortener/internal/service"
)

func main() {
	cfg := config.NewConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	db, err := db.NewDB(db.Config{DataSource: cfg.DB.DataSourceName})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	defer db.Close()

	repo, err := link.NewLinkRepository(cfg, db)
	if err != nil {
		log.Fatal("Rep didn't loaded")
	}

	rt := router.NewRouter()

	linkService := service.NewLinkService(repo)
	webLinkHandler := web.NewLinkHandler(linkService, cfg.BaseURL)
	apiLinkHandler := api.NewLinkHandler(linkService, cfg.BaseURL)

	repoHealth := repository.NewHealthRepository(db)
	healthService := service.NewHealthService(repoHealth)
	healthHandler := health.NewHealthHandler(healthService)

	router.RegisterHealthRouters(rt, healthHandler)
	router.RegisterAPIRouters(rt, apiLinkHandler)
	router.RegisterWebRouters(rt, webLinkHandler)

	fmt.Println("Running server on", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, rt))
}
