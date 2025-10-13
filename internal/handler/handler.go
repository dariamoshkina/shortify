package handler

import (
	"io"
	"net/http"
	"regexp"

	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	service interface {
		Shorten(string) (string, error)
		Restore(string) (string, error)
	}
}

func NewURLHandler(s *service.ShortenerService) *URLHandler {
	return &URLHandler{service: s}
}

func (h *URLHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.ShortenHandler)
	r.Get("/{id}", h.RestoreHandler)

	return r
}

func (h *URLHandler) ShortenHandler(res http.ResponseWriter, req *http.Request) {
	originalURL, _ := io.ReadAll(req.Body)
	defer req.Body.Close()

	shortURL, err := h.service.Shorten(string(originalURL))
	if err != nil {
		http.Error(res, "can't shorten", http.StatusInternalServerError)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func (h *URLHandler) RestoreHandler(res http.ResponseWriter, req *http.Request) {
	urlID := chi.URLParam(req, "id")

	if !pathIsValid(urlID) {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.Restore(urlID)
	if err != nil {
		http.NotFound(res, req)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func pathIsValid(path string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, path)
	return matched
}
