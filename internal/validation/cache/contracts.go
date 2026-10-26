package cache

import (
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

// CacheStats holds cache statistics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Size        int
	MaxSize     int
	MemoryUsage int64 // in bytes (approximate)
}

// ModelCacheEntry represents a cached model validation result
type ModelCacheEntry struct {
	Model     models.Model
	Scenario  errors.Scenario
	Result    error
	Timestamp time.Time
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled              bool
	MaxConstraintEntries int
	MaxModelEntries      int
	DefaultTTL           time.Duration
	CleanupInterval      time.Duration
}
