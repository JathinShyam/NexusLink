package analytics

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JathinShyam/NexusLink/internal/logging"
)

type statsRecorder struct {
	mu        sync.Mutex
	published int
	dropped   int
	failed    int
}

func (s *statsRecorder) ClickEventPublished() { s.inc(&s.published) }
func (s *statsRecorder) ClickEventDropped()   { s.inc(&s.dropped) }
func (s *statsRecorder) ClickEventError()     { s.inc(&s.failed) }

func (s *statsRecorder) inc(field *int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	*field++
}

func (s *statsRecorder) counts() (published, dropped, failed int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.published, s.dropped, s.failed
}

type recordingSink struct {
	mu     sync.Mutex
	events []Event
	err    error
	closed bool
}

func (s *recordingSink) Send(_ context.Context, e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.events = append(s.events, e)
	return nil
}

func (s *recordingSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *recordingSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.events)
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func TestAsyncPublisherDeliversEvents(t *testing.T) {
	sink := &recordingSink{}
	stats := &statsRecorder{}
	pub := NewAsyncPublisher(sink, 8, logging.New("error", "test"), stats)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pub.Run(ctx)

	pub.Publish(context.Background(), Event{ShortCode: "abc"})
	pub.Publish(context.Background(), Event{ShortCode: "def"})

	waitFor(t, func() bool { return sink.count() == 2 })
	waitFor(t, func() bool {
		published, _, _ := stats.counts()
		return published == 2
	})

	if _, dropped, failed := stats.counts(); dropped != 0 || failed != 0 {
		t.Errorf("dropped=%d failed=%d, want 0", dropped, failed)
	}
}

func TestAsyncPublisherDropsWhenBufferFull(t *testing.T) {
	// No Run goroutine: the buffer never drains, so extra events are dropped.
	sink := &recordingSink{}
	stats := &statsRecorder{}
	pub := NewAsyncPublisher(sink, 1, logging.New("error", "test"), stats)

	pub.Publish(context.Background(), Event{ShortCode: "kept"})
	pub.Publish(context.Background(), Event{ShortCode: "dropped-1"})
	pub.Publish(context.Background(), Event{ShortCode: "dropped-2"})

	_, dropped, _ := stats.counts()
	if dropped != 2 {
		t.Errorf("dropped = %d, want 2", dropped)
	}
}

func TestAsyncPublisherCountsSinkErrors(t *testing.T) {
	sink := &recordingSink{err: errors.New("kafka down")}
	stats := &statsRecorder{}
	pub := NewAsyncPublisher(sink, 4, logging.New("error", "test"), stats)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pub.Run(ctx)

	pub.Publish(context.Background(), Event{ShortCode: "abc"})

	waitFor(t, func() bool {
		_, _, failed := stats.counts()
		return failed == 1
	})
}

func TestAsyncPublisherCloseClosesSink(t *testing.T) {
	sink := &recordingSink{}
	pub := NewAsyncPublisher(sink, 1, logging.New("error", "test"), nil)

	if err := pub.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !sink.closed {
		t.Error("sink was not closed")
	}
}

func TestNoopPublisher(t *testing.T) {
	NoopPublisher{}.Publish(context.Background(), Event{ShortCode: "abc"})
}
