package main

import (
	"fmt"
	"io"
	"shortener/internal/config"
	"shortener/internal/handler"
	"shortener/internal/middleware"
	"shortener/internal/repository"
	"shortener/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	if !cfg.GinLogs {
		gin.DefaultWriter = io.Discard // Отключает логи запросов
	}

	logger, _ := zap.NewProduction()

	storageName := cfg.Storage

	storage, err := storageSelector(storageName, cfg)

	if err != nil {
		logger.Fatal("failed to init storage", zap.Error(err))
	}

	defer func() {
		if closer, ok := storage.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Error("failed to close storage", zap.Error(err))
			}
		}
	}()

	shortener := service.NewShortener(storage, cfg.BaseURL, service.RandomUUID)

	var pinger handler.Pinger
	if p, ok := storage.(handler.Pinger); ok {
		pinger = p
	}

	h := handler.NewHandler(shortener, pinger)

	router := gin.New()

	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.GzipDecompressor())
	router.Use(middleware.GzipCompressor())

	router.POST("/", h.CreateShortLink)
	router.POST("/api/shorten", h.CreateShortLinkFromJSON)
	router.GET("/:shortLink", h.Redirect)
	router.GET("/ping", h.Ping)

	router.NoRoute(h.Default)

	if err := router.Run(cfg.ServerAddress); err != nil {
		logger.Fatal("failed to run server", zap.Error(err))
	}
}

// storageSelector создаёт хранилище по его имени.
func storageSelector(storageName string, cfg *config.Config) (service.Storage, error) {
	switch storageName {
	case "memory":
		return repository.NewInMemoryStorage()
	case "file":
		return repository.NewFileStorage(cfg.FileStoragePath)
	case "database":
		return repository.NewDatabaseStorage(cfg.DBDSN)
	default:
		return nil, fmt.Errorf("unknown storage: %q", storageName)
	}
}
