package service

import "github.com/google/uuid"

// Storage — интерфейс хранилища ссылок.
type Storage interface {
	CreateShortUrl(url, shortUrl string)   // сохраняет соответствие между исходным URL и короткой ссылкой
	GetShortUrl(url string) (string, bool) // возвращает короткую ссылку по исходному URL, если она существует
	GetURL(shortUrl string) (string, bool) // возвращает исходный URL по короткой ссылке, если она существует (вспомогательная функция чтобы избежать дублей)
}

// ShortUrlGenerator генерирует идентификатор короткой ссылки.
type ShortUrlGenerator func() (string, error)

// Shortener — сервис сокращения ссылок.
type Shortener struct {
	storage           Storage
	baseURL           string
	shortUrlGenerator ShortUrlGenerator
}

func NewShortener(storage Storage, baseURL string, generator ShortUrlGenerator) *Shortener {
	return &Shortener{
		storage:           storage,
		baseURL:           baseURL,
		shortUrlGenerator: generator,
	}
}

// Shorten сохраняет URL и возвращает короткую ссылку.
// Если URL уже сокращался, возвращает существующую короткую ссылку.
func (s *Shortener) Shorten(url string) (string, error) {
	if shortUrl, ok := s.storage.GetShortUrl(url); ok {
		return s.baseURL + "/" + shortUrl, nil
	}

	shortUrl, err := s.shortUrlGenerator()
	if err != nil {
		return "", err
	}

	s.storage.CreateShortUrl(url, shortUrl)

	return s.baseURL + "/" + shortUrl, nil
}

// Resolve возвращает исходный URL по короткой ссылке.
func (s *Shortener) Resolve(shortUrl string) (string, bool) {
	return s.storage.GetURL(shortUrl)
}

// RandomUUID генерирует случайный UUID.
func RandomUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
