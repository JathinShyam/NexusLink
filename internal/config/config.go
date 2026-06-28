package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        int
	LogLevel    string
	LogDir      string
	LogFile     string
	DatabaseURL string
	BaseURL     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	cfg := &Config{
		Env:         getEnv("ENV", "development"),
		Port:        port,
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogDir:      getEnv("LOG_DIR", "logs/api"),
		LogFile:     getEnv("LOG_FILE", "app.log"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://nexuslink:nexuslink@localhost:5432/nexuslink?sslmode=disable"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
