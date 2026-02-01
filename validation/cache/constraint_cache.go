package cache

import (
	"sync"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
)

// ConstraintCacheEntry represents a cached constraint validation result
type ConstraintCacheEntry struct {
	Value       any
	Constraints models.Constraints
	Scenario    errors.Scenario
	Result      error
	Timestamp   time.Time
	AccessCount int
}

// ConstraintCache implements caching for constraint validation results
type ConstraintCache struct {
	mu         sync.RWMutex
	cache      map[string]*ConstraintCacheEntry
	maxSize    int
	defaultTTL time.Duration
	stats      CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Size        int
	MaxSize     int
	MemoryUsage int64 // in bytes (approximate)
}

// NewConstraintCache creates a new constraint cache
func NewConstraintCache(maxSize int, defaultTTL time.Duration) *ConstraintCache {
	return &ConstraintCache{
		cache:      make(map[string]*ConstraintCacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
		stats: CacheStats{
			MaxSize: maxSize,
		},
	}
}

// GenerateCacheKey generates a unique cache key for constraint validation
func (c *ConstraintCache) GenerateCacheKey(value any, constraints models.Constraints, scenario errors.Scenario) string {
	// Create a simple hash-based key
	// In a real implementation, this would be more sophisticated
	key := ""

	// Add value type and hash
	if value != nil {
		// Simple type-based key
		key += getTypeHash(value)
	}

	// Add constraints
	if constraints != nil {
		directives := constraints.Directives()
		for _, d := range directives {
			key += "|" + string(d.Key())
			if d.HasArgs() {
				for _, arg := range d.Args() {
					key += ":" + arg
				}
			}
		}
	}

	// Add scenario
	key += "|" + string(scenario)

	return key
}

// Get retrieves a cached validation result
func (c *ConstraintCache) Get(key string) (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.defaultTTL {
		c.stats.Misses++
		return nil, false
	}

	// Update access count and timestamp
	entry.AccessCount++
	entry.Timestamp = time.Now()

	c.stats.Hits++
	return entry.Result, true
}

// Set stores a validation result in the cache
func (c *ConstraintCache) Set(key string, value any, constraints models.Constraints, scenario errors.Scenario, result error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.cache) >= c.maxSize {
		c.evictLRU()
	}

	entry := &ConstraintCacheEntry{
		Value:       value,
		Constraints: constraints,
		Scenario:    scenario,
		Result:      result,
		Timestamp:   time.Now(),
		AccessCount: 1,
	}

	c.cache[key] = entry
	c.stats.Size = len(c.cache)

	// Update approximate memory usage
	c.updateMemoryUsage()
}

// Clear removes all entries from the cache
func (c *ConstraintCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*ConstraintCacheEntry)
	c.stats.Size = 0
	c.stats.MemoryUsage = 0
}

// ClearExpired removes expired entries from the cache
func (c *ConstraintCache) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiredKeys := make([]string, 0)
	now := time.Now()

	for key, entry := range c.cache {
		if now.Sub(entry.Timestamp) > c.defaultTTL {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.cache, key)
	}

	c.stats.Size = len(c.cache)
	c.updateMemoryUsage()
}

// GetStats returns cache statistics
func (c *ConstraintCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.cache)
	return stats
}

// evictLRU evicts the least recently used entries
func (c *ConstraintCache) evictLRU() {
	if len(c.cache) == 0 {
		return
	}

	// Find entry with lowest access count and oldest timestamp
	var lruKey string
	var lruEntry *ConstraintCacheEntry

	for key, entry := range c.cache {
		if lruEntry == nil {
			lruKey = key
			lruEntry = entry
			continue
		}

		// Compare access count first, then timestamp
		if entry.AccessCount < lruEntry.AccessCount ||
			(entry.AccessCount == lruEntry.AccessCount && entry.Timestamp.Before(lruEntry.Timestamp)) {
			lruKey = key
			lruEntry = entry
		}
	}

	// Remove the LRU entry
	delete(c.cache, lruKey)
	c.stats.Evictions++
}

// updateMemoryUsage updates the approximate memory usage
func (c *ConstraintCache) updateMemoryUsage() {
	// Simple approximation: each entry ~1KB
	c.stats.MemoryUsage = int64(len(c.cache) * 1024)
}

// getTypeHash generates a simple hash for a value's type
func getTypeHash(value any) string {
	// This is a simplified implementation
	// In a real system, you'd want a proper hash
	switch v := value.(type) {
	case string:
		return "string:" + v
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	case []byte:
		return "bytes"
	default:
		return "complex"
	}
}
