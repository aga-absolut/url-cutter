package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/storage/database"
	"github.com/aga-absolut/url-cutter/internal/storage/file"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handler struct {
	memory *memory.MemoryStorage
	config *config.Config
	file   *file.File
	pgxDB  *database.DBPostgreSQL
	db     *sql.DB
}

func NewHandler(memory *memory.MemoryStorage, config *config.Config, file *file.File, pgxDB *database.DBPostgreSQL, db *sql.DB) *Handler {
	handler := &Handler{
		memory: memory,
		config: config,
		file:   file,
		pgxDB:  pgxDB,
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

func (h *Handler) CheckConnecToDB(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

func (h *Handler) PostBatchHandler(w http.ResponseWriter, r *http.Request) {
	var batch []database.ShotenBatchRequest
	var responseItem database.ShortenResponseItem
	var response []database.ShortenResponseItem

	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	_, err := h.db.Exec(`
	CREATE TABLE IF NOT EXISTS urls (
		short_url TEXT NOT NULL PRIMARY KEY,
		original_url TEXT NOT NULL
	);`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO urls VALUES ($1, $2)`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, v := range batch {
		shortKey := h.generate()
		_, err = stmt.Exec(shortKey, v.OriginalURL)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		responseItem.ShortURL = h.config.Host + "/" + shortKey
		responseItem.CorrelationID = v.CorrelationID
		response = append(response, responseItem)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
