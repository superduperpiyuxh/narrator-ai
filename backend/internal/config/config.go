package config

import "os"

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
		JWTSecret:     getEnv("JWT_SECRET", "narrator-ai-jwt-secret-change-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
