package middleware

import (
	"compress/gzip"
	"strings"

	"github.com/gin-gonic/gin"
)

// compressibleContentTypes — типы контента, для которых ответ сжимается.
var compressibleContentTypes = []string{"application/json", "text/html"}

// GzipCompressor сжимает ответ, если клиент передал gzip в Accept-Encoding.
//
// Middleware подменяет c.Writer на gzipWriter: на момент входа в middleware
// Content-Type ответа ещё неизвестен (его выставит хендлер), поэтому решение
// о сжатии откладывается до момента записи заголовков — см. gzipWriter.
func GzipCompressor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Клиент не поддерживает gzip — отдаём ответ как есть.
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Подменяем writer: все записи хендлера пройдут через gzipWriter.
		gw := &gzipWriter{ResponseWriter: c.Writer, gz: gzip.NewWriter(c.Writer)}
		defer gw.Close()

		c.Writer = gw

		c.Next()
	}
}

// gzipWriter оборачивает gin.ResponseWriter, чтобы перехватывать запись ответа.
// Это единственная точка, где можно одновременно:
//   - увидеть Content-Type, который выставил хендлер;
//   - успеть добавить Content-Encoding: gzip до отправки заголовков клиенту;
//   - перенаправить тело ответа в gzip-компрессор.
type gzipWriter struct {
	gin.ResponseWriter
	gz          *gzip.Writer
	compressing bool // решение о сжатии, принятое в WriteHeader
}

// WriteHeader вызывается при отправке статуса — к этому моменту хендлер
// уже выставил Content-Type, а заголовки клиенту ещё не ушли.
// Здесь мы решаем, сжимать ли ответ, и успеваем изменить заголовки:
// Content-Length удаляем, так как длина сжатого тела заранее неизвестна.
func (w *gzipWriter) WriteHeader(code int) {
	if isCompressibleResponseType(w.Header().Get("Content-Type")) {
		w.compressing = true
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
	}
	w.ResponseWriter.WriteHeader(code)
}

// WriteHeaderOriginal нужен, чтобы gin при рендеринге вызывал наш WriteHeader.
// Без переопределения gin записал бы заголовки через встроенный writer
// в обход WriteHeader — и Content-Encoding: gzip не попал бы в ответ.
func (w *gzipWriter) WriteHeaderOriginal() {
	w.WriteHeader(w.Status())
}

// Write перенаправляет тело в gzip.Writer, если сжатие включено,
// иначе пишет напрямую в исходный writer.
func (w *gzipWriter) Write(p []byte) (int, error) {
	if !w.Written() {
		w.WriteHeaderOriginal()
	}
	if w.compressing {
		return w.gz.Write(p)
	}
	return w.ResponseWriter.Write(p)
}

// Close завершает сжатый поток, если сжатие было включено.
func (w *gzipWriter) Close() {
	if w.compressing {
		_ = w.gz.Close()
	}
}

// isCompressibleResponseType проверяет, входит ли contentType в список доступных для сжатия типов.
func isCompressibleResponseType(contentType string) bool {
	for _, t := range compressibleContentTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}
	return false
}
