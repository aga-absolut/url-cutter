package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/storage/file"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	memory *memory.MemoryStorage
	config *config.Config
	file   *file.File
	db     *sql.DB
}

func NewHandler(memory *memory.MemoryStorage, config *config.Config, file *file.File, db *sql.DB) *Handler {
	handler := &Handler{
		memory: memory,
		config: config,
		file:   file,
		db:     db,
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

func (h *Handler) GetPingHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := h.generate()
	h.memory.Set(shortURL, string(originalURL))
	h.file.Set(shortURL, string(originalURL))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.config.Host + "/" + shortURL))
}

func (h *Handler) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	JSONRequest := model.JSONRequest{}
	if err := json.NewDecoder(r.Body).Decode(&JSONRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL := h.generate()
	h.memory.Set(shortURL, JSONRequest.URL)
	h.file.Set(shortURL, JSONRequest.URL)

	JSONResponse := model.JSONResponse{Result: h.config.Host + "/" + shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(JSONResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if resURL, exist := h.memory.Get(shortURL); exist {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
