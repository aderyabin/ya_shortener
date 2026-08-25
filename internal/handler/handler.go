package handler

import (
	"io"
	"net/http"
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

	slug, err := h.shortener.Shorten(string(bodyBytes))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, slug)
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

// Default отвечает на некорректные запросы.
func (h *Handler) Default(c *gin.Context) {
	c.Status(http.StatusBadRequest)
}
