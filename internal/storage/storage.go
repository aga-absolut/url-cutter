package storage

import (
	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/repository"
	"github.com/aga-absolut/url-cutter/internal/storage/database"
	"github.com/aga-absolut/url-cutter/internal/storage/file"
	"github.com/aga-absolut/url-cutter/internal/storage/memory"
	"go.uber.org/zap"
)

func NewStorage(config *config.Config, logger zap.SugaredLogger) repository.Storage {
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
