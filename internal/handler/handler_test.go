package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dariamoshkina/shortify/internal/model"
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

func (m *MockShortenerService) ShortenMany(urls []*model.BatchURL) ([]*model.BatchURL, error) {
	args := m.Called(urls)
	return args.Get(0).([]*model.BatchURL), args.Error(1)
}

var (
	requestURL   = "https://example.com"
	shortenedID  = "abc123"
	responseURL  = "http://localhost:8080/abc123"
	requestJSON  = fmt.Sprintf(`{"url":"%s"}`, requestURL)
	responseJSON = fmt.Sprintf(`{"result":"%s"}`, responseURL)
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
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			path:           "/",
			body:           requestURL,
			expectedStatus: http.StatusCreated,
			expectedBody:   responseURL,
		},
		{
			name:           "bad method",
			method:         http.MethodGet,
			path:           "/",
			body:           requestURL,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "bad path",
			method:         http.MethodPost,
			path:           "/shorten",
			body:           requestURL,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "service error",
			method:         http.MethodPost,
			path:           "/",
			body:           requestURL,
			mockReturnErr:  errors.New("boom"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestHandler()
			if tt.expectedBody != "" || tt.mockReturnErr != nil {
				mockService.On("Shorten", tt.body).Return(tt.expectedBody, tt.mockReturnErr)
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

func TestShortenJSONHandler(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		jsonBody         string
		mockResponseURL  string
		mockReturnErr    error
		expectedStatus   int
		expectedJSONBody string
	}{
		{
			name:             "success",
			method:           http.MethodPost,
			path:             "/api/shorten",
			jsonBody:         requestJSON,
			mockResponseURL:  responseURL,
			expectedStatus:   http.StatusCreated,
			expectedJSONBody: responseJSON,
		},
		{
			name:           "bad method",
			method:         http.MethodGet,
			path:           "/api/shorten",
			jsonBody:       requestJSON,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "bad path",
			method:         http.MethodPost,
			path:           "/shorten",
			jsonBody:       requestJSON,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "service error",
			method:         http.MethodPost,
			path:           "/api/shorten",
			jsonBody:       requestJSON,
			mockReturnErr:  errors.New("boom"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestHandler()
			if tt.expectedJSONBody != "" || tt.mockReturnErr != nil {
				mockService.On("Shorten", requestURL).Return(responseURL, tt.mockReturnErr)
			}

			router := chi.NewRouter()
			router.Mount("/", handler.Routes())

			w := doRequest(router, tt.method, tt.path, tt.jsonBody)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			if tt.expectedJSONBody != "" {
				assert.Equal(t, tt.expectedJSONBody, strings.TrimSpace(string(body)))
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
			mockReturnURL:  requestURL,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: requestURL,
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
