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

const testJSONBody = `{"url":"https://practicum.yandex.ru/"}`

// gzipCompress сжимает body и возвращает буфер со сжатыми данными.
func gzipCompress(t *testing.T, body string) *bytes.Buffer {
	t.Helper()

	var compressed bytes.Buffer
	gw := gzip.NewWriter(&compressed)
	if _, err := gw.Write([]byte(body)); err != nil {
		t.Fatalf("failed to write gzip body: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return &compressed
}

type decompressorTestStruct struct {
	name            string
	requestBody     string
	compressBody    bool   // сжать ли тело запроса перед отправкой
	contentEncoding string // заголовок Content-Encoding запроса
	wantCode        int
	wantBody        string // тело, которое должен увидеть хендлер
	wantHandlerCall bool
}

func TestGzipDecompressor(t *testing.T) {
	tests := []decompressorTestStruct{
		{
			name:            "positive: gzipped request body",
			requestBody:     testJSONBody,
			compressBody:    true,
			contentEncoding: "gzip",
			wantCode:        http.StatusOK,
			wantBody:        testJSONBody,
			wantHandlerCall: true,
		},
		{
			name:            "positive: plain request without encoding",
			requestBody:     testJSONBody,
			contentEncoding: "",
			wantCode:        http.StatusOK,
			wantBody:        testJSONBody,
			wantHandlerCall: true,
		},
		{
			name:            "negative: broken gzip body",
			requestBody:     "not-gzip",
			contentEncoding: "gzip",
			wantCode:        http.StatusBadRequest,
			wantHandlerCall: false,
		},
		{
			name:            "negative: empty gzip body",
			requestBody:     "",
			contentEncoding: "gzip",
			wantCode:        http.StatusBadRequest,
			wantHandlerCall: false,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			receivedCode, receivedBody, handlerCalled := prepareDecompressorTestResponse(t, tt)

			if receivedCode != tt.wantCode {
				t.Errorf("status code: got %d, want %d", receivedCode, tt.wantCode)
				if handlerCalled != tt.wantHandlerCall {
					t.Errorf("handler called: got %v, want %v", handlerCalled, tt.wantHandlerCall)
				}
				if tt.wantHandlerCall && receivedBody != tt.wantBody {
					t.Errorf("body: got %q, want %q", receivedBody, tt.wantBody)
				}
			}
		})
	}
}

// prepareDecompressorTestResponse прогоняет запрос из кейса tt через
// GzipDecompressor и возвращает записанный ответ вместе с тем,
// что увидел хендлер-шпион.
func prepareDecompressorTestResponse(t *testing.T, tt decompressorTestStruct) (int, string, bool) {
	t.Helper()

	// Подготавливаем тело запроса
	body := []byte(tt.requestBody)
	if tt.compressBody {
		body = gzipCompress(t, tt.requestBody).Bytes()
	}

	// Подготавливаем роутер с middleware и хендлером-шпионом
	var receivedBody string
	handlerCalled := false
	handler := func(c *gin.Context) {
		handlerCalled = true
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		receivedBody = string(b)
		c.Status(http.StatusOK)
	}

	router := gin.New()
	router.Use(GzipDecompressor())
	router.POST("/", handler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	if tt.contentEncoding != "" {
		req.Header.Set("Content-Encoding", tt.contentEncoding)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder.Code, receivedBody, handlerCalled
}
