package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// KafkaConsumer reads click events from Kafka and writes them to a Store in
// batches. Offsets are committed only after a successful insert, so delivery
// is at-least-once; the ClickHouse table deduplicates on request_id.
type KafkaConsumer struct {
	client        *kgo.Client
	store         Store
	batchSize     int
	flushInterval time.Duration
	logger        *slog.Logger
}

// NewKafkaConsumer joins group and consumes topic, starting from the earliest
// offset for a brand-new group.
func NewKafkaConsumer(
	brokers []string,
	topic, group string,
	store Store,
	batchSize int,
	flushInterval time.Duration,
	logger *slog.Logger,
) (*KafkaConsumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka consumer: no brokers configured")
	}
	if store == nil {
		return nil, fmt.Errorf("kafka consumer: store is required")
	}
	if topic == "" {
		topic = DefaultKafkaTopic
	}
	if group == "" {
		group = "nexuslink-analytics"
	}
	if batchSize <= 0 {
		batchSize = 500
	}
	if flushInterval <= 0 {
		flushInterval = 2 * time.Second
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		client:        client,
		store:         store,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		logger:        logger,
	}, nil
}

// Ping checks broker connectivity (used by the readiness endpoint).
func (c *KafkaConsumer) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}

// Close leaves the consumer group and releases the client.
func (c *KafkaConsumer) Close() error {
	c.client.Close()
	return nil
}

// Run consumes until ctx is cancelled, flushing buffered records on shutdown.
func (c *KafkaConsumer) Run(ctx context.Context) error {
	var (
		events  []Event
		records []*kgo.Record
	)
	lastFlush := time.Now()

	flush := func(flushCtx context.Context) error {
		if len(records) == 0 {
			return nil
		}
		if len(events) > 0 {
			if err := c.insertWithRetry(flushCtx, events); err != nil {
				return err
			}
		}
		if err := c.client.CommitRecords(flushCtx, records...); err != nil {
			return fmt.Errorf("commit offsets: %w", err)
		}
		c.logger.Info("flushed click batch", "events", len(events), "records", len(records))
		events = events[:0]
		records = records[:0]
		lastFlush = time.Now()
		return nil
	}

	for {
		fetchCtx, cancel := context.WithTimeout(ctx, c.flushInterval)
		fetches := c.client.PollFetches(fetchCtx)
		cancel()

		if ctx.Err() != nil {
			// Give buffered events a bounded chance to land on shutdown.
			flushCtx, flushCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			err := flush(flushCtx)
			flushCancel()
			if err != nil {
				c.logger.Error("final click batch flush failed", "error", err)
			}
			return nil
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			// PollFetches uses a timeout to flush partial batches; that
			// deadline is expected, not a broker problem.
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return
			}
			c.logger.Warn("kafka fetch error", "topic", topic, "partition", partition, "error", err)
		})

		fetches.EachRecord(func(record *kgo.Record) {
			records = append(records, record)
			if event, ok := decodeClickRecord(record, c.logger); ok {
				events = append(events, event)
			}
		})

		if len(records) >= c.batchSize || time.Since(lastFlush) >= c.flushInterval {
			if err := flush(ctx); err != nil {
				return err
			}
		}
	}
}

func decodeClickRecord(record *kgo.Record, logger *slog.Logger) (Event, bool) {
	var event Event
	if err := json.Unmarshal(record.Value, &event); err != nil {
		logger.Warn("skipping malformed click event",
			"error", err,
			"topic", record.Topic,
			"partition", record.Partition,
			"offset", record.Offset,
		)
		return Event{}, false
	}
	return event, true
}

// insertWithRetry blocks (with backoff) until ClickHouse accepts the batch or
// the consumer shuts down. Blocking here is deliberate backpressure: Kafka
// retains the uncommitted records while the warehouse is unavailable.
func (c *KafkaConsumer) insertWithRetry(ctx context.Context, events []Event) error {
	backoff := time.Second
	for attempt := 1; ; attempt++ {
		err := c.store.Insert(ctx, events)
		if err == nil {
			return nil
		}

		c.logger.Error("clickhouse insert failed; retrying",
			"error", err,
			"attempt", attempt,
			"events", len(events),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}
