package main

import (
	"github.com/dariamoshkina/shortify/internal/handler"
	"github.com/dariamoshkina/shortify/internal/repository/in_memory"
	"github.com/dariamoshkina/shortify/internal/service"

	"net/http"
)

func main() {
	baseURL := "http://localhost:8080"
	repo := in_memory.NewInMemoryUrlRepository()
	shortener := service.NewShortenerService(repo, baseURL)
	urlHandler := handler.NewURLHandler(shortener)

	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", urlHandler.GetHandler)
	mux.HandleFunc("/", urlHandler.PostHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
