package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeProducer struct {
	mu     sync.Mutex
	key    []byte
	value  []byte
	err    error
	pinged bool
	closed bool
}

func (p *fakeProducer) Produce(_ context.Context, key, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.key = append([]byte(nil), key...)
	p.value = append([]byte(nil), value...)
	return nil
}

func (p *fakeProducer) Ping(context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pinged = true
	return p.err
}

func (p *fakeProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

func TestKafkaSinkSend(t *testing.T) {
	producer := &fakeProducer{}
	sink := newKafkaSink(producer)

	event := Event{
		ShortCode: "abc123",
		Timestamp: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC),
		IP:        "203.0.113.9",
		RequestID: "req-1",
	}

	if err := sink.Send(context.Background(), event); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if string(producer.key) != "abc123" {
		t.Errorf("record key = %q, want short code", producer.key)
	}

	var got Event
	if err := json.Unmarshal(producer.value, &got); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if got.ShortCode != event.ShortCode || got.IP != event.IP || got.RequestID != event.RequestID {
		t.Errorf("payload = %+v, want %+v", got, event)
	}
}

func TestKafkaSinkSendError(t *testing.T) {
	producer := &fakeProducer{err: errors.New("broker unavailable")}
	sink := newKafkaSink(producer)

	if err := sink.Send(context.Background(), Event{ShortCode: "abc"}); err == nil {
		t.Fatal("expected error from producer")
	}
}

func TestKafkaSinkPingAndClose(t *testing.T) {
	producer := &fakeProducer{}
	sink := newKafkaSink(producer)

	if err := sink.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if !producer.pinged {
		t.Error("Ping did not reach the producer")
	}

	if err := sink.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !producer.closed {
		t.Error("Close did not reach the producer")
	}
}

func TestNewKafkaSinkRequiresBrokers(t *testing.T) {
	if _, err := NewKafkaSink(nil, ""); err == nil {
		t.Fatal("expected error when no brokers are configured")
	}
}
