package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dariamoshkina/shortify/internal/repository/file"
	"github.com/dariamoshkina/shortify/internal/repository/inmemory"
	"github.com/dariamoshkina/shortify/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/dariamoshkina/shortify/internal/config"
	"github.com/dariamoshkina/shortify/internal/handler"
	"github.com/dariamoshkina/shortify/internal/middleware"
	"github.com/dariamoshkina/shortify/internal/service"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	r := chi.NewRouter()
	appConfig := config.Parse()

	var (
		repo service.URLRepository
		pool *pgxpool.Pool
	)
	switch {
	case appConfig.DatabaseDSN != "":
		if err = postgres.RunMigrations(appConfig.DatabaseDSN); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
			os.Exit(1)
		}

		pool, err = pgxpool.New(context.Background(), appConfig.DatabaseDSN)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create pgx pool: %v\n", err)
			os.Exit(1)
		}
		defer pool.Close()

		repo = postgres.NewPostgresRepository(pool)

		dbHandler := handler.NewDBPinger(pool, sugar)
		r.Get("/ping", dbHandler.PingHandler)
	case appConfig.FileStorage != "":
		repo = file.NewFileRepository(appConfig.FileStorage)
	default:
		repo = inmemory.NewInMemoryRepository()
	}

	shortener := service.NewShortenerService(repo, appConfig.BaseURL, 6)
	urlHandler := handler.NewURLHandler(shortener, sugar)

	r.Post("/", urlHandler.ShortenHandler)
	r.Post("/api/shorten", urlHandler.ShortenJSONHandler)
	r.Post("/api/shorten/batch", urlHandler.ShortenBatchHandler)
	r.Get("/{id}", urlHandler.RestoreHandler)

	loggerMiddleware := middleware.WithLogging(sugar)
	loggedRouter := loggerMiddleware(r)
	compressedRouter := middleware.WithCompress(loggedRouter)

	if err = http.ListenAndServe(appConfig.Addr, compressedRouter); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
}
