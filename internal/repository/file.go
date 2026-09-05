package repository

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("open storage file %q: %w", path, err)
	}

	inMemStorage, err := NewInMemoryStorage()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("init in-memory storage: %w", err)
	}

	s := &FileStorage{
		InMemoryStorage: inMemStorage,
		file:            file,
		encoder:         json.NewEncoder(file),
	}

	if err := s.restore(); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("restore storage from file %q: %w", path, err)
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
			return fmt.Errorf("unmarshal storage record %q: %w", string(line), err)
		}

		s.InMemoryStorage.SaveURL(rec.OriginalURL, rec.ShortURL)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read storage file: %w", err)
	}

	return nil
}

// SaveURL сохраняет ссылку в память и дублирует запись в файл.
func (s *FileStorage) SaveURL(url, slug string) error {
	s.InMemoryStorage.SaveURL(url, slug)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.encoder.Encode(fileRecord{UUID: s.InMemoryStorage.Size(), ShortURL: slug, OriginalURL: url}); err != nil {
		return fmt.Errorf("encode storage record for short_url %q: %w", slug, err)
	}

	return nil
}

func (s *FileStorage) Exists(url string) (string, bool) {
	return s.InMemoryStorage.Exists(url)
}

func (s *FileStorage) GetURL(slug string) (string, bool) {
	return s.InMemoryStorage.GetURL(slug)
}

// Close закрывает файл хранилища.
func (s *FileStorage) Close() error {
	if err := s.file.Close(); err != nil {
		return fmt.Errorf("close storage file: %w", err)
	}
	return nil
}
