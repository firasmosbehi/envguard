package schema

import (
	"os"
	"sync"
	"time"
)

// schemaCacheEntry holds a cached schema with its mtime.
type schemaCacheEntry struct {
	schema *Schema
	mtime  time.Time
}

// SchemaCache provides mtime-based schema parsing cache.
type SchemaCache struct {
	mu    sync.RWMutex
	cache map[string]schemaCacheEntry
}

// DefaultSchemaCache is the global schema cache.
var DefaultSchemaCache = &SchemaCache{}

// Get returns a cached schema if it exists and hasn't been modified.
func (c *SchemaCache) Get(path string) (*Schema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return nil, false
	}

	entry, ok := c.cache[path]
	if !ok {
		return nil, false
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}

	if info.ModTime().After(entry.mtime) {
		return nil, false
	}

	return entry.schema, true
}

// Set caches a schema with its current mtime.
func (c *SchemaCache) Set(path string, s *Schema) {
	info, err := os.Stat(path)
	var mtime time.Time
	if err == nil {
		mtime = info.ModTime()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string]schemaCacheEntry)
	}
	c.cache[path] = schemaCacheEntry{schema: s, mtime: mtime}
}

// Clear resets the cache.
func (c *SchemaCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = nil
}
