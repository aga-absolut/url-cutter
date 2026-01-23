package memory

import (
	"errors"
	"os"

	"github.com/aga-absolut/url-cutter/internal/storage/database"
)

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) Set(shortURL, originalURL string) (string, error) {
	for k, v := range s.data {
		if v == originalURL {
			return k, os.ErrExist
		}
	}
	s.data[shortURL] = originalURL
	return shortURL, nil
}

func (s *MemoryStorage) Get(shortURL string) (string, bool) {
	str, exist := s.data[shortURL]
	return str, exist
}

func (s *MemoryStorage) SetBatchURL(batch []database.ShotenBatchRequest) ([]database.ShortenResponseItem, error) {
	return nil, errors.New("SetBatchURL not implemented for MemoryStorage")
}
