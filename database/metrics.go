package database

import (
	"database/sql"
	"strings"
	"time"

	"github.com/muidea/magicOrm/metrics/metricsdb"
)

const (
	DatabasePostgreSQL = "postgresql"
	DatabaseMySQL      = "mysql"
)

func RecordDatabaseQuery(databaseName string, sqlText string, duration time.Duration, err error) {
	collector := metricsdb.GetDatabaseMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordQuery(databaseName, normalizeDatabaseOperation(sqlText, "query"), duration, err)
}

func RecordDatabaseExecution(databaseName string, sqlText string, success bool) {
	collector := metricsdb.GetDatabaseMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordExecution(databaseName, normalizeDatabaseOperation(sqlText, "execute"), success)
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
	collector.UpdateConnectionStats(databaseName, "active", int64(stats.InUse))
	collector.UpdateConnectionStats(databaseName, "idle", int64(stats.Idle))
	collector.UpdateConnectionStats(databaseName, "open", int64(stats.OpenConnections))
	collector.UpdateConnectionStats(databaseName, "max", int64(stats.MaxOpenConnections))
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
