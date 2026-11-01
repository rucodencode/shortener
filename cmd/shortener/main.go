package main

import (
	"database/sql"
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

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
			log.Fatalf("failed to open DB conntection: %v", err)
		}

		if err = dbConn.Ping(); err != nil {
			log.Fatalf("DB ping failed: %v", err)
		}

		driver, err := postgres.WithInstance(dbConn, &postgres.Config{})
		if err != nil {
			log.Fatalf("failed to create migrate driver: %v", err)
		}

		migrationPath := "file://./migrations"

		m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
		if err != nil {
			log.Fatalf("failed to create migrate instance: %v", err)
		}
		defer m.Close()
		m.Up()

		log.Println("Migrations applied successfully")
	}

	repo, err := link.NewLinkRepository(cfg, dbConn)
	if err != nil {
		log.Fatalf("repository initialization failed: %v", err)
	}

	rt := router.NewRouter()

	linkService := service.NewLinkService(repo)
	webLinkHandler := web.NewLinkHandler(linkService, cfg.BaseURL)
	apiLinkHandler := api.NewLinkHandler(linkService, cfg.BaseURL)

	repoHealth := repository.NewHealthRepository(dbConn)
	healthService := service.NewHealthService(repoHealth)
	healthHandler := health.NewHealthHandler(healthService)

	router.RegisterHealthRouters(rt, healthHandler)
	router.RegisterAPIRouters(rt, apiLinkHandler)
	router.RegisterWebRouters(rt, webLinkHandler)

	fmt.Println("Running server on", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, rt))
}
