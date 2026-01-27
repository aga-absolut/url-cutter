package memory

import (
	"os"

	"github.com/aga-absolut/url-cutter/internal/model"
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

func (f *MemoryStorage) SetBatchURL(batch []model.ShotenBatchRequest) ([]model.ShortenResponseItem, error) {
	return nil, nil
}

func (s *MemoryStorage) GetByUserID(userID int) (map[string]string, error) {
	return nil, nil
}
