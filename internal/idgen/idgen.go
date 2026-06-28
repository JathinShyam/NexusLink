package idgen

import "context"

// Generator produces unique numeric IDs for short code generation.
// A Snowflake implementation can satisfy this interface in a later phase.
type Generator interface {
	Next(ctx context.Context) (uint64, error)
}
