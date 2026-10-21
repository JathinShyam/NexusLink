package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("Addr() = %q, want :8080", cfg.Addr())
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/nexuslink?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want production", cfg.Env)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-port")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}

func TestGetEnvFallback(t *testing.T) {
	t.Setenv("NEXUSLINK_TEST_VAR", "")
	if got := getEnv("NEXUSLINK_TEST_VAR", "fallback"); got != "fallback" {
		t.Errorf("getEnv() = %q, want fallback", got)
	}
}

func TestLoadAnalyticsDefaults(t *testing.T) {
	t.Setenv("ANALYTICS_ENABLED", "")
	t.Setenv("ANALYTICS_SINK", "")
	t.Setenv("ANALYTICS_BUFFER_SIZE", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("KAFKA_TOPIC", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AnalyticsEnabled {
		t.Error("AnalyticsEnabled = true, want false")
	}
	if cfg.AnalyticsSink != "log" {
		t.Errorf("AnalyticsSink = %q, want log", cfg.AnalyticsSink)
	}
	if cfg.AnalyticsBufferSize != 1024 {
		t.Errorf("AnalyticsBufferSize = %d, want 1024", cfg.AnalyticsBufferSize)
	}
	if cfg.KafkaTopic != "url_clicks" {
		t.Errorf("KafkaTopic = %q, want url_clicks", cfg.KafkaTopic)
	}
	if len(cfg.KafkaBrokers) != 0 {
		t.Errorf("KafkaBrokers = %v, want empty", cfg.KafkaBrokers)
	}
}

func TestLoadKafkaSink(t *testing.T) {
	t.Setenv("ANALYTICS_ENABLED", "true")
	t.Setenv("ANALYTICS_SINK", "kafka")
	t.Setenv("KAFKA_BROKERS", "localhost:9092, localhost:9093")
	t.Setenv("KAFKA_TOPIC", "clicks")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.AnalyticsEnabled {
		t.Error("AnalyticsEnabled = false, want true")
	}
	if cfg.AnalyticsSink != "kafka" {
		t.Errorf("AnalyticsSink = %q, want kafka", cfg.AnalyticsSink)
	}
	if len(cfg.KafkaBrokers) != 2 {
		t.Errorf("KafkaBrokers = %v, want 2 brokers", cfg.KafkaBrokers)
	}
	if cfg.KafkaTopic != "clicks" {
		t.Errorf("KafkaTopic = %q, want clicks", cfg.KafkaTopic)
	}
}

func TestLoadKafkaSinkRequiresBrokers(t *testing.T) {
	t.Setenv("ANALYTICS_ENABLED", "true")
	t.Setenv("ANALYTICS_SINK", "kafka")
	t.Setenv("KAFKA_BROKERS", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when ANALYTICS_SINK=kafka without KAFKA_BROKERS")
	}
}

func TestLoadInvalidAnalyticsSink(t *testing.T) {
	t.Setenv("ANALYTICS_SINK", "s3")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid ANALYTICS_SINK")
	}
}
