package memory

import (
	"context"
	"os"
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
