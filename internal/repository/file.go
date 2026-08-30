package repository

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"sync"
)

// fileRecord — запись о ссылке в файле хранилища.
type fileRecord struct {
	UUID        int    `json:"uuid"` // порядковый номер записи, оставлено для совместимости с автотестами
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// FileStorage — хранилище ссылок с персистентностью в файле.
// Данные держатся в памяти (InMemoryStorage) и каждая новая запись
// дублируется в файл построчно в формате JSON Lines.
// При создании хранилища записи из файла восстанавливаются в память.
type FileStorage struct {
	*InMemoryStorage
	file    *os.File
	mu      sync.Mutex // сериализует записи в файл (json.Encoder не потокобезопасен)
	encoder *json.Encoder
}

// NewFileStorage открывает (или создаёт) файл по пути path
// и восстанавливает из него сохранённые ссылки.
func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	inMemStorage, err := NewInMemoryStorage()
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	s := &FileStorage{
		InMemoryStorage: inMemStorage,
		file:            file,
		encoder:         json.NewEncoder(file),
	}

	if err := s.restore(); err != nil {
		_ = file.Close()
		return nil, err
	}

	return s, nil
}

// restore читает записи из файла и переносит их в память.
func (s *FileStorage) restore() error {
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var rec fileRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return err
		}

		s.InMemoryStorage.SaveURL(rec.OriginalURL, rec.ShortURL)
	}

	return scanner.Err()
}

// SaveURL сохраняет ссылку в память и дублирует запись в файл.
func (s *FileStorage) SaveURL(url, slug string) {
	s.InMemoryStorage.SaveURL(url, slug)

	s.mu.Lock()
	_ = s.encoder.Encode(fileRecord{UUID: s.InMemoryStorage.Size(), ShortURL: slug, OriginalURL: url})
	defer s.mu.Unlock()
}

func (s *FileStorage) Exists(url string) (string, bool) {
	return s.InMemoryStorage.Exists(url)
}

func (s *FileStorage) GetURL(slug string) (string, bool) {
	return s.InMemoryStorage.GetURL(slug)
}

// Close закрывает файл хранилища.
func (s *FileStorage) Close() error {
	return s.file.Close()
}
