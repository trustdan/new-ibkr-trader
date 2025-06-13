package cache

import (
	"time"
	"github.com/patrickmn/go-cache"
)

// New creates a new cache instance with the specified expiration times
func New(defaultExpiration, cleanupInterval time.Duration) *cache.Cache {
	return cache.New(defaultExpiration, cleanupInterval)
}