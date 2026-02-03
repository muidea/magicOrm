// Package database provides database monitoring adapters.
// This is a simplified version that only provides monitoring collection.
package database

import (
	"fmt"
	"time"
)

// DatabaseMonitor is a simplified monitor for database operations.
type DatabaseMonitor struct {
	collector DatabaseCollector
	enabled   bool
}

// NewDatabaseMonitor creates a new database monitor.
func NewDatabaseMonitor(collector DatabaseCollector) *DatabaseMonitor {
	return &DatabaseMonitor{
		collector: collector,
		enabled:   collector != nil,
	}
}

// RecordQuery records a database query.
func (m *DatabaseMonitor) RecordQuery(
	dbType string,
	queryType string,
	success bool,
	duration time.Duration,
	rowsAffected int,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	startTime := time.Now().Add(-duration)
	labels := mergeLabels(additionalLabels)
	labels["rows_affected"] = fmt.Sprintf("%d", rowsAffected)

	var err error
	if !success {
		err = &QueryError{Message: "query failed"}
	}

	m.collector.RecordQuery(dbType, queryType, rowsAffected, startTime, err, labels)
}

// RecordTransaction records a database transaction.
func (m *DatabaseMonitor) RecordTransaction(
	dbType string,
	operation string,
	success bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	startTime := time.Now().Add(-duration)
	labels := mergeLabels(additionalLabels)

	var err error
	if !success {
		err = &TransactionError{Message: "transaction failed"}
	}

	m.collector.RecordTransaction(dbType, operation, startTime, err, labels)
}

// RecordExecution records a SQL execution.
func (m *DatabaseMonitor) RecordExecution(
	dbType string,
	operation string,
	success bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	labels := mergeLabels(additionalLabels)

	var err error
	if !success {
		err = &ExecutionError{Message: "execution failed"}
	}

	m.collector.RecordExecution(dbType, operation, duration, err, labels)
}

// RecordConnection records a database connection operation.
func (m *DatabaseMonitor) RecordConnection(
	dbType string,
	operation string,
	success bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	startTime := time.Now().Add(-duration)
	labels := mergeLabels(additionalLabels)

	var err error
	if !success {
		err = &ConnectionError{Message: "connection failed"}
	}

	m.collector.RecordConnection(dbType, operation, startTime, err, labels)
}

// RecordConnectionPool records connection pool statistics.
func (m *DatabaseMonitor) RecordConnectionPool(
	dbType string,
	activeConnections int,
	idleConnections int,
	waitingConnections int,
	maxConnections int,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	labels := mergeLabels(additionalLabels)
	m.collector.RecordConnectionPool(dbType, activeConnections, idleConnections, waitingConnections, maxConnections, labels)
}

// RecordError records a database error.
func (m *DatabaseMonitor) RecordError(
	dbType string,
	operation string,
	errorType string,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	labels := mergeLabels(additionalLabels)

	// Check if collector supports RecordError method
	if collector, ok := m.collector.(interface {
		RecordError(dbType string, operation string, errorType string, labels map[string]string)
	}); ok {
		collector.RecordError(dbType, operation, errorType, labels)
	}
}

// Helper functions

func mergeLabels(additional map[string]string) map[string]string {
	labels := make(map[string]string)
	if additional != nil {
		for k, v := range additional {
			labels[k] = v
		}
	}
	return labels
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// Error types

type QueryError struct{ Message string }

func (e *QueryError) Error() string { return e.Message }

type TransactionError struct{ Message string }

func (e *TransactionError) Error() string { return e.Message }

type ExecutionError struct{ Message string }

func (e *ExecutionError) Error() string { return e.Message }

type ConnectionError struct{ Message string }

func (e *ConnectionError) Error() string { return e.Message }

// SimpleDatabaseMonitor for backward compatibility
type SimpleDatabaseMonitor struct {
	*DatabaseMonitor
}

func NewSimpleDatabaseMonitor(collector DatabaseCollector) *SimpleDatabaseMonitor {
	return &SimpleDatabaseMonitor{
		DatabaseMonitor: NewDatabaseMonitor(collector),
	}
}
