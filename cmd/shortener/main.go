package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dariamoshkina/shortify/internal/config"
	"github.com/dariamoshkina/shortify/internal/handler"
	"github.com/dariamoshkina/shortify/internal/repository/inmemory"
	"github.com/dariamoshkina/shortify/internal/service"
)

func main() {
	appConfig := config.Parse()

	repo := inmemory.NewInMemoryURLRepository()
	shortener := service.NewShortenerService(repo, appConfig.BaseURL, 6)
	urlHandler := handler.NewURLHandler(shortener)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", urlHandler.Routes())

	if err := http.ListenAndServe(appConfig.Addr, r); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
}
