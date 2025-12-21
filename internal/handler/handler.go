package handler

import (
	"io"
	"math/rand/v2"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/go-chi/chi/v5"
)

var Symbols = []rune("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")

var Storage = make(map[string]string)

func Generate() string {
	res := make([]rune, 8)
	for i := range res {
		res[i] = Symbols[rand.IntN(len(Symbols))]
	}
	return string(res)
}

type Handler struct {
	storage storage.Storage
	config  config.Config
}

func NewHandler(st storage.Storage, cfg config.Config) *Handler {
	handler := &Handler{
		storage: st,
		config:  cfg,
	}
	return handler
}

func (h *Handler) ShortPostReq(w http.ResponseWriter, r *http.Request) {
	resp, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := Generate()
	Storage[shortURL] = string(resp)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(h.config.ServerAddress + shortURL))
}

func (h *Handler) ShortGetReq(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "rf")

	if path == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	resURL, err := h.storage.Get(path)
	if !err {
		http.Error(w, "Not found", http.StatusNotFound)
	}
	if resURL != "" {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
