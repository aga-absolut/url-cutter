package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/aga-absolut/url-cutter/internal/errs"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/service"
	"github.com/aga-absolut/url-cutter/internal/transport/http/middleware/jwt"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Структура обработчика
type Handler struct {
	service *service.Service
	logger  *zap.SugaredLogger
}

// NewHandler создает новую структуру Handler
func NewHandler(service *service.Service, logger *zap.SugaredLogger) *Handler {
	handler := &Handler{
		service: service,
		logger:  logger,
	}
	return handler
}

// GetUserURLs обрабатывает GET-запросы для возврата списка URL, только для авторизованного пользователя.
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
		h.logger.Errorw("Failed to get userID", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	urls, err := h.service.GetUserURLs(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInGettingURLs):
			h.logger.Errorw("Failed to get user URLs", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

		default:
			h.logger.Errorw("failed to get user URLs", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CheckConnecToDB проверяет соединение с базой данных
func (h *Handler) CheckConnecToDB(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// PostHandler обрабатывает POST-запросы для создания короткого URL.
func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	shortURL, err := h.service.SetURL(r.Context(), originalURL, userID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrEmptyBody):
			http.Error(w, "error empty body", http.StatusBadRequest)

		case errors.Is(err, errs.ErrURLAlreadyExists):
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))

		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// PostBatchHandler обрабатывает POST-запросы с телом в формате JSON для создания списка коротких URL.
func (h *Handler) PostBatchHandler(w http.ResponseWriter, r *http.Request) {
	var batch []model.ShortenBatchRequest
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

	response, err := h.service.SetBatchURLs(r.Context(), batch, userID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrEmptyBatch):
			http.Error(w, "error batch is empty", http.StatusBadRequest)

		case errors.Is(err, errs.ErrInGettingUserID):
			http.Error(w, "Failed to get userID", http.StatusInternalServerError)
			h.logger.Errorw("Failed to get userID", "error", err)

		default:
			h.logger.Errorw("error set batch url", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// JSONPostHandler обрабатывает POST-запросы с телом в формате JSON для создания короткого URL.
func (h *Handler) JSONPostHandler(w http.ResponseWriter, r *http.Request) {
	JSONRequest := model.JSONRequest{}
	if err := json.NewDecoder(r.Body).Decode(&JSONRequest); err != nil {
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

	response, err := h.service.SetURLFromJSON(r.Context(), JSONRequest, userID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrEmptyBody):
			http.Error(w, "error body is empty", http.StatusBadRequest)

		case errors.Is(err, errs.ErrURLAlreadyExists):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return

		default:
			h.logger.Errorw("error set batch url", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetHandler обрабатывает GET-запросы для возврата URL для авторизованного пользователя.
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if resURL, exist := h.service.GetURL(r.Context(), shortURL); exist {
		w.Header().Set("Location", resURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusGone)
	}
}

// DeleteUserURLs обрабатывает DELETE-запросы для удаления URL.
func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	var arrShortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&arrShortURLs); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	h.service.DeleteURLs(arrShortURLs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(202)
}

// GetStatsHandler обрабатывает GET-запросы для возврата количества URL и Users только доверенным сетям.
func (h *Handler) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	ipStr := r.Header.Get("X-Real-IP")
	forwarded := r.Header.Get("X-Forwarded-For")

	ip := net.ParseIP(ipStr)
	if ip == nil {
		ipStrs := strings.Split(forwarded, ",")
		if len(ipStrs) > 0 {
			ip = net.ParseIP(ipStrs[0])
		}
	}

	response, err := h.service.GetStats(r.Context(), ip)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrTrustedSubnetIsEmpty):
			http.Error(w, err.Error(), http.StatusForbidden)

		case errors.Is(err, errs.ErrInGettingUrlsCount):
			http.Error(w, err.Error(), http.StatusInternalServerError)

		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
