package middleware

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// compressedHeaders — заголовки, которые указывают на сжатый контент.
var compressedHeaders = []struct {
	headerName  string
	headerValue string
}{
	{"Content-Encoding", "gzip"},
}

// GzipDecompressor распаковывает запросы с заголовком Content-Encoding: gzip
func GzipDecompressor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isCompressed(c.Request) {
			c.Next()
			return
		}

		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer gz.Close()

		c.Request.Body = io.NopCloser(gz)

		c.Next()
	}
}

func isCompressed(r *http.Request) bool {
	for _, h := range compressedHeaders {
		if r.Header.Get(h.headerName) == h.headerValue {
			return true
		}
	}
	return false
}
