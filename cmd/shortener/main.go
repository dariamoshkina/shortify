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

	sugar := *logger.Sugar()

	appConfig := config.Parse()

	var repo service.URLRepository
	switch {
	case appConfig.DatabaseDSN != "":
		repo = postgres.NewPostgresRepository(appConfig.DatabaseDSN)
	case appConfig.FileStorage != "":
		repo = file.NewFileRepository(appConfig.FileStorage)
	default:
		repo = inmemory.NewInMemoryRepository()
	}

	shortener := service.NewShortenerService(repo, appConfig.BaseURL, 6)
	urlHandler := handler.NewURLHandler(shortener)
	dbHandler := handler.NewDBHandler(context.Background(), appConfig.DatabaseDSN)

	r := chi.NewRouter()
	r.Mount("/", urlHandler.Routes())
	r.Mount("/ping", dbHandler.Routes())

	loggerMiddleware := middleware.WithLogging(sugar)
	loggedRouter := loggerMiddleware(r)
	compressedRouter := middleware.WithCompress(loggedRouter)

	if err := http.ListenAndServe(appConfig.Addr, compressedRouter); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
}
