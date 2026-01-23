package repository

import (
	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/storage/database"
	"github.com/aga-absolut/url-cutter/internal/storage/file"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"go.uber.org/zap"
)

type Storage interface {
	Get(shortURL string) (string, bool)
	Set(shortURL, originalURL string) error
    SetBatchURL(batch []database.ShotenBatchRequest) ([]database.ShortenResponseItem, error)
}

func NewStorage(config *config.Config, logger zap.SugaredLogger) Storage {
	if config.DBDSN != "" {
		logger.Infow("config PostgreSQL")
		return database.NewDBPostgreSQL(config)
	}
	if config.FilePath != "" {
		logger.Infow("config file")
		return file.NewFile(config, logger)
	}
	logger.Infow("config memory")
	return memory.NewMemoryStorage()
}
