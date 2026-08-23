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
		wantServerAddr string
		wantBaseURL    string
	}{
		{
			name:           "Sets default values",
			wantServerAddr: "localhost:8080",
			wantBaseURL:    "http://localhost:8080",
		},
		{
			name:           "Sets values from flags",
			args:           []string{"-a", "localhost:9090", "-b", "http://example.com"},
			wantServerAddr: "localhost:9090",
			wantBaseURL:    "http://example.com",
		},
		{
			name:           "Sets values from environment variables",
			envServer:      "localhost:7070",
			wantServerAddr: "localhost:7070",
			envBase:        "http://example.com",
			wantBaseURL:    "http://example.com",
		},
		{
			name:           "Envs prioritized over flags",
			args:           []string{"-a", "localhost:9090", "-b", "http://example.com"},
			envServer:      "localhost:7070",
			envBase:        "http://env.example.com",
			wantServerAddr: "localhost:7070",
			wantBaseURL:    "http://env.example.com",
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

			if tt.envServer != "" {
				t.Setenv("SERVER_ADDRESS", tt.envServer)
			}
			if tt.envBase != "" {
				t.Setenv("BASE_URL", tt.envBase)
			}

			cfg := NewConfig()

			if cfg.ServerAddress != tt.wantServerAddr {
				t.Errorf("ServerAddress: got %q, want %q", cfg.ServerAddress, tt.wantServerAddr)
			}
			if cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, tt.wantBaseURL)
			}
		})
	}
}
