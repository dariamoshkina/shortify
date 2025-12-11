package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var validPathRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

//go:generate mockery --name=ShortenerService --output=../mocks --case=underscore
type ShortenerService interface {
	Shorten(context.Context, string) (string, error)
	ShortenMany(context.Context, []model.BatchURL) ([]*model.BatchURL, error)
	Restore(context.Context, string) (*model.URL, error)
	GetUserURLs(context.Context) ([]model.BatchURL, error)
	Delete(context.Context, <-chan string)
}

type URLHandler struct {
	service ShortenerService
	logger  *zap.SugaredLogger
}

func NewURLHandler(s *service.ShortenerService, l *zap.SugaredLogger) *URLHandler {
	return &URLHandler{service: s, logger: l}
}

func (h *URLHandler) ShortenHandler(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	originalURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "failed to read request body", http.StatusBadRequest)
		return
	}

	resStatus := http.StatusCreated
	shortURL, err := h.service.Shorten(req.Context(), string(originalURL))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			http.Error(res, "invalid URL", http.StatusBadRequest)
			return
		case errors.Is(err, service.ErrDuplicate):
			resStatus = http.StatusConflict
		default:
			h.logger.Error("failed to shorten URL", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(resStatus)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
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

	resStatus := http.StatusCreated
	shortURL, err := h.service.Shorten(req.Context(), reqBody.Original)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			http.Error(res, "invalid URL", http.StatusBadRequest)
			return
		case errors.Is(err, service.ErrDuplicate):
			resStatus = http.StatusConflict
		default:
			h.logger.Error("failed to shorten URL", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("content-type", "application/json")
	res.WriteHeader(resStatus)

	respBody.Shortened = shortURL
	enc := json.NewEncoder(res)
	if err = enc.Encode(respBody); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) ShortenBatchHandler(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var reqBody []model.BatchURL
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqBody); err != nil {
		http.Error(res, "invalid request body", http.StatusBadRequest)
		return
	}

	respBody, err := h.service.ShortenMany(req.Context(), reqBody)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(res, "invalid URL", http.StatusBadRequest)
		} else {
			h.logger.Error("failed to shorten URLs", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(res)
	if err := enc.Encode(respBody); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) RestoreHandler(res http.ResponseWriter, req *http.Request) {
	urlID := chi.URLParam(req, "id")

	if !validPathRegexp.MatchString(urlID) {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	restoredURL, err := h.service.Restore(req.Context(), urlID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(res, req)
		} else {
			h.logger.Error("failed to restore URL", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if restoredURL.Deleted != nil && *restoredURL.Deleted {
		res.WriteHeader(http.StatusGone)
		return
	}

	res.Header().Set("Location", restoredURL.Original)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) UserURLsHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	urls, err := h.service.GetUserURLs(ctx)
	if err != nil {
		h.logger.Error("failed to get user URLs", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("content-type", "application/json")

	if len(urls) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}
	res.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(res)
	if err = enc.Encode(urls); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) DeleteHandler(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var ids []string
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&ids); err != nil {
		http.Error(res, "invalid request body", http.StatusBadRequest)
		return
	}

	inputCh := make(chan string)
	go func() {
		defer close(inputCh)
		for _, id := range ids {
			inputCh <- id
		}
	}()

	h.service.Delete(req.Context(), inputCh)

	res.WriteHeader(http.StatusAccepted)
}
