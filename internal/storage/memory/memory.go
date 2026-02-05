package memory

import (
	"context"
	"os"

	"github.com/aga-absolut/url-cutter/internal/model"
)

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) Set(ctx context.Context, shortURL, originalURL string, userID int) (string, error) {
	for k, v := range s.data {
		if v == originalURL {
			return k, os.ErrExist
		}
	}
	s.data[shortURL] = originalURL
	return shortURL, nil
}

func (s *MemoryStorage) Get(ctx context.Context, shortURL string) (string, bool) {
	str, exist := s.data[shortURL]
	return str, exist
}

func (s *MemoryStorage) SetBatchURL(ctx context.Context, batch []model.ShotenBatchRequest, userID int) ([]model.ShortenResponseItem, error) {
	return nil, nil
}

func (s *MemoryStorage) GetByUserID(ctx context.Context, userID int) ([]model.ShortenURLs, error) {
	return nil, nil
}

func (s *MemoryStorage) DeletedFlag(ctx context.Context, shortURL string) error {
	return nil
}

func (s *MemoryStorage) Ping() error {
	return nil
}
