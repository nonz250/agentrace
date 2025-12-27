package config

import "os"

type Config struct {
	Port        string
	DBType      string
	DatabaseURL string
	APIKeyFixed string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBType:      getEnv("DB_TYPE", "memory"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		APIKeyFixed: getEnv("API_KEY_FIXED", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
