package config

import (
	"flag"
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		envServer      string
		envBase        string
		envFile        string
		wantServerAddr string
		wantBaseURL    string
		wantFilePath   string
	}{
		{
			name:           "Sets default values",
			wantServerAddr: "localhost:8080",
			wantBaseURL:    "http://localhost:8080",
			wantFilePath:   "storage.json",
		},
		{
			name:           "Sets values from flags",
			args:           []string{"-a", "localhost:9090", "-b", "http://example.com", "-f", "/tmp/urls.json"},
			wantServerAddr: "localhost:9090",
			wantBaseURL:    "http://example.com",
			wantFilePath:   "/tmp/urls.json",
		},
		{
			name:           "Sets values from environment variables",
			envServer:      "localhost:7070",
			wantServerAddr: "localhost:7070",
			envBase:        "http://example.com",
			wantBaseURL:    "http://example.com",
			envFile:        "/tmp/env-urls.json",
			wantFilePath:   "/tmp/env-urls.json",
		},
		{
			name:           "Envs prioritized over flags",
			args:           []string{"-a", "localhost:9090", "-b", "http://example.com", "-f", "/tmp/urls.json"},
			envServer:      "localhost:7070",
			envBase:        "http://env.example.com",
			envFile:        "/tmp/env-urls.json",
			wantServerAddr: "localhost:7070",
			wantBaseURL:    "http://env.example.com",
			wantFilePath:   "/tmp/env-urls.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// flag.CommandLine глобальный: сбрасываем его, иначе повторная
			// регистрация флагов в NewConfig вызовет panic "flag redefined".
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			os.Args = append([]string{"test"}, tt.args...)

			// Гарантируем отсутствие переменных, если кейс их не задаёт.
			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
			os.Unsetenv("FILE_STORAGE_PATH")

			if tt.envServer != "" {
				t.Setenv("SERVER_ADDRESS", tt.envServer)
			}
			if tt.envBase != "" {
				t.Setenv("BASE_URL", tt.envBase)
			}
			if tt.envFile != "" {
				t.Setenv("FILE_STORAGE_PATH", tt.envFile)
			}

			cfg := NewConfig()

			if cfg.ServerAddress != tt.wantServerAddr {
				t.Errorf("ServerAddress: got %q, want %q", cfg.ServerAddress, tt.wantServerAddr)
			}
			if cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, tt.wantBaseURL)
			}
			if cfg.FileStoragePath != tt.wantFilePath {
				t.Errorf("FileStoragePath: got %q, want %q", cfg.FileStoragePath, tt.wantFilePath)
			}
		})
	}
}
