package monitoring

import (
	"sync"
	"time"
)

// MetricsCollector collects validation system metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// Performance metrics
	totalValidations      int64
	totalValidationTime   time.Duration
	averageValidationTime time.Duration

	// Cache metrics
	cacheHits   int64
	cacheMisses int64

	// Error metrics
	totalErrors  int64
	errorsByType map[string]int64

	// Layer performance metrics
	layerTimes  map[string]time.Duration
	layerCounts map[string]int64

	// Resource metrics
	peakConcurrentValidations    int
	currentConcurrentValidations int
	memoryUsage                  int64 // bytes

	// Timestamps
	startTime time.Time
	lastReset time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	now := time.Now()
	return &MetricsCollector{
		errorsByType: make(map[string]int64),
		layerTimes:   make(map[string]time.Duration),
		layerCounts:  make(map[string]int64),
		startTime:    now,
		lastReset:    now,
	}
}

// RecordValidation records a validation operation
func (m *MetricsCollector) RecordValidation(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalValidations++
	m.totalValidationTime += duration

	// Update average
	if m.totalValidations > 0 {
		m.averageValidationTime = m.totalValidationTime / time.Duration(m.totalValidations)
	}

	// Record error if any
	if err != nil {
		m.totalErrors++
		errorType := getErrorType(err)
		m.errorsByType[errorType]++
	}

	// Update concurrent count
	m.currentConcurrentValidations++
	if m.currentConcurrentValidations > m.peakConcurrentValidations {
		m.peakConcurrentValidations = m.currentConcurrentValidations
	}

	// Decrement after a short delay (simulated)
	go func() {
		time.Sleep(duration)
		m.mu.Lock()
		m.currentConcurrentValidations--
		m.mu.Unlock()
	}()
}

// RecordCacheHit records a cache hit
func (m *MetricsCollector) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHits++
}

// RecordCacheMiss records a cache miss
func (m *MetricsCollector) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMisses++
}

// RecordLayerTime records time spent in a specific validation layer
func (m *MetricsCollector) RecordLayerTime(layer string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.layerTimes[layer] += duration
	m.layerCounts[layer]++
}

// UpdateMemoryUsage updates memory usage metrics
func (m *MetricsCollector) UpdateMemoryUsage(usage int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memoryUsage = usage
}

// GetMetrics returns current metrics
func (m *MetricsCollector) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate rates
	uptime := time.Since(m.startTime)
	validationRate := float64(m.totalValidations) / uptime.Seconds()
	errorRate := float64(m.totalErrors) / float64(m.totalValidations)

	// Calculate cache hit rate
	var cacheHitRate float64
	if m.cacheHits+m.cacheMisses > 0 {
		cacheHitRate = float64(m.cacheHits) / float64(m.cacheHits+m.cacheMisses)
	}

	// Calculate layer averages
	layerAverages := make(map[string]time.Duration)
	for layer, totalTime := range m.layerTimes {
		count := m.layerCounts[layer]
		if count > 0 {
			layerAverages[layer] = totalTime / time.Duration(count)
		}
	}

	return Metrics{
		TotalValidations:      m.totalValidations,
		ValidationRate:        validationRate,
		AverageValidationTime: m.averageValidationTime,
		TotalValidationTime:   m.totalValidationTime,

		CacheHits:    m.cacheHits,
		CacheMisses:  m.cacheMisses,
		CacheHitRate: cacheHitRate,

		TotalErrors:  m.totalErrors,
		ErrorRate:    errorRate,
		ErrorsByType: copyMap(m.errorsByType),

		LayerTimes:    copyDurationMap(m.layerTimes),
		LayerAverages: layerAverages,
		LayerCounts:   copyMap(m.layerCounts),

		PeakConcurrentValidations:    m.peakConcurrentValidations,
		CurrentConcurrentValidations: m.currentConcurrentValidations,
		MemoryUsage:                  m.memoryUsage,

		Uptime:    uptime,
		StartTime: m.startTime,
		LastReset: m.lastReset,
	}
}

// Reset resets all metrics
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalValidations = 0
	m.totalValidationTime = 0
	m.averageValidationTime = 0

	m.cacheHits = 0
	m.cacheMisses = 0

	m.totalErrors = 0
	m.errorsByType = make(map[string]int64)

	m.layerTimes = make(map[string]time.Duration)
	m.layerCounts = make(map[string]int64)

	m.peakConcurrentValidations = 0
	m.currentConcurrentValidations = 0
	m.memoryUsage = 0

	m.lastReset = time.Now()
}

// Metrics represents validation system metrics
type Metrics struct {
	// Performance metrics
	TotalValidations      int64         `json:"total_validations"`
	ValidationRate        float64       `json:"validation_rate"` // validations per second
	AverageValidationTime time.Duration `json:"average_validation_time"`
	TotalValidationTime   time.Duration `json:"total_validation_time"`

	// Cache metrics
	CacheHits    int64   `json:"cache_hits"`
	CacheMisses  int64   `json:"cache_misses"`
	CacheHitRate float64 `json:"cache_hit_rate"`

	// Error metrics
	TotalErrors  int64            `json:"total_errors"`
	ErrorRate    float64          `json:"error_rate"`
	ErrorsByType map[string]int64 `json:"errors_by_type"`

	// Layer performance metrics
	LayerTimes    map[string]time.Duration `json:"layer_times"`
	LayerAverages map[string]time.Duration `json:"layer_averages"`
	LayerCounts   map[string]int64         `json:"layer_counts"`

	// Resource metrics
	PeakConcurrentValidations    int   `json:"peak_concurrent_validations"`
	CurrentConcurrentValidations int   `json:"current_concurrent_validations"`
	MemoryUsage                  int64 `json:"memory_usage"` // bytes

	// System metrics
	Uptime    time.Duration `json:"uptime"`
	StartTime time.Time     `json:"start_time"`
	LastReset time.Time     `json:"last_reset"`
}

// Helper functions

func getErrorType(err error) string {
	// Extract error type from error message or structure
	// This is a simplified implementation
	errStr := err.Error()

	// Common error patterns
	switch {
	case contains(errStr, "type"):
		return "type_validation"
	case contains(errStr, "constraint"):
		return "constraint_validation"
	case contains(errStr, "database"):
		return "database_validation"
	case contains(errStr, "scenario"):
		return "scenario_adaptation"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

func copyMap(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func copyDurationMap(src map[string]time.Duration) map[string]time.Duration {
	dst := make(map[string]time.Duration, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
