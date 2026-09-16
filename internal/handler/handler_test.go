package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shortener/internal/repository"
	"shortener/internal/service"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const (
	testBaseURL = "http://example.com"
	testSlug    = "abc12345"
	targetURL   = "https://practicum.yandex.ru/"
)

var responseJSONHeaders = map[string]string{
	"Content-Type": "application/json",
}

func init() {
	gin.SetMode(gin.TestMode) // Отключает debug-логи gin в тестах
}

// newTestHandler собирает Handler с in-memory хранилищем
// и стабом генератора ID, возвращающим фиксированный идентификатор.
func newTestHandler() (*Handler, *service.Shortener) {
	storage, _ := repository.NewInMemoryStorage()
	shortener := service.NewShortener(storage, testBaseURL, func() (string, error) {
		return testSlug, nil
	})
	return NewHandler(shortener, nil), shortener
}

// want описывает ожидаемый результат запроса.
type want struct {
	code            int               // ожидаемый HTTP-статус
	responseBody    string            // ожидаемое тело ответа
	responseHeaders map[string]string // ожидаемые заголовки ответа
	location        string            // ожидаемый заголовок Location
}

// actual описывает параметры входящего запроса.
type actual struct {
	method string // HTTP-метод запроса
	target string // путь или абсолютный URL запроса
	body   string // тело запроса
}

// testCase описывает один табличный тест хендлера.
type testCase struct {
	name   string
	actual actual
	want   want
}

// runTests выполняет таблицу тестов хендлера: каждый кейс
// запускается как сабтест через runTest.
func runTests(t *testing.T, handler gin.HandlerFunc, tests []testCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, handler, tt)
		})
	}
}

// runTest выполняет один табличный тест: вызывает хендлер
// с параметрами из actual и сверяет ответ с want.
func runTest(t *testing.T, handler gin.HandlerFunc, tt testCase) {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(tt.actual.method, tt.actual.target, strings.NewReader(tt.actual.body))
	// Эмулируем параметр маршрута /:shortLink, т.к. хендлер вызывается без роутера.
	c.Params = gin.Params{{Key: "shortLink", Value: strings.TrimPrefix(c.Request.URL.Path, "/")}}

	handler(c)

	res := w.Result()
	defer res.Body.Close()

	if got := c.Writer.Status(); got != tt.want.code {
		t.Errorf("status code: got %d, want %d", got, tt.want.code)
	}

	if tt.want.responseBody != "" {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("failed to read responseBody body: %v", err)
		}

		if got := string(bodyBytes); got != tt.want.responseBody {
			t.Errorf("body: got %q, want %q", got, tt.want.responseBody)
		}
	}

	for key, value := range tt.want.responseHeaders {
		if got := res.Header.Get(key); got != value {
			t.Errorf("header %q: got %q, want %q", key, got, value)
		}
	}

	if tt.want.location != "" {
		if got := res.Header.Get("Location"); got != tt.want.location {
			t.Errorf("Location header: got %q, want %q", got, tt.want.location)
		}
	}
}

