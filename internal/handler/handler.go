package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"shortener/internal/model"
	"shortener/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler обрабатывает HTTP-запросы сервиса сокращения ссылок.
type Handler struct {
	shortener *service.Shortener
}

func NewHandler(shortener *service.Shortener) *Handler {
	return &Handler{shortener: shortener}
}

// CreateShortLink принимает исходный URL в теле запроса
// и возвращает короткую ссылку.
func (h *Handler) CreateShortLink(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	slugURL, err := h.shortener.Shorten(string(bodyBytes))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, slugURL)
}

// Redirect перенаправляет короткую ссылку на исходный URL.
func (h *Handler) Redirect(c *gin.Context) {
	slug := c.Param("shortLink")

	url, ok := h.shortener.Resolve(slug)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) CreateShortLinkFromJSON(c *gin.Context) {
	var request model.ShortenRequest

	if err := c.ShouldBindJSON(&request); err != nil || request.URL == "" {
		respondJSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	slugURL, err := h.shortener.Shorten(request.URL)
	if err != nil {
		respondJSON(c, http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	respondJSON(c, http.StatusCreated, model.ShortenResponse{SlugURL: slugURL})
}

// Default отвечает на некорректные запросы.
func (h *Handler) Default(c *gin.Context) {
	c.Status(http.StatusBadRequest)
}

// respondJSON маршалит payload
// и отправляет его как JSON-ответ со статусом status.
// Не используем c.JSON, чтобы избежать автоматического
// добавления Content-Type: application/json; charset=utf-8.
func respondJSON(c *gin.Context, status int, payload any) {
	jsonData, _ := json.Marshal(payload)
	c.Data(status, "application/json", jsonData)
}
