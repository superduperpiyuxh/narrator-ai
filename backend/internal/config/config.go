package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabasePath string
	DataDir      string
	OpenRouterKey string
	JWTSecret    string
	CORSOrigins  []string
}

func Load() *Config {
	// Load .env file if it exists (no error if missing)
	godotenv.Load()

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabasePath:  getEnv("DATABASE_PATH", "./narratorai.db"),
		DataDir:       getEnv("DATA_DIR", "../data/sample_json_20260301"),
		OpenRouterKey: getEnv("OPENROUTER_API_KEY", ""),
		JWTSecret:     getJWTSecret(),
		CORSOrigins:   getCORSOrigins(),
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

func getCORSOrigins() []string {
	v := getEnv("CORS_ORIGINS", "")
	if v == "" {
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}
	return strings.Split(v, ",")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
