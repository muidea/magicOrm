package database

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/metrics/metricsdb"
)

const (
	DatabasePostgreSQL = "postgresql"
	DatabaseMySQL      = "mysql"
)

type connectionStatsSnapshot struct {
	active int64
	idle   int64
	open   int64
	max    int64
}

var lastConnectionStats sync.Map
var normalizedOperationCache sync.Map

func RecordDatabaseQuery(databaseName string, sqlText string, duration time.Duration, err *cd.Error) {
	collector := metricsdb.GetDatabaseMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordQuery(databaseName, cachedDatabaseOperation(sqlText, "query"), duration, cd.ToStdError(err))
}

func RecordDatabaseExecution(databaseName string, sqlText string, success bool) {
	collector := metricsdb.GetDatabaseMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordExecution(databaseName, cachedDatabaseOperation(sqlText, "execute"), success)
}

func RecordDatabaseTransaction(databaseName string, txType string, success bool) {
	collector := metricsdb.GetDatabaseMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordTransaction(databaseName, txType, success)
}

func UpdateDatabaseConnectionStats(databaseName string, dbHandle *sql.DB) {
	collector := metricsdb.GetDatabaseMetricsCollector()
	if collector == nil || dbHandle == nil {
		return
	}

	stats := dbHandle.Stats()
	current := connectionStatsSnapshot{
		active: int64(stats.InUse),
		idle:   int64(stats.Idle),
		open:   int64(stats.OpenConnections),
		max:    int64(stats.MaxOpenConnections),
	}

	if pre, ok := lastConnectionStats.Load(databaseName); ok && pre == current {
		return
	}

	lastConnectionStats.Store(databaseName, current)
	collector.UpdateConnectionStats(databaseName, "active", current.active)
	collector.UpdateConnectionStats(databaseName, "idle", current.idle)
	collector.UpdateConnectionStats(databaseName, "open", current.open)
	collector.UpdateConnectionStats(databaseName, "max", current.max)
}

func normalizeDatabaseOperation(sqlText string, fallback string) string {
	sqlText = strings.TrimSpace(strings.ToLower(sqlText))
	if sqlText == "" {
		return fallback
	}

	if strings.HasPrefix(sqlText, "with ") {
		if idx := strings.LastIndex(sqlText, ")"); idx >= 0 && idx+1 < len(sqlText) {
			remaining := strings.TrimSpace(sqlText[idx+1:])
			if remaining != "" {
				sqlText = remaining
			}
		}
	}

	fields := strings.Fields(sqlText)
	if len(fields) == 0 {
		return fallback
	}

	switch fields[0] {
	case "select", "insert", "update", "delete", "create", "drop", "alter", "truncate", "replace", "show", "describe", "explain":
		return fields[0]
	default:
		return fallback
	}
}

func cachedDatabaseOperation(sqlText string, fallback string) string {
	if sqlText == "" {
		return fallback
	}

	cacheKey := fallback + "\x00" + sqlText
	if op, ok := normalizedOperationCache.Load(cacheKey); ok {
		return op.(string)
	}

	op := normalizeDatabaseOperation(sqlText, fallback)
	normalizedOperationCache.Store(cacheKey, op)
	return op
}
