package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/go-chi/chi/v5"
)

var validPathRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

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
	defer req.Body.Close()

	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "failed to read request body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.Shorten(string(originalURL))
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(res, "invalid URL", http.StatusBadRequest)
		} else {
			http.Error(res, "can't shorten", http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		fmt.Println("error writing response:", err)
	}
}

func (h *URLHandler) RestoreHandler(res http.ResponseWriter, req *http.Request) {
	urlID := chi.URLParam(req, "id")

	if !validPathRegexp.MatchString(urlID) {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.Restore(urlID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(res, req)
		} else {
			http.Error(res, "internal error", http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
