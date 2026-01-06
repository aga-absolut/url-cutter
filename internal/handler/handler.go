package handler

import (
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/file"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/go-chi/chi/v5"
)

var Symbols = []rune("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")

func Generate() string {
	res := make([]rune, 8)
	for i := range res {
		res[i] = Symbols[rand.IntN(len(Symbols))]
	}
	return string(res)
}

type JSONRequest struct {
	URL string `json:"url"`
}
type JSONResponse struct {
	Result string `json:"result"`
}

type Handler struct {
	storage storage.MapStorage
	config  config.Config
	file    file.URLRecord
}

func NewHandler(st *storage.MapStorage, cfg *config.Config, file *file.URLRecord) *Handler {
	handler := &Handler{
		storage: *st,
		config:  *cfg,
		file:    *file,
	}
	return handler
}

func (h *Handler) ShortPostReq(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := Generate()
	h.storage.Set(shortURL, string(originalURL))
	h.file.Save(shortURL, string(originalURL))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	if !strings.HasSuffix(h.config.ServerAddress, "/") {
		h.config.ServerAddress = h.config.ServerAddress + "/"
	}

	w.Write([]byte(h.config.ServerAddress + shortURL))
}

func (h *Handler) HandlerJSON(w http.ResponseWriter, r *http.Request) {
	js := JSONRequest{}
	if err := json.NewDecoder(r.Body).Decode(&js); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL := Generate()
	h.storage.Set(shortURL, js.URL)
	h.file.Save(shortURL, js.URL)

	if !strings.HasSuffix(h.config.ServerAddress, "/") {
		h.config.ServerAddress = h.config.ServerAddress + "/"
	}
	jss := JSONResponse{Result: h.config.ServerAddress + shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(jss); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ShortGetReq(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "rf")

	if path == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if resURL, err := h.storage.Get(path); err {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if resURL, err := h.file.ReadFile(path); err {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
