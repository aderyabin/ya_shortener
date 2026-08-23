package repository

import (
	"testing"
)

const (
	testURL      = "https://practicum.yandex.ru/"
	testShortUrl = "d979ee5b-4f6b-4a1f-8c3d-0f2f1b2c3d4e"
)

func TestInMemoryStorageGetShortUrl(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantShortUrl string
		found        bool
	}{
		{
			name:         "positive: existing URL",
			url:          testURL,
			wantShortUrl: testShortUrl,
			found:        true,
		},
		{
			name:  "negative: unknown URL",
			url:   "https://unknown.example.com/",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Подготаливаем данные
			storage := NewInMemoryStorage()
			storage.CreateShortUrl(testURL, testShortUrl)

			got, found := storage.GetShortUrl(tt.url)

			if found != tt.found {
				t.Errorf("found: got %v, want %v", found, tt.found)
			}
			if got != tt.wantShortUrl {
				t.Errorf("short URL: got %q, want %q", got, tt.wantShortUrl)
			}
		})
	}
}

func TestInMemoryStorageGetURL(t *testing.T) {
	tests := []struct {
		name     string
		shortUrl string
		wantUrl  string
		found    bool
	}{
		{
			name:     "positive: existing short URL",
			shortUrl: testShortUrl,
			wantUrl:  testURL,
			found:    true,
		},
		{
			name:     "negative: unknown short URL",
			shortUrl: "00000000-0000-0000-0000-000000000000",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewInMemoryStorage()
			storage.CreateShortUrl(testURL, testShortUrl)

			got, found := storage.GetURL(tt.shortUrl)

			if found != tt.found {
				t.Errorf("found: got %v, want %v", found, tt.found)
			}
			if got != tt.wantUrl {
				t.Errorf("URL: got %q, want %q", got, tt.wantUrl)
			}
		})
	}
}

func TestInMemoryStorageCreateShortUrl(t *testing.T) {
	storage := NewInMemoryStorage()

	storage.CreateShortUrl(testURL, testShortUrl)
	got, found := storage.GetShortUrl(testURL)
	if !found {
		t.Fatal("URL not found after creation")
	}
	if got != testShortUrl {
		t.Errorf("short URL: got %q, want %q", got, testShortUrl)
	}

	gotURL, found := storage.GetURL(testShortUrl)
	if !found {
		t.Fatal("short URL not found after creation")
	}
	if gotURL != testURL {
		t.Errorf("URL: got %q, want %q", gotURL, testURL)
	}
}

func TestInMemoryStorageCreateShortUrlOverwrite(t *testing.T) {
	storage := NewInMemoryStorage()

	storage.CreateShortUrl(testURL, testShortUrl)

	const newShortUrl = "11111111"
	storage.CreateShortUrl(testURL, newShortUrl)

	got, found := storage.GetShortUrl(testURL)
	if !found {
		t.Fatal("URL not found after overwrite")
	}
	if got != newShortUrl {
		t.Errorf("short URL after overwrite: got %q, want %q", got, newShortUrl)
	}

	if _, found := storage.GetURL(testShortUrl); found {
		t.Errorf("old short URL %q should not resolve after overwrite", testShortUrl)
	}
}
