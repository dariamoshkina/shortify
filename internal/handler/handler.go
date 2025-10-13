package handler

import (
	"io"
	"net/http"
	"regexp"

	"github.com/dariamoshkina/shortify/internal/service"
)

type URLHandler struct {
	service *service.ShortenerService
}

func NewURLHandler(s *service.ShortenerService) *URLHandler {
	return &URLHandler{service: s}
}

func (h *URLHandler) PostHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost || req.URL.Path != "/" {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL, _ := io.ReadAll(req.Body)
	shortURL, err := h.service.Shorten(string(originalURL))
	if err != nil {
		http.Error(res, "can't shorten", http.StatusInternalServerError)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func (h *URLHandler) GetHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet || !pathIsValid(req.URL.Path) {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	urlID := req.PathValue("id")

	originalURL, err := h.service.Restore(urlID)
	if err != nil {
		http.NotFound(res, req)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func pathIsValid(path string) bool {
	matched, _ := regexp.MatchString(`^/[a-zA-Z0-9]+$`, path)
	return matched
}
