package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DatabaseStorage struct {
	db *sql.DB
}

func NewDatabaseStorage(dsn string) (*DatabaseStorage, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	s := &DatabaseStorage{
		db: db,
	}

	return s, nil
}

func (s *DatabaseStorage) SaveURL(shortURL, originalURL string) error {
	return nil
}

func (s *DatabaseStorage) Exists(url string) (string, bool) {
	return "", false
}

func (s *DatabaseStorage) GetURL(slug string) (string, bool) {
	return "", false
}

// Ping проверяет доступность базы данных.
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
