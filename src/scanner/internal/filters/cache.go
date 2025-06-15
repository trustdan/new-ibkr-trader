package filters

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"github.com/ibkr-trader/scanner/internal/models"
)

// FilterCache implements a thread-safe cache for filter results
type FilterCache struct {
	contractCache map[string]*cacheEntry
	spreadCache   map[string]*cacheEntry
	mu            sync.RWMutex
	ttl           time.Duration
	
	// Stats
	hits      int64
	misses    int64
	evictions int64
}

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

// NewFilterCache creates a new filter cache with TTL
func NewFilterCache(ttl time.Duration) *FilterCache {
	cache := &FilterCache{
		contractCache: make(map[string]*cacheEntry),
		spreadCache:   make(map[string]*cacheEntry),
		ttl:           ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanupLoop()
	
	return cache
}

// GetContracts retrieves cached contract results
func (fc *FilterCache) GetContracts(input []models.OptionContract) ([]models.OptionContract, bool) {
	key := fc.generateKey(input)
	
	fc.mu.RLock()
	entry, exists := fc.contractCache[key]
	fc.mu.RUnlock()
	
	if !exists {
		fc.incrementMisses()
		return nil, false
	}
	
	if time.Since(entry.timestamp) > fc.ttl {
		fc.mu.Lock()
		delete(fc.contractCache, key)
		fc.evictions++
		fc.mu.Unlock()
		fc.incrementMisses()
		return nil, false
	}
	
	fc.incrementHits()
	return entry.data.([]models.OptionContract), true
}

// SetContracts caches contract filter results
func (fc *FilterCache) SetContracts(input, result []models.OptionContract) {
	key := fc.generateKey(input)
	
	fc.mu.Lock()
	fc.contractCache[key] = &cacheEntry{
		data:      result,
		timestamp: time.Now(),
	}
	fc.mu.Unlock()
}

// GetSpreads retrieves cached spread results
func (fc *FilterCache) GetSpreads(input []models.VerticalSpread) ([]models.VerticalSpread, bool) {
	key := fc.generateKey(input)
	
	fc.mu.RLock()
	entry, exists := fc.spreadCache[key]
	fc.mu.RUnlock()
	
	if !exists {
		fc.incrementMisses()
		return nil, false
	}
	
	if time.Since(entry.timestamp) > fc.ttl {
		fc.mu.Lock()
		delete(fc.spreadCache, key)
		fc.evictions++
		fc.mu.Unlock()
		fc.incrementMisses()
		return nil, false
	}
	
	fc.incrementHits()
	return entry.data.([]models.VerticalSpread), true
}

// SetSpreads caches spread filter results
func (fc *FilterCache) SetSpreads(input, result []models.VerticalSpread) {
	key := fc.generateKey(input)
	
	fc.mu.Lock()
	fc.spreadCache[key] = &cacheEntry{
		data:      result,
		timestamp: time.Now(),
	}
	fc.mu.Unlock()
}

// generateKey creates a cache key from input data
func (fc *FilterCache) generateKey(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%x", hash)
}

// cleanupLoop periodically removes expired entries
func (fc *FilterCache) cleanupLoop() {
	ticker := time.NewTicker(fc.ttl / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		fc.cleanup()
	}
}

// cleanup removes expired entries
func (fc *FilterCache) cleanup() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	now := time.Now()
	
	// Clean contract cache
	for key, entry := range fc.contractCache {
		if now.Sub(entry.timestamp) > fc.ttl {
			delete(fc.contractCache, key)
			fc.evictions++
		}
	}
	
	// Clean spread cache
	for key, entry := range fc.spreadCache {
		if now.Sub(entry.timestamp) > fc.ttl {
			delete(fc.spreadCache, key)
			fc.evictions++
		}
	}
}

// GetStats returns cache statistics
func (fc *FilterCache) GetStats() (hits, misses, evictions int64, hitRate float64) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	
	hits = fc.hits
	misses = fc.misses
	evictions = fc.evictions
	
	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}
	
	return
}

// incrementHits safely increments hit counter
func (fc *FilterCache) incrementHits() {
	fc.mu.Lock()
	fc.hits++
	fc.mu.Unlock()
}

// incrementMisses safely increments miss counter
func (fc *FilterCache) incrementMisses() {
	fc.mu.Lock()
	fc.misses++
	fc.mu.Unlock()
}

// Clear removes all cached entries
func (fc *FilterCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	fc.contractCache = make(map[string]*cacheEntry)
	fc.spreadCache = make(map[string]*cacheEntry)
	fc.hits = 0
	fc.misses = 0
	fc.evictions = 0
}