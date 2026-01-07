package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/file"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/go-chi/chi/v5"
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
			body:           "https://google.com",
		},
		{
			name:           "second simple test",
			postStatusCode: http.StatusCreated,
			getStatusCode:  http.StatusTemporaryRedirect,
			contentType:    "text/plain",
			body:           "https://yandex.ru",
		},
	}

	cfg := config.Config{
		ServerAddress: "http://localhost:8080/",
		Symbols:       []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"),
	}
	storage := storage.NewStorage()
	file := file.NewURLRecord(&cfg)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := NewHandler(storage, &cfg, file)

			router := chi.NewRouter()
			router.Get("/{id}", hand.GetHandler)
			router.Post("/", hand.PostHandler)

			//----------------------------------------Post request

			req := httptest.NewRequest(http.MethodPost, cfg.ServerAddress, strings.NewReader(tt.body))
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

			req = httptest.NewRequest(http.MethodGet, cfg.ServerAddress+shortURL, nil)
			w = httptest.NewRecorder()

			router.ServeHTTP(w, req)

			r = w.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.getStatusCode, r.StatusCode)
			assert.Equal(t, tt.body, r.Header.Get("Location"))
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
			body:        `{"url": "https://yandex.ru"}`,
		},
	}
	cfg := config.Config{
		ServerAddress: "http://localhost:8080/api/shorten",
		Symbols:       []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"),
	}
	storage := storage.NewStorage()
	file := file.NewURLRecord(&cfg)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(storage, &cfg, file)

			router := chi.NewRouter()
			router.Post("/api/shorten", handler.JSONPostHandler)

			req := httptest.NewRequest(http.MethodPost, cfg.ServerAddress, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			r := w.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.statusCode, r.StatusCode)
			assert.Equal(t, tt.contentType, r.Header.Get("Content-Type"))

			body := w.Body.String()
			assert.Contains(t, body, `"result"`, "answer must have a result")

			resp := model.JSONResponse{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Failed to conver to Json")
		})
	}
}
