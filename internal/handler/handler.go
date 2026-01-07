package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/file"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage *storage.MapStorage
	config  *config.Config
	file    *file.URLRecord
}

func NewHandler(storage *storage.MapStorage, config *config.Config, file *file.URLRecord) *Handler {
	handler := &Handler{
		storage: storage,
		config:  config,
		file:    file,
	}
	return handler
}

func (h Handler) generate() string {
	res := make([]byte, 8)
	for i := range res {
		res[i] = h.config.Symbols[rand.IntN(len(h.config.Symbols))]
	}
	return string(res)
}

func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := h.generate()
	h.storage.Set(shortURL, string(originalURL))
	h.file.Save(shortURL, string(originalURL))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", h.config.Host, shortURL)
}

func (h *Handler) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	JSONRequest := model.JSONRequest{}
	if err := json.NewDecoder(r.Body).Decode(&JSONRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL := h.generate()
	h.storage.Set(shortURL, JSONRequest.URL)
	h.file.Save(shortURL, JSONRequest.URL)

	JSONResponse := model.JSONResponse{Result: h.config.ServerAddress + shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(JSONResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if resURL, exist := h.storage.Get(shortURL); exist {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
