package config

import "os"

type Config struct {
	Port        string
	DBType      string
	DatabaseURL string
	DevMode     bool
	APIKeyFixed string // For backward compatibility in dev mode
	WebURL      string // Frontend URL for redirects (defaults to self)
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	return &Config{
		Port:        port,
		DBType:      getEnv("DB_TYPE", "memory"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DevMode:     getEnv("DEV_MODE", "") == "true",
		APIKeyFixed: getEnv("API_KEY_FIXED", ""),
		WebURL:      getEnv("WEB_URL", ""), // Empty means same origin
	}
}

func (c *Config) IsDevMode() bool {
	return c.DevMode || c.APIKeyFixed != ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
