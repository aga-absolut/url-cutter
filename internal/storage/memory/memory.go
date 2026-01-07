package memory

type Storage interface {
	Set(string, string)
	Get(string) (string, bool)
}

type MemoryStorage struct {
	data map[string]string
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) Set(shortURL, longURL string) {
	s.data[shortURL] = longURL
}

func (s *MemoryStorage) Get(shortURL string) (string, bool) {
	str, exist := s.data[shortURL]
	return str, exist
}
