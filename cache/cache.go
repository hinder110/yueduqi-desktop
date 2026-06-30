package cache

import (
	"sync"
	"sync/atomic"
	"time"

	"yueduqi-desktop/model"
)

// Entry holds a cached value and its expiry.
type Entry[T any] struct {
	Value     T
	ExpiresAt time.Time
}

// Cache is a TTL-based in-memory cache built on sync.RWMutex + map.
// Safe for concurrent use. Disable at startup (before any requests) to
// bypass caching entirely — useful when debugging parse logic.
type Cache[T any] struct {
	mu       sync.RWMutex
	entries  map[string]Entry[T]
	ttl      time.Duration
	hits     atomic.Int64
	misses   atomic.Int64
	disabled atomic.Bool
}

// New creates a cache whose entries expire after ttl.
func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		entries: make(map[string]Entry[T]),
		ttl:    ttl,
	}
}

// Disable turns off caching. Existing entries are preserved but ignored.
func (c *Cache[T]) Disable() { c.disabled.Store(true) }

// Enable turns caching back on. Call once during setup if ever needed.
func (c *Cache[T]) Enable() { c.disabled.Store(false) }

// IsDisabled reports whether the cache is currently bypassed.
func (c *Cache[T]) IsDisabled() bool { return c.disabled.Load() }

// Get returns the cached value and true when a valid, non-expired entry is found.
func (c *Cache[T]) Get(key string) (T, bool) {
	if c.disabled.Load() {
		var zero T
		return zero, false
	}
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		c.misses.Add(1)
		var zero T
		return zero, false
	}
	c.hits.Add(1)
	return entry.Value, true
}

// Set stores a value keyed by key, expiring after the cache's TTL.
func (c *Cache[T]) Set(key string, value T) {
	if c.disabled.Load() {
		return
	}
	c.mu.Lock()
	c.entries[key] = Entry[T]{Value: value, ExpiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Stats returns the current hit and miss counters.
func (c *Cache[T]) Stats() (hits, misses int64) {
	return c.hits.Load(), c.misses.Load()
}

// --- global cache instances for parser operations ---

// HotBooks caches the discover/hot list (TTL: 5min). Single-key because
// GetHotBooks takes no parameters — front-page lists change infrequently.
var HotBooks = New[[]model.Book](5 * time.Minute)

// Search caches search results keyed by keyword (TTL: 3min). Users
// re-search the same term within a session; short TTL avoids staleness.
var Search = New[[]model.Book](3 * time.Minute)

// Chapters caches chapter lists keyed by "bookID|innerSource|innerTab"
// (TTL: 10min). Chapter catalogs rarely change, so longer expiry is safe.
var Chapters = New[[]model.Chapter](10 * time.Minute)

// AllStats reports the combined hit/miss counters across all caches.
func AllStats() (hits, misses int64) {
	h1, m1 := HotBooks.Stats()
	h2, m2 := Search.Stats()
	h3, m3 := Chapters.Stats()
	return h1 + h2 + h3, m1 + m2 + m3
}
