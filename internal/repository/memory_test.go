package repository

import (
	"testing"
)

const (
	testURL  = "https://practicum.yandex.ru/"
	testSlug = "d979ee5b-4f6b-4a1f-8c3d-0f2f1b2c3d4e"
)

func TestInMemoryStorageExists(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantSlug string
		found    bool
	}{
		{
			name:     "positive: existing URL",
			url:      testURL,
			wantSlug: testSlug,
			found:    true,
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
			storage, _ := NewInMemoryStorage()
			storage.SaveURL(testURL, testSlug)

			got, found := storage.Exists(tt.url)

			if found != tt.found {
				t.Errorf("found: got %v, want %v", found, tt.found)
			}
			if got != tt.wantSlug {
				t.Errorf("slug: got %q, want %q", got, tt.wantSlug)
			}
		})
	}
}

func TestInMemoryStorageGetURL(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		wantURL string
		found   bool
	}{
		{
			name:    "positive: existing slug",
			slug:    testSlug,
			wantURL: testURL,
			found:   true,
		},
		{
			name:  "negative: unknown slug",
			slug:  "00000000-0000-0000-0000-000000000000",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, _ := NewInMemoryStorage()
			storage.SaveURL(testURL, testSlug)

			got, found := storage.GetURL(tt.slug)

			if found != tt.found {
				t.Errorf("found: got %v, want %v", found, tt.found)
			}
			if got != tt.wantURL {
				t.Errorf("URL: got %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestInMemoryStorageSaveURL(t *testing.T) {
	storage, _ := NewInMemoryStorage()

	storage.SaveURL(testURL, testSlug)
	got, found := storage.Exists(testURL)
	if !found {
		t.Fatal("URL not found after creation")
	}
	if got != testSlug {
		t.Errorf("slug: got %q, want %q", got, testSlug)
	}

	gotURL, found := storage.GetURL(testSlug)
	if !found {
		t.Fatal("slug not found after creation")
	}
	if gotURL != testURL {
		t.Errorf("URL: got %q, want %q", gotURL, testURL)
	}
}

func TestInMemoryStorageCreateShortUrlOverwrite(t *testing.T) {
	storage, _ := NewInMemoryStorage()

	storage.SaveURL(testURL, testSlug)

	const newSlug = "11111111"
	storage.SaveURL(testURL, newSlug)

	got, found := storage.Exists(testURL)
	if !found {
		t.Fatal("URL not found after overwrite")
	}
	if got != newSlug {
		t.Errorf("slug after overwrite: got %q, want %q", got, newSlug)
	}

	if _, found := storage.GetURL(testSlug); found {
		t.Errorf("old short URL %q should not resolve after overwrite", testSlug)
	}
}
