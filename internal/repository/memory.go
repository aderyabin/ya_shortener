package repository

import "sync"

// InMemoryStorage — потокобезопасное in-memory хранилище ссылок.
// urls — от короткой ссылки к исходному URL для быстрого разрешения по ID,
// shortIDs — от исходного URL к короткой ссылке для проверки дубликатов,
// registered — исходные URL в порядке регистрации.
type InMemoryStorage struct {
	mu        sync.RWMutex
	slugToURL map[string]string // короткая ссылка -> исходный URL
	urlToSlug map[string]string // исходный URL -> короткая ссылка (делаем ради ускорения поиска существующих коротких ссылок)
}

func NewInMemoryStorage() (*InMemoryStorage, error) {
	return &InMemoryStorage{
		slugToURL: make(map[string]string),
		urlToSlug: make(map[string]string),
	}, nil
}

// SaveURL сохраняет ссылку.
func (s *InMemoryStorage) SaveURL(url, slug string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Если URL уже существует, удаляем старую короткую ссылку,
	// чтобы избежать дубликатов и сохранить целостность данных.
	if oldSlug, ok := s.urlToSlug[url]; ok {
		delete(s.slugToURL, oldSlug)
	}

	s.slugToURL[slug] = url
	s.urlToSlug[url] = slug
}

// Exists Возвращает короткий урл по исходному URL и true, если она существует,
// иначе пустую строку и false.
func (s *InMemoryStorage) Exists(url string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	slug, ok := s.urlToSlug[url]
	return slug, ok
}

// GetURL возвращает исходный URL по короткой ссылке.
// Если короткая ссылка не найдена, возвращает пустую строку и false.
func (s *InMemoryStorage) GetURL(slug string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.slugToURL[slug]
	return url, ok
}

func (s *InMemoryStorage) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.slugToURL)
}

// Close реализует интерфейс Storage: для in-memory хранилища закрывать нечего.
func (s *InMemoryStorage) Close() error {
	return nil
}
