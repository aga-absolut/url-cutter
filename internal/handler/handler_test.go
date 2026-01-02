package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name           string
		postStatusCode int
		getStatusCode  int
		contentType    string
		request        string
		body           string
		serverAddress  string
	}{
		{
			name:           "first simple test",
			postStatusCode: http.StatusCreated,
			getStatusCode:  http.StatusTemporaryRedirect,
			contentType:    "text/plain",
			request:        "http://localhost:8080/",
			body:           "https://google.com",
		},
		{
			name:           "second simple test",
			postStatusCode: http.StatusCreated,
			getStatusCode:  http.StatusTemporaryRedirect,
			contentType:    "text/plain",
			request:        "http://localhost:8080/",
			body:           "https://yandex.ru",
		},
	}

	cfg := config.Config{
		ServerAddress: "http://localhost:8080",
		Host:          ":8080",
	}
	storage := storage.NewStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := NewHandler(storage, &cfg)

			router := chi.NewRouter()
			router.Get("/{rf}", hand.ShortGetReq)
			router.Post("/", hand.ShortPostReq)

			//----------------------------------------Post request

			req := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			r := w.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.postStatusCode, r.StatusCode)
			assert.Equal(t, tt.contentType, r.Header.Get("Content-Type"))

			body := w.Body.String()

			shortURL := body[strings.LastIndex(body, "/")+1:]
			if len(shortURL) != 8 {
				t.Fatalf("Expected short key length 8, got %d: %s", len(shortURL), shortURL)
			}

			//----------------------------------------Get request

			req = httptest.NewRequest(http.MethodGet, tt.request+shortURL, nil)
			w = httptest.NewRecorder()

			router.ServeHTTP(w, req)

			r = w.Result()
			defer r.Body.Close()

			assert.Equal(t, r.StatusCode, tt.getStatusCode)

			location := r.Header.Get("Location")
			assert.Equal(t, location, tt.body)
		})
	}
}

func TestPostReqJSON(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		contentType string
		request     string
		body        string
	}{
		{
			name:        "first simple test",
			statusCode:  http.StatusCreated,
			contentType: "application/json",
			request:     "http://localhost:8080/api/shorten",
			body:        `{"url": "https://yandex.ru"}`,
		},
	}
	cfg := config.Config{
		ServerAddress: "http://localhost:8080",
		Host:          ":8080",
	}
	storage := storage.NewStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(storage, &cfg)

			router := chi.NewRouter()
			router.Use(middleware.CleanPath)
			router.Post("/api/shorten", handler.HandlerJSON)

			req := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			r := w.Result()
			defer r.Body.Close()

			assert.Equal(t, r.StatusCode, tt.statusCode)
			assert.Equal(t, r.Header.Get("Content-Type"), tt.contentType)

			body := w.Body.String()
			assert.Contains(t, body, `"result"`, "в ответе должен быть ключ result")

			var resp JsonG
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Failed to conver to Json")
		})
	}
}
