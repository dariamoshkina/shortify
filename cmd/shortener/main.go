package main

import (
	"github.com/dariamoshkina/shortify/internal/handler"
	"github.com/dariamoshkina/shortify/internal/repository/inmemory"
	"github.com/dariamoshkina/shortify/internal/service"

	"net/http"
)

func main() {
	baseURL := "http://localhost:8080"
	repo := inmemory.NewInMemoryURLRepository()
	shortener := service.NewShortenerService(repo, baseURL)
	urlHandler := handler.NewURLHandler(shortener)

	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", urlHandler.RestoreHandler)
	mux.HandleFunc("/", urlHandler.ShortenHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
