package validation

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewValidationMetricsCollector(t *testing.T) {
	collector := NewValidationMetricsCollector()
	assert.NotNil(t, collector)
}

func TestRecordValidation(t *testing.T) {
	collector := NewValidationMetricsCollector()

	// Record successful validation
	collector.RecordValidation("validate", "User", "insert", 50*time.Millisecond, nil)
	collector.RecordValidation("validate", "User", "update", 30*time.Millisecond, nil)

	// Record failed validation
	collector.RecordValidation("validate", "Product", "insert", 20*time.Millisecond, assert.AnError)

	// Check validation counters
	counters := collector.GetValidationCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("validate", "User", "insert", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("validate", "User", "update", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("validate", "Product", "insert", "error")])

	// Check error counters
	errorCounters := collector.GetErrorCounters()
	assert.Equal(t, int64(1), errorCounters[metrics.BuildKey("validate", "Product", "insert", "unknown")])

	// Check durations
	durations := collector.GetValidationDurations()
	assert.Len(t, durations[metrics.BuildKey("validate", "User", "insert", "success")], 1)
	assert.Len(t, durations[metrics.BuildKey("validate", "User", "update", "success")], 1)
	assert.Len(t, durations[metrics.BuildKey("validate", "Product", "insert", "error")], 1)
}

func TestRecordCacheAccess(t *testing.T) {
	collector := NewValidationMetricsCollector()

	// Record cache hits and misses
	collector.RecordCacheAccess("type", true)
	collector.RecordCacheAccess("type", true)
	collector.RecordCacheAccess("type", false)
	collector.RecordCacheAccess("constraint", true)
	collector.RecordCacheAccess("constraint", false)
	collector.RecordCacheAccess("constraint", false)

	// Check cache access counters
	counters := collector.GetCacheAccessCounters()
	assert.Equal(t, int64(2), counters[metrics.BuildKey("type", "hit")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("type", "miss")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("constraint", "hit")])
	assert.Equal(t, int64(2), counters[metrics.BuildKey("constraint", "miss")])

	// Check cache hit ratio
	hitRatio := collector.GetCacheHitRatio("type")
	expectedRatio := 2.0 / 3.0 // 2 hits / (2 hits + 1 miss)
	assert.InDelta(t, expectedRatio, hitRatio, 0.001)

	hitRatio = collector.GetCacheHitRatio("constraint")
	expectedRatio = 1.0 / 3.0 // 1 hit / (1 hit + 2 misses)
	assert.InDelta(t, expectedRatio, hitRatio, 0.001)

	// Test cache type with no accesses
	hitRatio = collector.GetCacheHitRatio("nonexistent")
	assert.Equal(t, 0.0, hitRatio)
}

func TestRecordConstraintCheck(t *testing.T) {
	collector := NewValidationMetricsCollector()

	// Record constraint checks
	collector.RecordConstraintCheck("required", "Name", true)
	collector.RecordConstraintCheck("required", "Email", true)
	collector.RecordConstraintCheck("required", "Password", false)
	collector.RecordConstraintCheck("range", "Age", true)
	collector.RecordConstraintCheck("format", "Email", false)

	// Check constraint check counters
	counters := collector.GetConstraintCheckCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("required", "Name", "passed")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("required", "Email", "passed")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("required", "Password", "failed")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("range", "Age", "passed")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("format", "Email", "failed")])
}

func TestClassifyError(t *testing.T) {
	collector := NewValidationMetricsCollector()

	// Test error classification
	tests := []struct {
		err      error
		expected string
	}{
		{nil, string(metrics.ErrorTypeUnknown)},
		{assert.AnError, "unknown"},
	}

	for _, test := range tests {
		result := collector.classifyError(test.err)
		assert.Equal(t, test.expected, result)
	}
}

func TestClear(t *testing.T) {
	collector := NewValidationMetricsCollector()

	// Add some data
	collector.RecordValidation("validate", "User", "insert", 50*time.Millisecond, nil)
	collector.RecordCacheAccess("type", true)
	collector.RecordConstraintCheck("required", "Name", true)

	// Verify data exists
	counters := collector.GetValidationCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("validate", "User", "insert", "success")])

	cacheCounters := collector.GetCacheAccessCounters()
	assert.Equal(t, int64(1), cacheCounters[metrics.BuildKey("type", "hit")])

	constraintCounters := collector.GetConstraintCheckCounters()
	assert.Equal(t, int64(1), constraintCounters[metrics.BuildKey("required", "Name", "passed")])

	// Clear all data
	collector.Clear()

	// Verify data is cleared
	counters = collector.GetValidationCounters()
	assert.Equal(t, 0, len(counters))

	cacheCounters = collector.GetCacheAccessCounters()
	assert.Equal(t, 0, len(cacheCounters))

	constraintCounters = collector.GetConstraintCheckCounters()
	assert.Equal(t, 0, len(constraintCounters))
}

func TestThreadSafety(t *testing.T) {
	collector := NewValidationMetricsCollector()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				model := "User"
				if idx%2 == 0 {
					model = "Product"
				}

				scenario := "insert"
				if j%2 == 0 {
					scenario = "update"
				}

				collector.RecordValidation("validate", model, scenario, time.Duration(j)*time.Millisecond, nil)
				collector.RecordCacheAccess("type", j%3 == 0)
				collector.RecordConstraintCheck("required", "Field", j%5 != 0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counters are consistent
	counters := collector.GetValidationCounters()
	totalValidations := int64(0)
	for _, count := range counters {
		totalValidations += count
	}
	assert.Equal(t, int64(1000), totalValidations) // 10 goroutines * 100 iterations

	cacheCounters := collector.GetCacheAccessCounters()
	totalCacheAccesses := int64(0)
	for _, count := range cacheCounters {
		totalCacheAccesses += count
	}
	assert.Equal(t, int64(1000), totalCacheAccesses)

	constraintCounters := collector.GetConstraintCheckCounters()
	totalConstraintChecks := int64(0)
	for _, count := range constraintCounters {
		totalConstraintChecks += count
	}
	assert.Equal(t, int64(1000), totalConstraintChecks)
}
