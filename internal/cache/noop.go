package cache

import "context"

// NoopCache is a cache that always misses. Used when Redis is unavailable.
type NoopCache struct{}

func (NoopCache) Get(context.Context, string) (Entry, bool) { return Entry{}, false }
func (NoopCache) Set(context.Context, string, Entry) error  { return nil }
func (NoopCache) Delete(context.Context, string) error     { return nil }
