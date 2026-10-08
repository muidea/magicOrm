package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"strconv"
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

// GenerateCacheKey hashes the complete scalar value, its type, the ordered
// directive sequence and the scenario. Unsupported complex values bypass the
// cache; serializing an object could hide state from its custom validator.
func (c *ConstraintCache) GenerateCacheKey(value any, constraints models.Constraints, scenario errors.Scenario) string {
	valueKey := getTypeHash(value)
	if valueKey == "" {
		return ""
	}
	type directiveKey struct {
		Key  models.Key
		Args []string
	}
	directives := []directiveKey{}
	if constraints != nil {
		for _, d := range constraints.Directives() {
			directives = append(directives, directiveKey{Key: d.Key(), Args: d.Args()})
		}
	}
	// Directive order affects custom validators and StopOnFirstError. Preserve
	// it, while JSON boundaries prevent delimiter collisions in arbitrary args.
	encoded, err := json.Marshal(struct {
		Value      string
		Directives []directiveKey
		Scenario   errors.Scenario
	}{valueKey, directives, scenario})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// Get retrieves a cached validation result
func (c *ConstraintCache) Get(key string) (error, bool) {
	if key == "" {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

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
	if key == "" {
		return
	}
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

// getTypeHash returns a value-sensitive, type-sensitive fingerprint. The
// empty string means that the value is not safe to cache.
func getTypeHash(value any) string {
	if value == nil {
		return "nil"
	}
	rv := reflect.ValueOf(value)
	var encoded string
	switch rv.Kind() {
	case reflect.String:
		encoded = rv.String()
	case reflect.Bool:
		encoded = strconv.FormatBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		encoded = strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		encoded = strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return ""
		}
		encoded = strconv.FormatFloat(f, 'g', -1, rv.Type().Bits())
	case reflect.Slice:
		if rv.Type().Elem().Kind() != reflect.Uint8 {
			return ""
		}
		if rv.IsNil() {
			encoded = "nil"
		} else {
			encoded = "bytes:" + hex.EncodeToString(rv.Bytes())
		}
	default:
		return ""
	}
	// Named types include their package, so equal strings used by different
	// custom validators cannot share a successful validation result.
	sum := sha256.Sum256([]byte(rv.Type().PkgPath() + "|" + rv.Type().String() + "|" + encoded))
	return hex.EncodeToString(sum[:])
}
