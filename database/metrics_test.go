package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/metrics/metricsdb"
	"github.com/stretchr/testify/assert"
)

func TestRecordDatabaseQuery(t *testing.T) {
	oldCollector := metricsdb.GetDatabaseMetricsCollector()
	collector := metricsdb.NewDatabaseMetricsCollector()
	metricsdb.SetDatabaseMetricsCollectorForTest(collector)
	defer metricsdb.SetDatabaseMetricsCollectorForTest(oldCollector)

	RecordDatabaseQuery(DatabasePostgreSQL, "SELECT * FROM users", 25*time.Millisecond, nil)
	RecordDatabaseQuery(DatabasePostgreSQL, "WITH cte AS (SELECT 1) SELECT * FROM users", 30*time.Millisecond, errors.New("connection reset"))

	assert.Equal(t, int64(1), collector.GetQueryCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", "success")])
	assert.Equal(t, int64(1), collector.GetQueryCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", "error")])
	assert.Equal(t, int64(1), collector.GetErrorCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", string(metrics.ErrorTypeConnection))])
}

func TestRecordDatabaseExecution(t *testing.T) {
	oldCollector := metricsdb.GetDatabaseMetricsCollector()
	collector := metricsdb.NewDatabaseMetricsCollector()
	metricsdb.SetDatabaseMetricsCollectorForTest(collector)
	defer metricsdb.SetDatabaseMetricsCollectorForTest(oldCollector)

	RecordDatabaseExecution(DatabaseMySQL, "INSERT INTO users(name) VALUES(?)", true)
	RecordDatabaseExecution(DatabaseMySQL, "ALTER TABLE users ADD COLUMN age INT", false)

	assert.Equal(t, int64(1), collector.GetExecutionCounters()[metrics.BuildKey(DatabaseMySQL, "insert", "success")])
	assert.Equal(t, int64(1), collector.GetExecutionCounters()[metrics.BuildKey(DatabaseMySQL, "alter", "error")])
}

func TestUpdateDatabaseConnectionStats(t *testing.T) {
	oldCollector := metricsdb.GetDatabaseMetricsCollector()
	collector := metricsdb.NewDatabaseMetricsCollector()
	metricsdb.SetDatabaseMetricsCollectorForTest(collector)
	defer metricsdb.SetDatabaseMetricsCollectorForTest(oldCollector)

	dbHandle := &sql.DB{}
	UpdateDatabaseConnectionStats(DatabasePostgreSQL, dbHandle)

	stats := collector.GetConnectionStats()
	assert.Equal(t, int64(0), stats[metrics.BuildKey(DatabasePostgreSQL, "active")])
	assert.Equal(t, int64(0), stats[metrics.BuildKey(DatabasePostgreSQL, "idle")])
	assert.Equal(t, int64(0), stats[metrics.BuildKey(DatabasePostgreSQL, "open")])
	assert.Equal(t, int64(0), stats[metrics.BuildKey(DatabasePostgreSQL, "max")])
}
