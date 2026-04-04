package datasource

import (
	"sync"
	"time"
)

// DataSource is the interface for fetching Beads data.
// Both Client and Cache implement this interface.
type DataSource interface {
	ListIssues() ([]Issue, error)
	GetIssue(id string) (*IssueDetail, error)
	ListReady() ([]Issue, error)
}

// Cache wraps a Client and caches parsed results with a TTL.
type Cache struct {
	client  *Client
	ttl     time.Duration
	mu      sync.RWMutex
	entries map[string]cacheEntry
	now     func() time.Time
}

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

// NewCache creates a Cache with the given client and TTL.
func NewCache(client *Client, ttl time.Duration) *Cache {
	return &Cache{
		client:  client,
		ttl:     ttl,
		entries: make(map[string]cacheEntry),
		now:     time.Now,
	}
}

// ListIssues returns cached issues or delegates to the client.
func (c *Cache) ListIssues() ([]Issue, error) {
	if val, ok := c.get("list"); ok {
		return val.([]Issue), nil
	}
	issues, err := c.client.ListIssues()
	if err != nil {
		return nil, err
	}
	c.set("list", issues)
	return issues, nil
}

// GetIssue returns a cached issue detail or delegates to the client.
func (c *Cache) GetIssue(id string) (*IssueDetail, error) {
	key := "show:" + id
	if val, ok := c.get(key); ok {
		return val.(*IssueDetail), nil
	}
	detail, err := c.client.GetIssue(id)
	if err != nil {
		return nil, err
	}
	c.set(key, detail)
	return detail, nil
}

// ListReady returns cached ready issues or delegates to the client.
func (c *Cache) ListReady() ([]Issue, error) {
	if val, ok := c.get("ready"); ok {
		return val.([]Issue), nil
	}
	issues, err := c.client.ListReady()
	if err != nil {
		return nil, err
	}
	c.set("ready", issues)
	return issues, nil
}

// Invalidate clears all cached entries.
func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]cacheEntry)
}

func (c *Cache) get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if c.now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *Cache) set(key string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		data:      data,
		expiresAt: c.now().Add(c.ttl),
	}
}
