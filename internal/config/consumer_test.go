package config

import (
	"testing"
	"time"
)

func TestLoadConsumerDefaults(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KAFKA_TOPIC", "")
	t.Setenv("KAFKA_GROUP_ID", "")
	t.Setenv("CLICKHOUSE_URL", "")
	t.Setenv("CLICKHOUSE_DATABASE", "")
	t.Setenv("CLICKHOUSE_USER", "")
	t.Setenv("CLICKHOUSE_PASSWORD", "")
	t.Setenv("CONSUMER_BATCH_SIZE", "")
	t.Setenv("CONSUMER_FLUSH_INTERVAL_MS", "")
	t.Setenv("CONSUMER_HEALTH_PORT", "")

	cfg, err := LoadConsumer()
	if err != nil {
		t.Fatalf("LoadConsumer() error = %v", err)
	}

	if len(cfg.KafkaBrokers) != 1 {
		t.Errorf("KafkaBrokers = %v, want one broker", cfg.KafkaBrokers)
	}
	if cfg.KafkaTopic != "url_clicks" {
		t.Errorf("KafkaTopic = %q, want url_clicks", cfg.KafkaTopic)
	}
	if cfg.KafkaGroupID != "nexuslink-analytics" {
		t.Errorf("KafkaGroupID = %q, want nexuslink-analytics", cfg.KafkaGroupID)
	}
	if cfg.ClickHouseURL != "http://localhost:8123" {
		t.Errorf("ClickHouseURL = %q", cfg.ClickHouseURL)
	}
	if cfg.ClickHouseDatabase != "nexuslink" {
		t.Errorf("ClickHouseDatabase = %q", cfg.ClickHouseDatabase)
	}
	if cfg.BatchSize != 500 {
		t.Errorf("BatchSize = %d, want 500", cfg.BatchSize)
	}
	if cfg.FlushInterval != 2*time.Second {
		t.Errorf("FlushInterval = %v, want 2s", cfg.FlushInterval)
	}
	if cfg.HealthPort != 8081 {
		t.Errorf("HealthPort = %d, want 8081", cfg.HealthPort)
	}
}

func TestLoadConsumerRequiresBrokers(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "")

	if _, err := LoadConsumer(); err == nil {
		t.Fatal("expected error when KAFKA_BROKERS is empty")
	}
}

func TestLoadConsumerInvalidBatchSize(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("CONSUMER_BATCH_SIZE", "0")

	if _, err := LoadConsumer(); err == nil {
		t.Fatal("expected error for non-positive CONSUMER_BATCH_SIZE")
	}
}

func TestLoadConsumerInvalidFlushInterval(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("CONSUMER_FLUSH_INTERVAL_MS", "-1")

	if _, err := LoadConsumer(); err == nil {
		t.Fatal("expected error for non-positive CONSUMER_FLUSH_INTERVAL_MS")
	}
}
