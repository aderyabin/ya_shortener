package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

const (
	testJSONResponse  = `{"result":"ok"}`
	testHTMLResponse  = "<html><body>ok</body></html>"
	testPlainResponse = "hello"
)

// gunzip распаковывает тело ответа и возвращает исходные данные.
func gunzip(t *testing.T, body *bytes.Buffer) string {
	t.Helper()

	gr, err := gzip.NewReader(body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read decompressed body: %v", err)
	}

	return string(decompressed)
}

func TestGzipCompressor(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		contentType    string
		responseBody   string
		acceptEncoding string
		wantEncoding   string // ожидаемый Content-Encoding ответа
	}{
		{
			name:           "positive: JSON response compressed",
			statusCode:     http.StatusCreated,
			contentType:    "application/json",
			responseBody:   testJSONResponse,
			acceptEncoding: "gzip",
			wantEncoding:   "gzip",
		},
		{
			name:           "positive: HTML response compressed",
			statusCode:     http.StatusOK,
			contentType:    "text/html; charset=utf-8",
			responseBody:   testHTMLResponse,
			acceptEncoding: "gzip",
			wantEncoding:   "gzip",
		},
		{
			name:           "negative: text/plain not compressed",
			statusCode:     http.StatusOK,
			contentType:    "text/plain; charset=utf-8",
			responseBody:   testPlainResponse,
			acceptEncoding: "gzip",
			wantEncoding:   "",
		},
		{
			name:           "negative: no Accept-Encoding header",
			statusCode:     http.StatusOK,
			contentType:    "application/json",
			responseBody:   testJSONResponse,
			acceptEncoding: "",
			wantEncoding:   "",
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготавливаем роутер с middleware и хендлером,
			// отдающим ответ заданного типа
			handler := func(c *gin.Context) {
				c.Data(tt.statusCode, tt.contentType, []byte(tt.responseBody))
			}

			router := gin.New()
			router.Use(GzipCompressor())
			router.GET("/", handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("status code: got %d, want %d", w.Code, tt.statusCode)
			}
			if got := w.Header().Get("Content-Encoding"); got != tt.wantEncoding {
				t.Errorf("Content-Encoding: got %q, want %q", got, tt.wantEncoding)
			}

			// Сжатое тело распаковываем, несжатое сравниваем как есть
			gotBody := w.Body.String()
			if tt.wantEncoding == "gzip" {
				gotBody = gunzip(t, w.Body)
			}
			if gotBody != tt.responseBody {
				t.Errorf("body: got %q, want %q", gotBody, tt.responseBody)
			}
		})
	}
}
