package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dariamoshkina/shortify/internal/repository/file"
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

	repo := file.NewFileURLRepository(appConfig.FileStorage)
	shortener := service.NewShortenerService(repo, appConfig.BaseURL, 6)
	urlHandler := handler.NewURLHandler(shortener)

	r := chi.NewRouter()
	r.Mount("/", urlHandler.Routes())

	loggerMiddleware := middleware.WithLogging(sugar)
	loggedRouter := loggerMiddleware(r)
	compressedRouter := middleware.WithCompress(loggedRouter)

	if err := http.ListenAndServe(appConfig.Addr, compressedRouter); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
}
