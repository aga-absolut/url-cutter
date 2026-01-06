package storage

type Storage interface {
	Set(string, string)
	Get(string) (string, bool)
}

type MapStorage struct {
	data map[string]string
}

func NewStorage() *MapStorage {
	return &MapStorage{data: make(map[string]string)}
}

func (s *MapStorage) Set(shortURL, longURL string) {
	s.data[shortURL] = longURL
}

func (s *MapStorage) Get(shortURL string) (string, bool) {
	str, exist := s.data[shortURL]
	return str, exist
}
