// Package config содержит конфигурацию сервиса и функцию её инициализации.
package config

import (
	"flag"
	"os"
	"slices"
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultGinLogs         = false
	defaultFileStoragePath = "storage.json"
	defaultDBDSN           = "postgres://@localhost:5432/url_shortener?sslmode=disable"
	defaultStorage         = "database"
)

// Config — конфигурация сервиса сокращения ссылок.
type Config struct {
	ServerAddress   string // адрес запуска HTTP-сервера
	BaseURL         string // базовый адрес результирующего сокращённого URL
	GinLogs         bool   // флаг включения логов запросов Gin
	FileStoragePath string // путь до JSON-файла, куда сохраняются сокращённые URL
	Storage         string // тип хранилища: memory, file, database
	DBDSN           string // DSN для подключения к базе данных PostgreSQL
}

// NewConfig инициализирует конфигурацию из аргументов командной строки
// и переменных окружения.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func NewConfig() *Config {
	cfg := &Config{}

	availableStorageTypes := []string{"memory", "file", "database"}

	flag.StringVar(&cfg.Storage, "s", defaultStorage, "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "базовый адрес результирующего сокращённого URL")
	flag.BoolVar(&cfg.GinLogs, "g", defaultGinLogs, "включение логов запросов Gin (true/false)")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "путь до JSON-файла для хранения сокращённых URL")
	flag.StringVar(&cfg.DBDSN, "d", defaultDBDSN, "DSN для подключения к базе данных PostgreSQL")
	flag.Parse()

	if !slices.Contains(availableStorageTypes, cfg.Storage) {
		cfg.Storage = defaultStorage
	}

	if env, ok := os.LookupEnv("SERVER_ADDRESS"); ok && env != "" {
		cfg.ServerAddress = env
	}
	if env, ok := os.LookupEnv("BASE_URL"); ok && env != "" {
		cfg.BaseURL = env
	}
	if env, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && env != "" {
		cfg.FileStoragePath = env
	}

	if env, ok := os.LookupEnv("DATABASE_URL"); ok && env != "" {
		cfg.DBDSN = env
	}

	return cfg
}
