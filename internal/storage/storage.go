package storage

type Storage struct {
	data map[string]string
}

func NewStorage(dt map[string]string) *Storage {
	dataURL := &Storage{
		data: dt,
	}
	return dataURL
}

func (s *Storage) Set(longURL, shortURL string) {
	s.data[shortURL] = longURL
}

func (s *Storage) Get(shortURL string) (string, bool) {
	str, exist := s.data[shortURL]
	return str, exist
}

