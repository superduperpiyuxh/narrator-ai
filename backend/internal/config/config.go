package config

import "os"

type Config struct {
	Port         string
	DatabasePath string
	GraylogURL   string
	GraylogUser  string
	GraylogPass  string
	DataDir      string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./narratorai.db"),
		GraylogURL:   getEnv("GRAYLOG_URL", "http://localhost:9000"),
		GraylogUser:  getEnv("GRAYLOG_USER", "admin"),
		GraylogPass:  getEnv("GRAYLOG_PASS", "admin"),
		DataDir:      getEnv("DATA_DIR", "../data/sample_json_20260301"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
