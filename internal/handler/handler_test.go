package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockShortenerService struct {
	mock.Mock
}

func (m *MockShortenerService) Shorten(original string) (string, error) {
	args := m.Called(original)
	return args.String(0), args.Error(1)
}

func (m *MockShortenerService) Restore(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

const (
	originalURL  = "https://example.com"
	shortenedID  = "abc123"
	shortenedURL = "http://localhost:8080/abc123"
)

func newTestHandler() (*URLHandler, *MockShortenerService) {
	mockService := new(MockShortenerService)
	handler := &URLHandler{service: mockService}
	return handler, mockService
}

func doRequest(router http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		mockReturnURL  string
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			path:           "/",
			body:           originalURL,
			mockReturnURL:  shortenedURL,
			expectedStatus: http.StatusCreated,
			expectedBody:   shortenedURL,
		},
		{
			name:           "bad method",
			method:         http.MethodGet,
			path:           "/",
			body:           originalURL,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "bad path",
			method:         http.MethodPost,
			path:           "/shorten",
			body:           originalURL,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "service error",
			method:         http.MethodPost,
			path:           "/",
			body:           originalURL,
			mockReturnErr:  errors.New("boom"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestHandler()
			if tt.mockReturnURL != "" || tt.mockReturnErr != nil {
				mockService.On("Shorten", originalURL).Return(tt.mockReturnURL, tt.mockReturnErr)
			}

			router := chi.NewRouter()
			router.Mount("/", handler.Routes())

			w := doRequest(router, tt.method, tt.path, tt.body)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestRestoreHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		urlID          string
		mockReturnURL  string
		mockReturnErr  error
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "success",
			method:         http.MethodGet,
			path:           "/" + shortenedID,
			urlID:          shortenedID,
			mockReturnURL:  originalURL,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: originalURL,
		},
		{
			name:           "bad method",
			method:         http.MethodPost,
			path:           "/" + shortenedID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid path",
			method:         http.MethodGet,
			path:           "/invalid_path!",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			method:         http.MethodGet,
			path:           "/" + shortenedID,
			urlID:          shortenedID,
			mockReturnErr:  service.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestHandler()
			if tt.mockReturnURL != "" || tt.mockReturnErr != nil {
				mockService.On("Restore", shortenedID).Return(tt.mockReturnURL, tt.mockReturnErr)
			}

			router := chi.NewRouter()
			router.Mount("/", handler.Routes())

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, resp.Header.Get("Location"))
			}
		})
	}
}