func TestCreateShortLinkHandler(t *testing.T) {
	h, _ := newTestHandler()

	tests := []testCase{
		{
			name: "positive: with valid request body",
			actual: actual{
				method: http.MethodPost,
				target: "/",
				body:   targetURL,
			},
			want: want{
				code:         http.StatusCreated,
				responseBody: testBaseURL + "/" + testSlug,
			},
		},
		{
			name: "positive: request scheme does not affect responseBody",
			actual: actual{
				method: http.MethodPost,
				target: "https://example.com/",
				body:   targetURL,
			},
			want: want{
				code:         http.StatusCreated,
				responseBody: testBaseURL + "/" + testSlug,
			},
		},
		{
			name: "negative: empty body",
			actual: actual{
				method: http.MethodPost,
				target: "/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	runTests(t, h.CreateShortLink, tests)
}

func TestCreateShortLinkHandler_GeneratorError(t *testing.T) {
	storage, _ := repository.NewInMemoryStorage()
	shortener := service.NewShortener(storage, testBaseURL, func() (string, error) {
		return "", errors.New("random source unavailable")
	})
	h := NewHandler(shortener, nil)

	runTest(t, h.CreateShortLink, testCase{
		name: "negative: ID generator failure",
		actual: actual{
			method: http.MethodPost,
			target: "/",
			body:   targetURL,
		},
		want: want{
			code: http.StatusInternalServerError,
		},
	})
}

func TestRedirectHandler(t *testing.T) {
	h, shortener := newTestHandler()

	if _, err := shortener.Shorten(targetURL); err != nil {
		t.Fatalf("failed to shorten URL: %v", err)
	}

	tests := []testCase{
		{
			name: "positive: redirect to original URL",
			actual: actual{
				method: http.MethodGet,
				target: "/" + testSlug,
			},
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: targetURL,
			},
		},
		{
			name: "negative: unknown short link",
			actual: actual{
				method: http.MethodGet,
				target: "/unknown",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	runTests(t, h.Redirect, tests)
}

func TestDefaultHandler(t *testing.T) {
	h, _ := newTestHandler()

	tests := []testCase{
		{
			name: "negative: unknown route",
			actual: actual{
				method: http.MethodGet,
				target: "/unknown/path",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "negative: unsupported method on root",
			actual: actual{
				method: http.MethodPut,
				target: "/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	runTests(t, h.Default, tests)
}

func TestCreateShortLinkFromJSONHandler(t *testing.T) {
	h, _ := newTestHandler()

	tests := []testCase{
		{
			name: "positive: valid JSON request",
			actual: actual{
				method: http.MethodPost,
				target: "/api/shorten",
				body:   `{"url":"` + targetURL + `"}`,
			},
			want: want{
				code:            http.StatusCreated,
				responseBody:    `{"result":"` + testBaseURL + "/" + testSlug + `"}`,
				responseHeaders: responseJSONHeaders,
			},
		},
		{
			name: "negative: empty JSON body",
			actual: actual{
				method: http.MethodPost,
				target: "/api/shorten",
				body:   `{"url":""}`,
			},
			want: want{
				code:            http.StatusBadRequest,
				responseHeaders: responseJSONHeaders,
				responseBody:    `{"error":"invalid request body"}`,
			},
		},
		{
			name: "negative: invalid JSON",
			actual: actual{
				method: http.MethodPost,
				target: "/api/shorten",
				body:   `invalid json`,
			},
			want: want{
				code:            http.StatusBadRequest,
				responseHeaders: responseJSONHeaders,
				responseBody:    `{"error":"invalid request body"}`,
			},
		},
	}

	runTests(t, h.CreateShortLinkFromJSON, tests)
}

func TestCreateShortLinkFromJSONHandler_GeneratorError(t *testing.T) {
	storage, _ := repository.NewInMemoryStorage()
	shortener := service.NewShortener(storage, testBaseURL, func() (string, error) {
		return "", errors.New("random source unavailable")
	})
	h := NewHandler(shortener, nil)

	runTest(t, h.CreateShortLinkFromJSON, testCase{
		name: "negative: ID generator failure",
		actual: actual{
			method: http.MethodPost,
			target: "/api/shorten",
			body:   `{"url":"` + targetURL + `"}`,
		},
		want: want{
			code:            http.StatusInternalServerError,
			responseHeaders: responseJSONHeaders,
			responseBody:    `{"error":"internal server error"}`,
		},
	})
}

// mockPinger — тестовая реализация Pinger.
type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestPingHandler(t *testing.T) {
	storage, _ := repository.NewInMemoryStorage()
	shortener := service.NewShortener(storage, testBaseURL, func() (string, error) {
		return testSlug, nil
	})

	tests := []struct {
		name   string
		pinger Pinger
		want   int
	}{
		{
			name:   "negative: no pinger configured",
			pinger: nil,
			want:   http.StatusInternalServerError,
		},
		{
			name:   "negative: pinger returns error",
			pinger: &mockPinger{err: errors.New("db unavailable")},
			want:   http.StatusInternalServerError,
		},
		{
			name:   "positive: pinger succeeds",
			pinger: &mockPinger{err: nil},
			want:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(shortener, tt.pinger)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)

			h.Ping(c)

			if got := c.Writer.Status(); got != tt.want {
				t.Errorf("status code: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPingRoute(t *testing.T) {
	storage, _ := repository.NewInMemoryStorage()
	shortener := service.NewShortener(storage, testBaseURL, func() (string, error) {
		return testSlug, nil
	})

	tests := []struct {
		name   string
		pinger Pinger
		want   int
	}{
		{
			name:   "negative: no pinger configured",
			pinger: nil,
			want:   http.StatusInternalServerError,
		},
		{
			name:   "negative: pinger returns error",
			pinger: &mockPinger{err: errors.New("db unavailable")},
			want:   http.StatusInternalServerError,
		},
		{
			name:   "positive: pinger succeeds",
			pinger: &mockPinger{err: nil},
			want:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(shortener, tt.pinger)

			router := gin.New()
			router.GET("/ping", h.Ping)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if got := w.Code; got != tt.want {
				t.Errorf("status code: got %d, want %d", got, tt.want)
			}
		})
	}
}
