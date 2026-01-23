package memory

import "fmt"

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) Set(shortURL, originalURL string) error {
	for _, v := range s.data {
		if v == originalURL {
			return fmt.Errorf("not unique URL")
		}
	}
	s.data[shortURL] = originalURL
	return nil
}

func (s *MemoryStorage) Get(shortURL string) (string, bool) {
	str, exist := s.data[shortURL]
	return str, exist
}
