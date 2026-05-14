package validator

import (
	"regexp"
	"sync"
)

// regexCache caches compiled regular expressions.
var regexCache = &RegexCache{}

// RegexCache provides thread-safe regex compilation caching.
type RegexCache struct {
	mu    sync.RWMutex
	cache map[string]*regexp.Regexp
}

// Compile returns a cached or newly compiled regex.
func (c *RegexCache) Compile(pattern string) (*regexp.Regexp, error) {
	c.mu.RLock()
	if c.cache != nil {
		if re, ok := c.cache[pattern]; ok {
			c.mu.RUnlock()
			return re, nil
		}
	}
	c.mu.RUnlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[string]*regexp.Regexp)
	}
	c.cache[pattern] = re
	c.mu.Unlock()

	return re, nil
}

// Clear resets the cache.
func (c *RegexCache) Clear() {
	c.mu.Lock()
	c.cache = nil
	c.mu.Unlock()
}
