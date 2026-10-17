package analytics

import (
	"context"
	"log/slog"
)

// Sink delivers a single event to a durable backend (Kafka, a log, ...).
// Implementations should be safe for sequential use by a single publisher.
type Sink interface {
	Send(ctx context.Context, event Event) error
	Close() error
}

// NoopSink drops every event.
type NoopSink struct{}

// Send implements Sink.
func (NoopSink) Send(context.Context, Event) error { return nil }

// Close implements Sink.
func (NoopSink) Close() error { return nil }

// LogSink writes events to the structured logger. It is the default local sink
// until a real broker (Kafka) is configured.
type LogSink struct {
	logger *slog.Logger
}

// NewLogSink returns a Sink that logs each event.
func NewLogSink(logger *slog.Logger) *LogSink {
	return &LogSink{logger: logger}
}

// Send implements Sink.
func (s *LogSink) Send(_ context.Context, e Event) error {
	s.logger.Info("click event",
		"short_code", e.ShortCode,
		"timestamp", e.Timestamp,
		"ip", e.IP,
		"user_agent", e.UserAgent,
		"referrer", e.Referrer,
		"request_id", e.RequestID,
	)
	return nil
}

// Close implements Sink.
func (s *LogSink) Close() error { return nil }

var (
	_ Sink = NoopSink{}
	_ Sink = (*LogSink)(nil)
)
