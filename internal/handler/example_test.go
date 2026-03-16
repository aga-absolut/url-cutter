package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"github.com/aga-absolut/url-cutter/middleware/jwt"
	"github.com/go-chi/chi/v5"
)

func ExampleHandler_PostHandler() {
	shortURL := "http://localhost:8080/Afhbwof2"
	cfg := &config.Config{ServerAddress: "http://localhost:8080/"}
	memory := memory.NewMemoryStorage()

	handler := Handler{
		config:  cfg,
		storage: memory,
	}

	request := httptest.NewRequest(http.MethodPost, cfg.ServerAddress, strings.NewReader(shortURL))
	request.Header.Set("Content-Type", "application/json")

	token, _ := jwt.BuildJWTString()
	request.AddCookie(&http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	recoder := httptest.NewRecorder()

	h := http.HandlerFunc(handler.PostHandler)

	h.ServeHTTP(recoder, request)
	resp := recoder.Result()

	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandler_GetHandler() {
	memory := memory.NewMemoryStorage()
	memory.Set(context.Background(), "afhbw222", "https://absolute.ru", 0)

	cfg := &config.Config{ServerAddress: "/afhbw222"}
	handler := &Handler{config: cfg, storage: memory}

	r := chi.NewRouter()
	r.Get("/{id}", handler.GetHandler)

	request := httptest.NewRequest(http.MethodGet, cfg.ServerAddress, nil)
	recoder := httptest.NewRecorder()

	r.ServeHTTP(recoder, request)

	fmt.Println(recoder.Code)
	fmt.Println(recoder.Header().Get("Location"))

	// Output:
	// 307
	// https://absolute.ru
}
