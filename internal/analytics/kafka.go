package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// DefaultKafkaTopic is the click-event topic from the architecture docs.
const DefaultKafkaTopic = "url_clicks"

// messageProducer is the narrow slice of a Kafka client that KafkaSink needs,
// which keeps the sink unit-testable without a broker.
type messageProducer interface {
	Produce(ctx context.Context, key, value []byte) error
	Ping(ctx context.Context) error
	Close() error
}

// KafkaSink publishes click events as JSON records to a Kafka topic. Records
// are keyed by short code so all clicks for a link land on the same partition.
type KafkaSink struct {
	producer messageProducer
}

// NewKafkaSink connects to the given brokers and returns a sink publishing to
// topic (DefaultKafkaTopic when empty). Idempotent producing is enabled by
// the client so retries do not duplicate records within a session.
func NewKafkaSink(brokers []string, topic string) (*KafkaSink, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka: no brokers configured")
	}
	if topic == "" {
		topic = DefaultKafkaTopic
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.RecordRetries(5),
		kgo.ProducerLinger(5*time.Millisecond),
		// Topics are normally created by the Compose init job, but allow the
		// first produce to create one for ad-hoc/external brokers.
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka client: %w", err)
	}

	return newKafkaSink(&kgoProducer{client: client}), nil
}

func newKafkaSink(producer messageProducer) *KafkaSink {
	return &KafkaSink{producer: producer}
}

// Send implements Sink.
func (s *KafkaSink) Send(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal click event: %w", err)
	}
	return s.producer.Produce(ctx, []byte(event.ShortCode), payload)
}

// Ping checks broker connectivity.
func (s *KafkaSink) Ping(ctx context.Context) error {
	return s.producer.Ping(ctx)
}

// Close implements Sink.
func (s *KafkaSink) Close() error {
	return s.producer.Close()
}

// kgoProducer adapts a franz-go client to messageProducer.
type kgoProducer struct {
	client *kgo.Client
}

func (p *kgoProducer) Produce(ctx context.Context, key, value []byte) error {
	results := p.client.ProduceSync(ctx, &kgo.Record{Key: key, Value: value})
	return results.FirstErr()
}

func (p *kgoProducer) Ping(ctx context.Context) error {
	return p.client.Ping(ctx)
}

func (p *kgoProducer) Close() error {
	p.client.Close()
	return nil
}

var _ Sink = (*KafkaSink)(nil)
