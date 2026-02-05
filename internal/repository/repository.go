package repository

import (
	"context"
)

type Storage interface {
	Get(ctx context.Context, shortURL string) (string, bool)
	Set(ctx context.Context, shortURL, originalURL string, userID int) (string, error)
}
