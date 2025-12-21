package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
)

var StorageTest = make(map[string]string)

func TestHandle(t *testing.T) {
	tests := []struct {
		name           string
		postStatusCode int
		getStatusCode  int
		contentType    string
		request        string
		body           string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := storage.NewStorage(StorageTest)

			hand := NewHandler(*data, config.Config{})

			router := chi.NewRouter()
			router.Get("/{rf}", hand.ShortGetReq)
			//----------------------------------------Post request
			req := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			hand.ShortPostReq(w, req)

			r := w.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.postStatusCode, r.StatusCode)
			assert.Equal(t, tt.contentType, r.Header.Get("Content-Type"))

			body := strings.TrimSpace(w.Body.String())

			if len(body) != 8 {
				t.Fatalf("Expected short key length 8, got %d: %s", len(body), body)
			}

			shortURL := body

			t.Logf("Storage after POST: %v", StorageTest)
			//----------------------------------------Get request

			req = httptest.NewRequest(http.MethodGet, "/" + shortURL, nil)
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
