package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                    string
	Port                   int
	LogLevel               string
	LogDir                 string
	LogFile                string
	DatabaseURL            string
	RedisURL               string
	BaseURL                string
	JWTSecret              string
	JWTTTL                 time.Duration
	MachineID              int64
	ShortenRateLimit       int64
	ShortenRateWindow      time.Duration
	RedirectBucketCapacity int64
	RedirectRefillPerSec   float64
	CORSOrigins            []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	machineID, err := strconv.ParseInt(getEnv("SNOWFLAKE_MACHINE_ID", "1"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid SNOWFLAKE_MACHINE_ID: %w", err)
	}

	shortenLimit, err := strconv.ParseInt(getEnv("RATE_LIMIT_SHORTEN", "30"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_SHORTEN: %w", err)
	}

	shortenWindowSec, err := strconv.Atoi(getEnv("RATE_LIMIT_SHORTEN_WINDOW_SEC", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_SHORTEN_WINDOW_SEC: %w", err)
	}

	bucketCapacity, err := strconv.ParseInt(getEnv("RATE_LIMIT_REDIRECT_CAPACITY", "100"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_REDIRECT_CAPACITY: %w", err)
	}

	refillRate, err := strconv.ParseFloat(getEnv("RATE_LIMIT_REDIRECT_REFILL_PER_SEC", "50"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_REDIRECT_REFILL_PER_SEC: %w", err)
	}

	jwtTTLHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL_HOURS: %w", err)
	}

	cfg := &Config{
		Env:                    getEnv("ENV", "development"),
		Port:                   port,
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		LogDir:                 getEnv("LOG_DIR", "logs/api"),
		LogFile:                getEnv("LOG_FILE", "app.log"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://nexuslink:nexuslink@localhost:5432/nexuslink?sslmode=disable"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379"),
		BaseURL:                getEnv("BASE_URL", "http://localhost:8080"),
		JWTSecret:              getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:                 time.Duration(jwtTTLHours) * time.Hour,
		MachineID:              machineID,
		ShortenRateLimit:       shortenLimit,
		ShortenRateWindow:      time.Duration(shortenWindowSec) * time.Second,
		RedirectBucketCapacity: bucketCapacity,
		RedirectRefillPerSec:   refillRate,
		CORSOrigins:            parseCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
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

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
