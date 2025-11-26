package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/dariamoshkina/shortify/internal/model"
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
	r.Post("/api/shorten", h.ShortenJSONHandler)
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

func (h *URLHandler) ShortenJSONHandler(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var reqBody, respBody model.URL
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqBody); err != nil {
		http.Error(res, "invalid request body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.Shorten(reqBody.Original)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(res, "invalid URL", http.StatusBadRequest)
		} else {
			http.Error(res, "can't shorten", http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)

	respBody.Shortened = shortURL
	enc := json.NewEncoder(res)
	if err = enc.Encode(respBody); err != nil {
		http.Error(res, "can't encode response", http.StatusInternalServerError)
		return
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
