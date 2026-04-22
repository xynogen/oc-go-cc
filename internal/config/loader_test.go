package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	cfgJSON := `{
		"api_key": "test-key-123",
		"host": "0.0.0.0",
		"port": 8080,
		"upstream": {
			"base_url": "https://custom.url/v1",
			"timeout_ms": 60000
		},
		"logging": {
			"level": "debug",
			"requests": true
		}
	}`

	if err := os.WriteFile(cfgPath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	os.Setenv("OGC_CONFIG", cfgPath)
	defer os.Unsetenv("OGC_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIKey != "test-key-123" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "test-key-123")
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Upstream.BaseURL != "https://custom.url/v1" {
		t.Errorf("BaseURL = %q, want %q", cfg.Upstream.BaseURL, "https://custom.url/v1")
	}
	if cfg.Upstream.TimeoutMs != 60000 {
		t.Errorf("TimeoutMs = %d, want %d", cfg.Upstream.TimeoutMs, 60000)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.Logging.Level, "debug")
	}
	if !cfg.Logging.Requests {
		t.Error("Logging.Requests = false, want true")
	}
}

func TestLoadMissingAPIKey(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	cfgJSON := `{"host": "127.0.0.1"}`
	if err := os.WriteFile(cfgPath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	os.Setenv("OGC_CONFIG", cfgPath)
	defer os.Unsetenv("OGC_CONFIG")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing API key, got nil")
	}
}

func TestEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	cfgJSON := `{"api_key": "file-key"}`
	if err := os.WriteFile(cfgPath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	os.Setenv("OGC_CONFIG", cfgPath)
	os.Setenv("OGC_API_KEY", "env-key")
	os.Setenv("OGC_HOST", "env-host")
	os.Setenv("OGC_PORT", "9999")
	os.Setenv("OGC_OPENAI_BASE", "https://env-url/v1")
	os.Setenv("OGC_LOG_LEVEL", "warn")
	defer func() {
		os.Unsetenv("OGC_CONFIG")
		os.Unsetenv("OGC_API_KEY")
		os.Unsetenv("OGC_HOST")
		os.Unsetenv("OGC_PORT")
		os.Unsetenv("OGC_OPENAI_BASE")
		os.Unsetenv("OGC_LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "env-key")
	}
	if cfg.Host != "env-host" {
		t.Errorf("Host = %q, want %q", cfg.Host, "env-host")
	}
	if cfg.Port != 9999 {
		t.Errorf("Port = %d, want %d", cfg.Port, 9999)
	}
	if cfg.Upstream.BaseURL != "https://env-url/v1" {
		t.Errorf("BaseURL = %q, want %q", cfg.Upstream.BaseURL, "https://env-url/v1")
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("LogLevel = %q, want %q", cfg.Logging.Level, "warn")
	}
}

func TestDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Minimal config — only API key, everything else should default.
	cfgJSON := `{"api_key": "test-key"}`
	if err := os.WriteFile(cfgPath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	os.Setenv("OGC_CONFIG", cfgPath)
	defer os.Unsetenv("OGC_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != defaultHost {
		t.Errorf("Host = %q, want %q", cfg.Host, defaultHost)
	}
	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.Upstream.BaseURL != defaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.Upstream.BaseURL, defaultBaseURL)
	}
	if cfg.Upstream.TimeoutMs != defaultTimeoutMs {
		t.Errorf("TimeoutMs = %d, want %d", cfg.Upstream.TimeoutMs, defaultTimeoutMs)
	}
	if cfg.Logging.Level != defaultLogLevel {
		t.Errorf("LogLevel = %q, want %q", cfg.Logging.Level, defaultLogLevel)
	}
}

func TestInterpolateEnvVars(t *testing.T) {
	os.Setenv("TEST_SECRET", "my-secret-value")
	defer os.Unsetenv("TEST_SECRET")

	input := `{"api_key": "${TEST_SECRET}", "host": "${UNSET_VAR:-fallback}"}`
	result := interpolateEnvVars(input)

	want := `{"api_key": "my-secret-value", "host": "${UNSET_VAR:-fallback}"}`
	if result != want {
		t.Errorf("interpolateEnvVars() = %q, want %q", result, want)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/some/path", filepath.Join(home, "some/path")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
