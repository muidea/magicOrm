package metricsdb

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseMetricsCollector(t *testing.T) {
	collector := NewDatabaseMetricsCollector()
	assert.NotNil(t, collector)
}

func TestRecordQuery(t *testing.T) {
	collector := NewDatabaseMetricsCollector()

	// Record successful query
	collector.RecordQuery("postgresql", "select", 100*time.Millisecond, nil)

	// Record failed query
	collector.RecordQuery("mysql", "insert", 200*time.Millisecond, assert.AnError)

	// Check query counters
	counters := collector.GetQueryCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "select", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("mysql", "insert", "error")])

	// Check error counters
	errorCounters := collector.GetErrorCounters()
	assert.Equal(t, int64(1), errorCounters[metrics.BuildKey("mysql", "insert", "unknown")])

	// Check durations
	durations := collector.GetQueryDurations()
	assert.Len(t, durations[metrics.BuildKey("postgresql", "select", "success")], 1)
	assert.Len(t, durations[metrics.BuildKey("mysql", "insert", "error")], 1)
}

func TestRecordTransaction(t *testing.T) {
	collector := NewDatabaseMetricsCollector()

	// Record successful transaction
	collector.RecordTransaction("postgresql", "begin", true)
	collector.RecordTransaction("postgresql", "commit", true)

	// Record failed transaction
	collector.RecordTransaction("mysql", "begin", false)

	// Check transaction counters
	counters := collector.GetTransactionCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "begin", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "commit", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("mysql", "begin", "error")])
}

func TestRecordExecution(t *testing.T) {
	collector := NewDatabaseMetricsCollector()

	// Record successful executions
	collector.RecordExecution("postgresql", "insert", true)
	collector.RecordExecution("postgresql", "update", true)
	collector.RecordExecution("postgresql", "delete", true)

	// Record failed execution
	collector.RecordExecution("mysql", "insert", false)

	// Check execution counters
	counters := collector.GetExecutionCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "insert", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "update", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "delete", "success")])
	assert.Equal(t, int64(1), counters[metrics.BuildKey("mysql", "insert", "error")])
}

func TestUpdateConnectionStats(t *testing.T) {
	collector := NewDatabaseMetricsCollector()

	// Update connection statistics
	collector.UpdateConnectionStats("postgresql", "active", 5)
	collector.UpdateConnectionStats("postgresql", "idle", 10)
	collector.UpdateConnectionStats("postgresql", "max", 20)

	collector.UpdateConnectionStats("mysql", "active", 3)
	collector.UpdateConnectionStats("mysql", "idle", 7)

	// Check connection stats
	stats := collector.GetConnectionStats()
	assert.Equal(t, int64(5), stats[metrics.BuildKey("postgresql", "active")])
	assert.Equal(t, int64(10), stats[metrics.BuildKey("postgresql", "idle")])
	assert.Equal(t, int64(20), stats[metrics.BuildKey("postgresql", "max")])
	assert.Equal(t, int64(3), stats[metrics.BuildKey("mysql", "active")])
	assert.Equal(t, int64(7), stats[metrics.BuildKey("mysql", "idle")])
}

func TestClassifyError(t *testing.T) {
	collector := NewDatabaseMetricsCollector()

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
	collector := NewDatabaseMetricsCollector()

	// Add some data
	collector.RecordQuery("postgresql", "select", 100*time.Millisecond, nil)
	collector.RecordTransaction("postgresql", "begin", true)
	collector.UpdateConnectionStats("postgresql", "active", 5)

	// Verify data exists
	counters := collector.GetQueryCounters()
	assert.Equal(t, int64(1), counters[metrics.BuildKey("postgresql", "select", "success")])

	// Clear all data
	collector.Clear()

	// Verify data is cleared
	counters = collector.GetQueryCounters()
	assert.Equal(t, 0, len(counters))

	txCounters := collector.GetTransactionCounters()
	assert.Equal(t, 0, len(txCounters))

	connStats := collector.GetConnectionStats()
	assert.Equal(t, 0, len(connStats))
}

func TestThreadSafety(t *testing.T) {
	collector := NewDatabaseMetricsCollector()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				database := "postgresql"
				if idx%2 == 0 {
					database = "mysql"
				}

				collector.RecordQuery(database, "select", time.Duration(j)*time.Millisecond, nil)
				collector.RecordTransaction(database, "begin", true)
				collector.UpdateConnectionStats(database, "active", int64(j))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counters are consistent
	counters := collector.GetQueryCounters()
	totalQueries := int64(0)
	for _, count := range counters {
		totalQueries += count
	}
	assert.Equal(t, int64(1000), totalQueries) // 10 goroutines * 100 iterations

	txCounters := collector.GetTransactionCounters()
	totalTransactions := int64(0)
	for _, count := range txCounters {
		totalTransactions += count
	}
	assert.Equal(t, int64(1000), totalTransactions)
}
