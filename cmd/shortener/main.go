package main

import (
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.ShortURL)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
