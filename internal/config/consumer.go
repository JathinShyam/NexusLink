package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// ConsumerConfig configures the Kafka -> ClickHouse analytics consumer.
type ConsumerConfig struct {
	Env        string
	LogLevel   string
	LogDir     string
	LogFile    string
	HealthPort int

	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	ClickHouseURL      string
	ClickHouseDatabase string
	ClickHouseUser     string
	ClickHousePassword string

	BatchSize     int
	FlushInterval time.Duration
}

// LoadConsumer reads only the settings the consumer needs, so it can run
// without the API's database/JWT configuration.
func LoadConsumer() (*ConsumerConfig, error) {
	_ = godotenv.Load()

	brokers := parseCSV(getEnv("KAFKA_BROKERS", ""))
	if len(brokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}

	healthPort, err := strconv.Atoi(getEnv("CONSUMER_HEALTH_PORT", "8081"))
	if err != nil {
		return nil, fmt.Errorf("invalid CONSUMER_HEALTH_PORT: %w", err)
	}

	batchSize, err := strconv.Atoi(getEnv("CONSUMER_BATCH_SIZE", "500"))
	if err != nil {
		return nil, fmt.Errorf("invalid CONSUMER_BATCH_SIZE: %w", err)
	}
	if batchSize <= 0 {
		return nil, fmt.Errorf("invalid CONSUMER_BATCH_SIZE: must be positive")
	}

	flushMS, err := strconv.Atoi(getEnv("CONSUMER_FLUSH_INTERVAL_MS", "2000"))
	if err != nil {
		return nil, fmt.Errorf("invalid CONSUMER_FLUSH_INTERVAL_MS: %w", err)
	}
	if flushMS <= 0 {
		return nil, fmt.Errorf("invalid CONSUMER_FLUSH_INTERVAL_MS: must be positive")
	}

	return &ConsumerConfig{
		Env:        getEnv("ENV", "development"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		LogDir:     getEnv("LOG_DIR", "logs/consumer"),
		LogFile:    getEnv("LOG_FILE", "consumer.log"),
		HealthPort: healthPort,

		KafkaBrokers: brokers,
		KafkaTopic:   getEnv("KAFKA_TOPIC", "url_clicks"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "nexuslink-analytics"),

		ClickHouseURL:      getEnv("CLICKHOUSE_URL", "http://localhost:8123"),
		ClickHouseDatabase: getEnv("CLICKHOUSE_DATABASE", "nexuslink"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),

		BatchSize:     batchSize,
		FlushInterval: time.Duration(flushMS) * time.Millisecond,
	}, nil
}
