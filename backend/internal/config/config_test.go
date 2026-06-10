package config

import (
	"os"
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
