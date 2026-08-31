package service

import "github.com/google/uuid"

// Storage — интерфейс хранилища ссылок.
type Storage interface {
	SaveURL(url, slug string) error    // сохраняет соответствие между исходным URL и короткой ссылкой
	Exists(url string) (string, bool)  // возвращает короткую ссылку по исходному URL, если она существует
	GetURL(slug string) (string, bool) // возвращает исходный URL по короткой ссылке, если она существует (вспомогательная функция чтобы избежать дублей)
}

// SlugGenerator генерирует идентификатор короткой ссылки.
type SlugGenerator func() (string, error)

// Shortener — сервис сокращения ссылок.
type Shortener struct {
	storage       Storage
	baseURL       string
	slugGenerator SlugGenerator
}

func NewShortener(storage Storage, baseURL string, generator SlugGenerator) *Shortener {
	return &Shortener{
		storage:       storage,
		baseURL:       baseURL,
		slugGenerator: generator,
	}
}

// Shorten сохраняет URL и возвращает короткую ссылку.
// Если URL уже сокращался, возвращает существующую короткую ссылку.
// Работает идемпотентно: повторные вызовы с одним и тем же URL возвращают один и тот же результат.
func (s *Shortener) Shorten(url string) (string, error) {
	if slug, ok := s.storage.Exists(url); ok {
		return s.baseURL + "/" + slug, nil
	}

	slug, err := s.slugGenerator()
	if err != nil {
		return "", err
	}

	if err := s.storage.SaveURL(url, slug); err != nil {
		return "", err
	}

	return s.baseURL + "/" + slug, nil
}

// Resolve возвращает исходный URL по короткой ссылке.
func (s *Shortener) Resolve(slug string) (string, bool) {
	return s.storage.GetURL(slug)
}

// RandomUUID генерирует случайный UUID.
func RandomUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
