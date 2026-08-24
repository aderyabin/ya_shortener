// Package config содержит конфигурацию сервиса и функцию её инициализации.
package config

import "flag"

// Config — конфигурация сервиса сокращения ссылок.
type Config struct {
	ServerAddress string // адрес запуска HTTP-сервера
	BaseURL       string // базовый адрес результирующего сокращённого URL
	GinLogs       bool   // флаг включения логов запросов Gin
}

// NewConfig инициализирует конфигурацию из аргументов командной строки.
func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL")
	flag.BoolVar(&cfg.GinLogs, "g", false, "включение логов запросов Gin (true/false)")
	flag.Parse()

	return cfg
}
