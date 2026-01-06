package storage

type Storage interface {
	Set(string, string)
	Get(string) (string, bool)
}

type MapStorage struct {
	Data map[string]string
}

func NewStorage() *MapStorage {
	return &MapStorage{Data: make(map[string]string)}
}

func (s *MapStorage) Set(shortURL, longURL string) {
	s.Data[shortURL] = longURL
}

func (s *MapStorage) Get(shortURL string) (string, bool) {
	str, exist := s.Data[shortURL]
	return str, exist
}
