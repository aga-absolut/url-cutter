package repository

import (
	"context"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/aga-absolut/url-cutter/internal/storage/database"
	"github.com/aga-absolut/url-cutter/internal/storage/file"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"go.uber.org/zap"
)

type Storage interface {
	Get(ctx context.Context, shortURL string) (string, bool)
	Set(ctx context.Context, shortURL, originalURL string) (string, error)
	SetBatchURL(ctx context.Context, batch []model.ShotenBatchRequest) ([]model.ShortenResponseItem, error)
	GetByUserID(ctx context.Context, userID int) (map[string]string, error)
	DeletedFlag(ctx context.Context, shortURL string) error
}

func NewStorage(config *config.Config, logger zap.SugaredLogger) Storage {
	if config.DBDSN != "" {
		logger.Infow("config PostgreSQL")
		return database.NewDBPostgreSQL(config, logger)
	}
	if config.FilePath != "" {
		logger.Infow("config file")
		return file.NewFile(config, logger)
	}
	logger.Infow("config memory")
	return memory.NewMemoryStorage()
}
