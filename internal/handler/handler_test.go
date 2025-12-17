package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
)

var StorageTest = make(map[string]string)

// func TestRequest(t *testing.T, ts httptest.Server, method, path string) (*http.Response, string) {
// 	req, err := http.NewRequest(method, ts.URL+path, nil)
// 	require.NoError(t, err)

// 	resp, err := ts.Client().Do(req)
// 	require.NoError(t, err)
// 	defer resp.Body.Close()

// 	res, err := io.ReadAll(resp.Body)
// 	require.NoError(t, err)

// 	return resp, string(res)
// }

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
			//----------------------------------------Post request
			req := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h := http.HandlerFunc(ShortPostReq)
			h(w, req)

			r := w.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.postStatusCode, r.StatusCode)
			assert.Equal(t, tt.contentType, r.Header.Get("Content-Type"))

			// Проверка, что создалась случайная ссылка
			if !strings.Contains(w.Body.String(), "http://localhost:8080/") {
				t.Error("missing prefix http://localhost:8080/")
			}

			longURL := w.Body.String()
			expBody := "http://localhost:8080/"
			shortURL := longURL[len(expBody):]

			//Создали мапу с ключом сген символов
			StorageTest[shortURL] = tt.body

			if len(shortURL) != 8 {
				t.Errorf("Expected 8 symbols after /, got %d . ", len(shortURL))
			}
			//----------------------------------------Get request
			router := chi.NewRouter()
			router.Get("/{rf}",ShortGetReq)

			req = httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
			w = httptest.NewRecorder()

			router.ServeHTTP(w, req)

			r = w.Result()
			defer r.Body.Close()

			assert.Equal(t, tt.getStatusCode, r.StatusCode)

			location := r.Header.Get("Location")
			assert.Equal(t, location, tt.body)

		})
	}
}
