package main

import (
	"io"
	"shortener/internal/config"
	"shortener/internal/handler"
	"shortener/internal/repository"
	"shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.NewConfig()

	if !cfg.GinLogs {
		gin.DefaultWriter = io.Discard // Отключает логи запросов
	}

	// Пока сохраним все в памяти, позже заменим на БД.
	storage := repository.NewInMemoryStorage()

	shortener := service.NewShortener(storage, cfg.BaseURL, service.RandomUUID)
	h := handler.NewHandler(shortener)

	router := gin.Default()

	router.POST("/", h.CreateShortLink)
	router.GET("/:shortLink", h.Redirect)

	router.NoRoute(h.Default)

	if err := router.Run(cfg.ServerAddress); err != nil {
		panic(err)
	}
}
