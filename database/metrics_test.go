package database

import (
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
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
	RecordDatabaseQuery(DatabasePostgreSQL, "WITH cte AS (SELECT 1) SELECT * FROM users", 30*time.Millisecond, cd.NewError(cd.Unexpected, errors.New("connection reset").Error()))

	assert.Equal(t, int64(1), collector.GetQueryCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", "success")])
	assert.Equal(t, int64(1), collector.GetQueryCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", "error")])
	assert.Equal(t, int64(1), collector.GetErrorCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", string(metrics.ErrorTypeConnection))])
}

func TestRecordDatabaseQueryTreatsTypedNilErrorAsSuccess(t *testing.T) {
	oldCollector := metricsdb.GetDatabaseMetricsCollector()
	collector := metricsdb.NewDatabaseMetricsCollector()
	metricsdb.SetDatabaseMetricsCollectorForTest(collector)
	defer metricsdb.SetDatabaseMetricsCollectorForTest(oldCollector)

	var typedNilErr *cd.Error
	RecordDatabaseQuery(DatabasePostgreSQL, "SELECT * FROM users", 25*time.Millisecond, typedNilErr)

	assert.Equal(t, int64(1), collector.GetQueryCounters()[metrics.BuildKey(DatabasePostgreSQL, "select", "success")])
	assert.Empty(t, collector.GetErrorCounters())
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

func TestRecordDatabaseTransaction(t *testing.T) {
	oldCollector := metricsdb.GetDatabaseMetricsCollector()
	collector := metricsdb.NewDatabaseMetricsCollector()
	metricsdb.SetDatabaseMetricsCollectorForTest(collector)
	defer metricsdb.SetDatabaseMetricsCollectorForTest(oldCollector)

	RecordDatabaseTransaction(DatabasePostgreSQL, "begin", true)
	RecordDatabaseTransaction(DatabasePostgreSQL, "rollback", false)

	assert.Equal(t, int64(1), collector.GetTransactionCounters()[metrics.BuildKey(DatabasePostgreSQL, "begin", "success")])
	assert.Equal(t, int64(1), collector.GetTransactionCounters()[metrics.BuildKey(DatabasePostgreSQL, "rollback", "error")])
}

func TestNormalizeDatabaseOperation(t *testing.T) {
	tests := []struct {
		name     string
		sqlText  string
		fallback string
		want     string
	}{
		{name: "blank sql", sqlText: "   ", fallback: "query", want: "query"},
		{name: "simple select", sqlText: "SELECT * FROM users", fallback: "query", want: "select"},
		{name: "uppercase ddl", sqlText: "ALTER TABLE users ADD COLUMN age INT", fallback: "execute", want: "alter"},
		{name: "cte select", sqlText: "WITH cte AS (SELECT 1) SELECT * FROM users", fallback: "query", want: "select"},
		{name: "unknown statement", sqlText: "VACUUM users", fallback: "execute", want: "execute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeDatabaseOperation(tt.sqlText, tt.fallback))
		})
	}
}

func TestDatabaseMetricHelpersWithoutCollector(t *testing.T) {
	oldCollector := metricsdb.GetDatabaseMetricsCollector()
	metricsdb.SetDatabaseMetricsCollectorForTest(nil)
	defer metricsdb.SetDatabaseMetricsCollectorForTest(oldCollector)

	RecordDatabaseQuery(DatabasePostgreSQL, "SELECT 1", 0, nil)
	RecordDatabaseExecution(DatabaseMySQL, "INSERT INTO users(id) VALUES(1)", true)
	RecordDatabaseTransaction(DatabasePostgreSQL, "begin", true)
	UpdateDatabaseConnectionStats(DatabasePostgreSQL, nil)
}

func TestCachedDatabaseOperation(t *testing.T) {
	normalizedOperationCache = sync.Map{}

	op := cachedDatabaseOperation("SELECT * FROM users", "query")
	assert.Equal(t, "select", op)

	cached, ok := normalizedOperationCache.Load("query\x00SELECT * FROM users")
	assert.True(t, ok)
	assert.Equal(t, "select", cached)
}
