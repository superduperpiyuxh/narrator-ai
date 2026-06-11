package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.DatabasePath != "./narratorai.db" {
		t.Errorf("expected ./narratorai.db, got %s", cfg.DatabasePath)
	}
}

func TestLoadWithEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	val := getEnv("TEST_KEY", "default")
	if val != "test_value" {
		t.Errorf("expected test_value, got %s", val)
	}

	val = getEnv("NONEXISTENT_KEY", "default")
	if val != "default" {
		t.Errorf("expected default, got %s", val)
	}
}

func TestGetCORSOrigins_Default(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")
	origins := getCORSOrigins()
	if len(origins) != 2 {
		t.Errorf("expected 2 default origins, got %d", len(origins))
	}
	if origins[0] != "http://localhost:3000" {
		t.Errorf("expected localhost:3000, got %s", origins[0])
	}
	if origins[1] != "http://localhost:5173" {
		t.Errorf("expected localhost:5173, got %s", origins[1])
	}
}

func TestGetCORSOrigins_Custom(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "https://example.com,https://app.example.com")
	defer os.Unsetenv("CORS_ORIGINS")

	origins := getCORSOrigins()
	if len(origins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(origins))
	}
	if origins[0] != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", origins[0])
	}
}

func TestGetCORSOrigins_Single(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "https://myapp.com")
	defer os.Unsetenv("CORS_ORIGINS")

	origins := getCORSOrigins()
	if len(origins) != 1 {
		t.Errorf("expected 1 origin, got %d", len(origins))
	}
	if origins[0] != "https://myapp.com" {
		t.Errorf("expected https://myapp.com, got %s", origins[0])
	}
}

func TestGetJWTSecret_FromEnv(t *testing.T) {
	os.Setenv("JWT_SECRET", "my-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	secret := getJWTSecret()
	if secret != "my-secret-key" {
		t.Errorf("expected 'my-secret-key', got '%s'", secret)
	}
}

func TestGetJWTSecret_Generated(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	secret := getJWTSecret()
	if secret == "" {
		t.Error("expected non-empty generated secret")
	}
	if len(secret) != 64 {
		t.Errorf("expected 64 char hex string, got %d chars", len(secret))
	}
}

func TestLoad_CORSOrigins(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "https://test.com")
	defer os.Unsetenv("CORS_ORIGINS")

	cfg := Load()
	if len(cfg.CORSOrigins) != 1 {
		t.Errorf("expected 1 origin, got %d", len(cfg.CORSOrigins))
	}
}

func TestLoad_FullConfig(t *testing.T) {
	os.Setenv("PORT", "3001")
	os.Setenv("DATABASE_PATH", "/tmp/test.db")
	os.Setenv("DATA_DIR", "/tmp/data")
	os.Setenv("OPENROUTER_API_KEY", "test-key-123")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("CORS_ORIGINS", "https://example.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("DATA_DIR")
		os.Unsetenv("OPENROUTER_API_KEY")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("CORS_ORIGINS")
	}()

	cfg := Load()
	if cfg.Port != "3001" {
		t.Errorf("expected 3001, got %s", cfg.Port)
	}
	if cfg.DatabasePath != "/tmp/test.db" {
		t.Errorf("expected /tmp/test.db, got %s", cfg.DatabasePath)
	}
	if cfg.DataDir != "/tmp/data" {
		t.Errorf("expected /tmp/data, got %s", cfg.DataDir)
	}
	if cfg.OpenRouterKey != "test-key-123" {
		t.Errorf("expected test-key-123, got %s", cfg.OpenRouterKey)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected test-secret, got %s", cfg.JWTSecret)
	}
	if !strings.Contains(cfg.CORSOrigins[0], "example.com") {
		t.Errorf("expected example.com, got %s", cfg.CORSOrigins[0])
	}
}
