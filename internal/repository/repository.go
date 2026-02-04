package repository

import (
	"context"

	"github.com/aga-absolut/url-cutter/internal/model"
)

type Storage interface {
	Get(ctx context.Context, shortURL string) (string, bool)
	Set(ctx context.Context, shortURL, originalURL string, userID int) (string, error)
	SetBatchURL(ctx context.Context, batch []model.ShotenBatchRequest, userID int) ([]model.ShortenResponseItem, error)
	GetByUserID(ctx context.Context, userID int) (map[string]string, error)
	DeletedFlag(ctx context.Context, shortURL string) error
}
