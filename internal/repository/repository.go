package repository

import (
	"context"
	"net"

	"github.com/aga-absolut/url-cutter/internal/model"
)

// Storage интерфейс для взаимодействия с базой данных.
type Storage interface {
	Get(ctx context.Context, shortURL string) (string, bool)
	Set(ctx context.Context, shortURL, originalURL string, userID int) (string, error)
	SetBatchURL(ctx context.Context, batch []model.ShortenBatchRequest, userID int) ([]model.ShortenResponseItem, error)
	GetByUserID(ctx context.Context, userID int) ([]model.ShortenURLs, error)
	DeletedFlag(ctx context.Context, shortURL string) error
	GetURLsCount(ctx context.Context) (int, error)
	Ping() error
}

// Storage интерфейс для взаимодействия с серверами.
type Service interface {
	Ping() error
	DeleteURLs(arrShortURLs []string)
	GetURL(ctx context.Context, shortURL string) (string, bool)
	GetStats(ctx context.Context, ip net.IP) (model.ResponseStats, error)
	GetUserURLs(ctx context.Context, userID int) ([]model.ShortenURLs, error)
	SetURL(ctx context.Context, originalURL []byte, userID int) (string, error)
	SetURLFromJSON(ctx context.Context, req model.JSONRequest, userID int) (model.JSONResponse, error)
	SetBatchURLs(ctx context.Context, batch []model.ShortenBatchRequest, userID int) ([]model.ShortenResponseItem, error)
}
