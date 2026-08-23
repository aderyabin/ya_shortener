// Package config содержит конфигурацию сервиса и функцию её инициализации.
package config

import (
	"flag"
	"os"
)

const (
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
	defaultGinLogs       = "off"
)

// Config — конфигурация сервиса сокращения ссылок.
type Config struct {
	ServerAddress string // адрес запуска HTTP-сервера
	BaseURL       string // базовый адрес результирующего сокращённого URL
	GinLogs       string // флаг включения логов запросов Gin
}

// NewConfig инициализирует конфигурацию из аргументов командной строки
// и переменных окружения.
// Приоритет: переменная окружения > флаг командной строки > значение по умолчанию.
func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "базовый адрес результирующего сокращённого URL")
	flag.StringVar(&cfg.GinLogs, "g", defaultGinLogs, "включение логов запросов Gin (on/off)")

	flag.Parse()

	if env, ok := os.LookupEnv("SERVER_ADDRESS"); ok && env != "" {
		cfg.ServerAddress = env
	}
	if env, ok := os.LookupEnv("BASE_URL"); ok && env != "" {
		cfg.BaseURL = env
	}

	return cfg
}
