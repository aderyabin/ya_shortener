package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testStoragePath возвращает путь до временного файла хранилища.
func testStoragePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "storage.json")
}

func TestFileStorageExists(t *testing.T) {
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
			storage, err := NewFileStorage(testStoragePath(t))
			if err != nil {
				t.Fatalf("failed to create file storage: %v", err)
			}
			defer storage.Close()

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

func TestFileStorageGetURL(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		wantUrl string
		found   bool
	}{
		{
			name:    "positive: existing slug",
			slug:    testSlug,
			wantUrl: testURL,
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
			storage, err := NewFileStorage(testStoragePath(t))
			if err != nil {
				t.Fatalf("failed to create file storage: %v", err)
			}
			defer storage.Close()

			storage.SaveURL(testURL, testSlug)

			got, found := storage.GetURL(tt.slug)

			if found != tt.found {
				t.Errorf("found: got %v, want %v", found, tt.found)
			}
			if got != tt.wantUrl {
				t.Errorf("URL: got %q, want %q", got, tt.wantUrl)
			}
		})
	}
}

func TestFileStorageSaveURL(t *testing.T) {
	path := testStoragePath(t)

	storage, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}
	defer storage.Close()

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

	// Проверяем, что запись попала в файл в формате JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read storage file: %v", err)
	}
	wantLine := `{"uuid":1,"short_url":"` + testSlug + `","original_url":"` + testURL + `"}`
	if !strings.Contains(string(data), wantLine) {
		t.Errorf("file content: got %q, want line %q", string(data), wantLine)
	}
}

func TestFileStorageSaveURLOverwrite(t *testing.T) {
	storage, err := NewFileStorage(testStoragePath(t))
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}
	defer storage.Close()

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

func TestFileStorageRestoreAfterReopen(t *testing.T) {
	path := testStoragePath(t)

	// Первая «сессия»: сохраняем ссылку, перезаписываем её и закрываем хранилище
	first, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}
	first.SaveURL(testURL, testSlug)
	const newSlug = "11111111"
	first.SaveURL(testURL, newSlug)
	if err := first.Close(); err != nil {
		t.Fatalf("failed to close storage: %v", err)
	}

	// Вторая «сессия»: данные должны восстановиться из файла,
	// при дублировании URL побеждает последняя запись
	second, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("failed to reopen file storage: %v", err)
	}
	defer second.Close()

	got, found := second.Exists(testURL)
	if !found {
		t.Fatal("URL not restored after reopen")
	}
	if got != newSlug {
		t.Errorf("slug after restore: got %q, want %q", got, newSlug)
	}

	gotURL, found := second.GetURL(newSlug)
	if !found {
		t.Fatal("slug not restored after reopen")
	}
	if gotURL != testURL {
		t.Errorf("URL: got %q, want %q", gotURL, testURL)
	}

	if _, found := second.GetURL(testSlug); found {
		t.Errorf("old short URL %q should not resolve after restore", testSlug)
	}
}

func TestFileStorageRestore(t *testing.T) {
	tests := []struct {
		name        string
		createFile  bool              // создавать ли файл с содержимым до открытия хранилища
		fileContent string            // содержимое файла (JSON Lines)
		wantRecords map[string]string // записи slug -> URL, ожидаемые после restore
		wantErr     bool
	}{
		{
			name:        "positive: single record",
			createFile:  true,
			fileContent: `{"uuid":1,"short_url":"` + testSlug + `","original_url":"` + testURL + `"}` + "\n",
			wantRecords: map[string]string{testSlug: testURL},
		},
		{
			name:       "positive: multiple records",
			createFile: true,
			fileContent: `{"uuid":1,"short_url":"` + testSlug + `","original_url":"` + testURL + `"}` + "\n" +
				`{"uuid":2,"short_url":"11111111","original_url":"https://yandex.ru/"}` + "\n",
			wantRecords: map[string]string{
				testSlug:   testURL,
				"11111111": "https://yandex.ru/",
			},
		},
		{
			name:        "positive: empty lines are skipped",
			createFile:  true,
			fileContent: "\n" + `{"uuid":1,"short_url":"` + testSlug + `","original_url":"` + testURL + `"}` + "\n\n",
			wantRecords: map[string]string{testSlug: testURL},
		},
		{
			name:        "positive: empty file",
			createFile:  true,
			fileContent: "",
			wantRecords: map[string]string{},
		},
		{
			name:        "positive: missing file is created",
			createFile:  false,
			wantRecords: map[string]string{},
		},
		{
			name:        "negative: broken JSON",
			createFile:  true,
			fileContent: "{not-json\n",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготавливаем файл хранилища
			path := testStoragePath(t)
			if tt.createFile {
				if err := os.WriteFile(path, []byte(tt.fileContent), 0o644); err != nil {
					t.Fatalf("failed to write storage file: %v", err)
				}
			}

			storage, err := NewFileStorage(path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to create file storage: %v", err)
			}
			defer storage.Close()

			// Проверяем восстановленные записи в обе стороны
			if got := storage.Size(); got != len(tt.wantRecords) {
				t.Errorf("size: got %d, want %d", got, len(tt.wantRecords))
			}
			for slug, wantURL := range tt.wantRecords {
				gotURL, found := storage.GetURL(slug)
				if !found {
					t.Errorf("slug %q not restored", slug)
					continue
				}
				if gotURL != wantURL {
					t.Errorf("URL: got %q, want %q", gotURL, wantURL)
				}

				gotSlug, found := storage.Exists(wantURL)
				if !found {
					t.Errorf("URL %q not restored", wantURL)
					continue
				}
				if gotSlug != slug {
					t.Errorf("slug: got %q, want %q", gotSlug, slug)
				}
			}

			// Отсутствующий файл должен быть создан
			if !tt.createFile {
				if _, err := os.Stat(path); err != nil {
					t.Errorf("storage file not created: %v", err)
				}
			}
		})
	}
}
