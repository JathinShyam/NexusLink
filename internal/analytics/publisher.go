package analytics

import "context"

// Publisher accepts click events for asynchronous delivery. Implementations
// must never block the request that produced the event.
type Publisher interface {
	Publish(ctx context.Context, event Event)
}

// NoopPublisher discards every event. It is used when analytics is disabled.
type NoopPublisher struct{}

// Publish implements Publisher.
func (NoopPublisher) Publish(context.Context, Event) {}

var _ Publisher = NoopPublisher{}
