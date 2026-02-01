package cache

import (
	"sync"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
)

// ValidationCache provides caching for validation results
type ValidationCache struct {
	constraintCache *ConstraintCache
	modelCache      *ModelCache
	mu              sync.RWMutex
	enabled         bool
	config          CacheConfig
}

// ModelCacheEntry represents a cached model validation result
type ModelCacheEntry struct {
	Model     models.Model
	Scenario  errors.Scenario
	Result    error
	Timestamp time.Time
}

// ModelCache implements caching for model validation results
type ModelCache struct {
	mu         sync.RWMutex
	cache      map[string]*ModelCacheEntry
	maxSize    int
	defaultTTL time.Duration
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled              bool
	MaxConstraintEntries int
	MaxModelEntries      int
	DefaultTTL           time.Duration
	CleanupInterval      time.Duration
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:              true,
		MaxConstraintEntries: 1000,
		MaxModelEntries:      500,
		DefaultTTL:           5 * time.Minute,
		CleanupInterval:      1 * time.Minute,
	}
}

// NewValidationCache creates a new validation cache
func NewValidationCache(config CacheConfig) *ValidationCache {
	cache := &ValidationCache{
		constraintCache: NewConstraintCache(config.MaxConstraintEntries, config.DefaultTTL),
		modelCache: &ModelCache{
			cache:      make(map[string]*ModelCacheEntry),
			maxSize:    config.MaxModelEntries,
			defaultTTL: config.DefaultTTL,
		},
		enabled: config.Enabled,
		config:  config,
	}

	// Start cleanup goroutine if enabled
	if config.Enabled && config.CleanupInterval > 0 {
		go cache.startCleanup()
	}

	return cache
}

// GetConstraintResult retrieves a cached constraint validation result
func (c *ValidationCache) GetConstraintResult(value any, constraints models.Constraints, scenario errors.Scenario) (error, bool) {
	if !c.enabled {
		return nil, false
	}

	key := c.constraintCache.GenerateCacheKey(value, constraints, scenario)
	return c.constraintCache.Get(key)
}

// SetConstraintResult stores a constraint validation result in the cache
func (c *ValidationCache) SetConstraintResult(value any, constraints models.Constraints, scenario errors.Scenario, result error) {
	if !c.enabled {
		return
	}

	key := c.constraintCache.GenerateCacheKey(value, constraints, scenario)
	c.constraintCache.Set(key, value, constraints, scenario, result)
}

// GetModelResult retrieves a cached model validation result
func (c *ValidationCache) GetModelResult(model models.Model, scenario errors.Scenario) (error, bool) {
	if !c.enabled {
		return nil, false
	}

	c.modelCache.mu.RLock()
	defer c.modelCache.mu.RUnlock()

	key := generateModelCacheKey(model, scenario)
	entry, exists := c.modelCache.cache[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.modelCache.defaultTTL {
		return nil, false
	}

	return entry.Result, true
}

// SetModelResult stores a model validation result in the cache
func (c *ValidationCache) SetModelResult(model models.Model, scenario errors.Scenario, result error) {
	if !c.enabled {
		return
	}

	c.modelCache.mu.Lock()
	defer c.modelCache.mu.Unlock()

	// Check if we need to evict entries
	if len(c.modelCache.cache) >= c.modelCache.maxSize {
		c.modelCache.evictOldest()
	}

	key := generateModelCacheKey(model, scenario)
	entry := &ModelCacheEntry{
		Model:     model,
		Scenario:  scenario,
		Result:    result,
		Timestamp: time.Now(),
	}

	c.modelCache.cache[key] = entry
}

// Clear clears all caches
func (c *ValidationCache) Clear() {
	c.constraintCache.Clear()

	c.modelCache.mu.Lock()
	defer c.modelCache.mu.Unlock()
	c.modelCache.cache = make(map[string]*ModelCacheEntry)
}

// ClearExpired clears expired entries from all caches
func (c *ValidationCache) ClearExpired() {
	c.constraintCache.ClearExpired()

	c.modelCache.mu.Lock()
	defer c.modelCache.mu.Unlock()

	expiredKeys := make([]string, 0)
	now := time.Now()

	for key, entry := range c.modelCache.cache {
		if now.Sub(entry.Timestamp) > c.modelCache.defaultTTL {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.modelCache.cache, key)
	}
}

// GetStats returns cache statistics
func (c *ValidationCache) GetStats() map[string]interface{} {
	constraintStats := c.constraintCache.GetStats()

	c.modelCache.mu.RLock()
	modelSize := len(c.modelCache.cache)
	c.modelCache.mu.RUnlock()

	return map[string]interface{}{
		"enabled": c.enabled,
		"constraint_cache": map[string]interface{}{
			"hits":         constraintStats.Hits,
			"misses":       constraintStats.Misses,
			"evictions":    constraintStats.Evictions,
			"size":         constraintStats.Size,
			"max_size":     constraintStats.MaxSize,
			"memory_usage": constraintStats.MemoryUsage,
		},
		"model_cache": map[string]interface{}{
			"size":     modelSize,
			"max_size": c.modelCache.maxSize,
		},
		"config": map[string]interface{}{
			"default_ttl":      c.config.DefaultTTL.String(),
			"cleanup_interval": c.config.CleanupInterval.String(),
		},
	}
}

// Enable enables the cache
func (c *ValidationCache) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = true
}

// Disable disables the cache
func (c *ValidationCache) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = false
}

// IsEnabled returns whether the cache is enabled
func (c *ValidationCache) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// startCleanup starts the cleanup goroutine
func (c *ValidationCache) startCleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.ClearExpired()
	}
}

// generateModelCacheKey generates a cache key for a model
func generateModelCacheKey(model models.Model, scenario errors.Scenario) string {
	// Create a key based on model name and scenario
	// In a real implementation, this would include field information
	return model.GetName() + "|" + string(scenario)
}

// evictOldest evicts the oldest entries from the model cache
func (mc *ModelCache) evictOldest() {
	if len(mc.cache) == 0 {
		return
	}

	// Find oldest entry
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range mc.cache {
		if first {
			oldestKey = key
			oldestTime = entry.Timestamp
			first = false
			continue
		}

		if entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	// Remove the oldest entry
	delete(mc.cache, oldestKey)
}
