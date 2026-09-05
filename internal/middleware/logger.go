// Package middleware содержит middleware для HTTP-сервера.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger возвращает middleware, логирующий сведения о запросе и ответе:
// метод, URI, статус, размер ответа и время обработки.
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("uri", c.Request.RequestURI),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("duration", duration),
		)
	}
}
