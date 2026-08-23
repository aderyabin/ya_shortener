package repository

import "sync"

// InMemoryStorage — потокобезопасное in-memory хранилище ссылок.
// Ключ мапы — исходный URL, значение — короткая ссылка.
type InMemoryStorage struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls: make(map[string]string),
	}
}

// CreateShortUrl сохраняет пару «URL — короткая ссылка».
func (s *InMemoryStorage) CreateShortUrl(url, shortUrl string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[url] = shortUrl
}

// GetShortUrl возвращает коротку по исходному URL.
// Если исходный URL не найден, возвращает пустую строку и false.
func (s *InMemoryStorage) GetShortUrl(url string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.urls[url]
	return id, ok
}

// GetURL возвращает исходный URL по короткой ссылке.
// Если короткая ссылка не найдена, возвращает пустую строку и false.
func (s *InMemoryStorage) GetURL(shortUrl string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for url, storedID := range s.urls {
		if storedID == shortUrl {
			return url, true
		}
	}

	return "", false
}
