package service

import (
	"errors"
	"testing"

	"shortener/internal/repository"

	"github.com/google/uuid"
)

const (
	testBaseURL  = "http://localhost:8080"
	testShortUrl = "d979ee5b-4f6b-4a1f-8c3d-0f2f1b2c3d4e"
	testURL      = "https://practicum.yandex.ru/"
)

func stubShortUrlGenerator() (string, error) {
	return testShortUrl, nil
}

func TestShortenerShorten(t *testing.T) {
	tests := []struct {
		name        string
		genShortUrl ShortUrlGenerator
		url         string
		want        string
		wantErr     bool
	}{
		{
			name:        "positive: returns short link with generated short URL",
			genShortUrl: stubShortUrlGenerator,
			url:         testURL,
			want:        testBaseURL + "/" + testShortUrl,
		},
		{
			name: "negative: ID generator failure",
			genShortUrl: func() (string, error) {
				return "", errors.New("random source unavailable")
			},
			url:     testURL,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewInMemoryStorage()
			s := NewShortener(storage, testBaseURL, tt.genShortUrl)

			got, err := s.Shorten(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("short URL: got %q, want %q", got, tt.want)
			}

			saved, ok := storage.GetURL(testShortUrl)
			if !ok {
				t.Fatal("URL was not saved to storage")
			}
			if saved != tt.url {
				t.Errorf("saved URL: got %q, want %q", saved, tt.url)
			}
		})
	}
}

func TestShortenerShortenDuplicate(t *testing.T) {
	calls := 0
	genShortUrl := func() (string, error) {
		calls++
		return testShortUrl, nil
	}

	storage := repository.NewInMemoryStorage()
	s := NewShortener(storage, testBaseURL, genShortUrl)

	first, err := s.Shorten(testURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := s.Shorten(testURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Errorf("duplicate URL produced different short links: %q and %q", first, second)
	}

	if calls != 1 {
		t.Errorf("ID generator called %d times, want 1 (UUID should be reused)", calls)
	}
}

func TestShortenerResolve(t *testing.T) {
	tests := []struct {
		name     string
		shortUrl string
		want     string
		found    bool
	}{
		{
			name:     "positive: existing short URL",
			shortUrl: testShortUrl,
			want:     testURL,
			found:    true,
		},
		{
			name:     "negative: unknown short URL",
			shortUrl: "00000000",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewInMemoryStorage()
			s := NewShortener(storage, testBaseURL, stubShortUrlGenerator)
			storage.CreateShortUrl(testURL, testShortUrl)

			got, found := s.Resolve(tt.shortUrl)

			if found != tt.found {
				t.Errorf("found: got %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("URL: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRandomUUID(t *testing.T) {
	id1, err := RandomUUID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := uuid.Parse(id1); err != nil {
		t.Errorf("generated ID %q is not a valid UUID: %v", id1, err)
	}

	id2, err := RandomUUID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 == id2 {
		t.Errorf("two generated UUIDs are equal: %q", id1)
	}
}
