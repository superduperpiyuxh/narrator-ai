package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

type Config struct {
	Port         string
	DatabasePath string
	DataDir      string
	OpenRouterKey string
	JWTSecret    string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabasePath:  getEnv("DATABASE_PATH", "./narratorai.db"),
		DataDir:       getEnv("DATA_DIR", "../data/sample_json_20260301"),
		OpenRouterKey: getEnv("OPENROUTER_API_KEY", ""),
		JWTSecret:     getJWTSecret(),
	}
}

func getJWTSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	// Generate random JWT secret for demo mode compatibility
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate JWT secret: %v", err)
	}
	secret := hex.EncodeToString(b)
	log.Printf("WARNING: No JWT_SECRET set. Generated random secret for this session.")
	return secret
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
