package handler

import (
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
	logger     *zap.SugaredLogger
	linkRepo   repository.PGLinkRepository
	storage    repository.Storage
	deleteChan chan string
}

func NewHandler(config *config.Config, storage repository.Storage, logger *zap.SugaredLogger, deleteChan chan string, linkRepo repository.PGLinkRepository) *Handler {
	handler := &Handler{
		storage:    storage,
		config:     config,
		deleteChan: deleteChan,
		linkRepo:   linkRepo,
		logger:     logger,
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
	cookie, err := r.Cookie("token")
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

	if err := cookie.Valid(); err != nil {
		h.logger.Errorw("Cookie validation failed", "error", err)
		http.Error(w, "Invalid cookie", http.StatusUnauthorized)
		return
	}

	userID, err := jwt.GetUserID(cookie.Value)
	if err != nil {
		http.Error(w, "Failed to get userID", http.StatusInternalServerError)
		h.logger.Errorw("Failed to get userID", "error", err)
		return
	}

	URLs, err := h.linkRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("Failed to get user URLs", "error", err, "userID", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(URLs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CheckConnecToDB(w http.ResponseWriter, r *http.Request) {
	if err := h.linkRepo.Ping(); err != nil {
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

	if len(originalURL) == 0 {
		http.Error(w, "error empty body", http.StatusBadRequest)
	}

	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := cookie.Valid(); err != nil {
		h.logger.Errorw("Cookie validation failed", "error", err)
		http.Error(w, "Invalid cookie", http.StatusUnauthorized)
		return
	}

	userID, err := jwt.GetUserID(cookie.Value)
	if err != nil {
		http.Error(w, "Failed to get userID", http.StatusInternalServerError)
		h.logger.Errorw("Failed to get userID", "error", err)
		return
	}

	shortURL := config.Generate(string(originalURL))
	if shortKey, err := h.storage.Set(r.Context(), shortURL, string(originalURL), userID); err != nil {
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
	defer r.Body.Close()

	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := cookie.Valid(); err != nil {
		h.logger.Errorw("Cookie validation failed", "error", err)
		http.Error(w, "Invalid cookie", http.StatusUnauthorized)
		return
	}

	userID, err := jwt.GetUserID(cookie.Value)
	if err != nil {
		http.Error(w, "Failed to get userID", http.StatusInternalServerError)
		h.logger.Errorw("Failed to get userID", "error", err)
		return
	}

	if len(batch) == 0 {
		http.Error(w, "error batch is empty", http.StatusBadRequest)
		return
	}

	response, err := h.linkRepo.SetBatchURL(r.Context(), batch, userID)
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

	if len(JSONRequest.URL) == 0 {
		http.Error(w, "error batch is empty", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := cookie.Valid(); err != nil {
		h.logger.Errorw("Cookie validation failed", "error", err)
		http.Error(w, "Invalid cookie", http.StatusUnauthorized)
		return
	}

	userID, err := jwt.GetUserID(cookie.Value)
	if err != nil {
		http.Error(w, "Failed to get userID", http.StatusInternalServerError)
		h.logger.Errorw("Failed to get userID", "error", err)
		return
	}

	shortURL := config.Generate(JSONRequest.URL)
	if shortKey, err := h.storage.Set(r.Context(), shortURL, JSONRequest.URL, userID); err != nil {
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
