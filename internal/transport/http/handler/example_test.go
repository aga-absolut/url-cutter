package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/service"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"github.com/aga-absolut/url-cutter/internal/transport/http/middleware/jwt"
	"github.com/go-chi/chi/v5"
)

func ExampleHandler_PostHandler() {
	shortURL := "http://localhost:8080/Afhbwof2"
	cfg := &config.Config{HTTPServerAddress: "http://localhost:8080/"}
	memory := memory.NewMemoryStorage()
	service := service.Service{Config: cfg, Storage: memory}
	handler := Handler{service: &service}

	request := httptest.NewRequest(http.MethodPost, cfg.HTTPServerAddress, strings.NewReader(shortURL))
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
	resp.Body.Close()

	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandler_GetHandler() {
	memory := memory.NewMemoryStorage()
	memory.Set(context.Background(), "afhbw222", "https://absolute.ru", 0)

	cfg := &config.Config{HTTPServerAddress: "/afhbw222"}
	service := service.Service{Config: cfg, Storage: memory}
	handler := Handler{service: &service}

	r := chi.NewRouter()
	r.Get("/{id}", handler.GetHandler)

	request := httptest.NewRequest(http.MethodGet, cfg.HTTPServerAddress, nil)
	recoder := httptest.NewRecorder()

	r.ServeHTTP(recoder, request)

	fmt.Println(recoder.Code)
	fmt.Println(recoder.Header().Get("Location"))

	// Output:
	// 307
	// https://absolute.ru
}
