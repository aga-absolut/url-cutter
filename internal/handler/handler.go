package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/repository"
	"github.com/aga-absolut/url-cutter/middleware/jwt"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type Handler struct {
	config     *config.Config
	logger     zap.SugaredLogger
	storage    repository.Storage
	deleteChan chan string
}

func NewHandler(config *config.Config, storage repository.Storage, logger zap.SugaredLogger, deleteChan chan string) *Handler {
	handler := &Handler{
		storage:    storage,
		config:     config,
		logger:     logger,
		deleteChan: deleteChan,
	}
	return handler
}

func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	var arrShortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&arrShortURLs); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	for _, shortURL := range arrShortURLs {
		h.deleteChan <- shortURL
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(202)
}

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	var ShortenURLs []model.ShortenURLs
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			token, _ := jwt.BuildJWTString()
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    token,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if c.Valid() != nil {
		h.logger.Errorw("Cookie validation failed", "error", err)
		http.Error(w, "Invalid cookie", http.StatusUnauthorized)
		return
	}

	tokenString := c.Value
	userID := jwt.GetUserID(tokenString)

	mapURLs, err := h.storage.GetByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("Failed to get user URLs", "error", err, "userID", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for shortKey, originalURL := range mapURLs {
		ShortenURLs = append(ShortenURLs, model.ShortenURLs{
			ShortURL:    h.config.Host + "/" + shortKey,
			OriginalURL: originalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ShortenURLs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CheckConnecToDB(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("pgx", h.config.DBDSN)
	if err != nil {
		http.Error(w, "error open database", http.StatusBadRequest)
		return
	}
	if err := db.Ping(); err != nil {
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

	shortURL := h.config.Generate(string(originalURL))
	if shortKey, err := h.storage.Set(r.Context(), shortURL, string(originalURL)); err != nil {
		if errors.Is(err, os.ErrExist) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(h.config.Host + "/" + shortKey))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.config.Host + "/" + shortURL))
}

func (h *Handler) PostBatchHandler(w http.ResponseWriter, r *http.Request) {
	var batch []model.ShotenBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(batch) == 0 {
		http.Error(w, "error batch is empty", http.StatusBadRequest)
		return
	}

	response, err := h.storage.SetBatchURL(r.Context(), batch)
	if err != nil {
		h.logger.Errorw("error set batch url", "error", err)
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

	shortURL := h.config.Generate(JSONRequest.URL)
	if shortKey, err := h.storage.Set(r.Context(), shortURL, JSONRequest.URL); err != nil {
		if errors.Is(err, os.ErrExist) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			JSONResponse := model.JSONResponse{Result: h.config.Host + "/" + shortKey}
			if err := json.NewEncoder(w).Encode(JSONResponse); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	if resURL, exist := h.storage.Get(r.Context(), shortURL); exist {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusGone)
	}
}
