package analytics

import (
	"context"
	"log/slog"
	"time"
)

const (
	defaultBufferSize = 1024
	sendTimeout       = 5 * time.Second
)

// Stats is the minimal metrics surface the publisher reports to. It is
// satisfied by *metrics.Metrics without that package depending on analytics.
type Stats interface {
	ClickEventPublished()
	ClickEventDropped()
	ClickEventError()
}

// AsyncPublisher decouples event production from delivery. Publish enqueues
// onto a bounded channel and never blocks: when the buffer is full, events are
// dropped and counted (fail-open) so a slow or down sink can never affect
// redirect latency. Start exactly one Run goroutine to drain the buffer.
type AsyncPublisher struct {
	sink   Sink
	logger *slog.Logger
	stats  Stats
	events chan Event
}

// NewAsyncPublisher returns a publisher that buffers up to bufferSize events
// (defaulting when <= 0) before dropping.
func NewAsyncPublisher(sink Sink, bufferSize int, logger *slog.Logger, stats Stats) *AsyncPublisher {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	return &AsyncPublisher{
		sink:   sink,
		logger: logger,
		stats:  stats,
		events: make(chan Event, bufferSize),
	}
}

// Publish enqueues the event, dropping it if the buffer is full.
func (p *AsyncPublisher) Publish(_ context.Context, event Event) {
	select {
	case p.events <- event:
	default:
		if p.stats != nil {
			p.stats.ClickEventDropped()
		}
		p.logger.Warn("analytics buffer full, dropping click event",
			"short_code", event.ShortCode)
	}
}

// Run drains buffered events until ctx is cancelled. Call it in its own
// goroutine. Buffered events are abandoned on cancellation so shutdown never
// blocks on a slow sink.
func (p *AsyncPublisher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-p.events:
			p.send(ctx, event)
		}
	}
}

func (p *AsyncPublisher) send(ctx context.Context, event Event) {
	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	if err := p.sink.Send(sendCtx, event); err != nil {
		if p.stats != nil {
			p.stats.ClickEventError()
		}
		p.logger.Error("failed to publish click event",
			"error", err,
			"short_code", event.ShortCode,
		)
		return
	}
	if p.stats != nil {
		p.stats.ClickEventPublished()
	}
}

// Close closes the underlying sink.
func (p *AsyncPublisher) Close() error {
	return p.sink.Close()
}

var _ Publisher = (*AsyncPublisher)(nil)
