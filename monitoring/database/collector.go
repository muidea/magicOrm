// Package database provides database-specific metric collectors.
package database

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// DatabaseCollector defines the interface for database-specific metric collectors.
type DatabaseCollector interface {
	// RecordQuery records a database query operation
	RecordQuery(dbType string, queryType string, rowsAffected int, startTime time.Time, err error, labels map[string]string)

	// RecordTransaction records a database transaction operation
	RecordTransaction(dbType string, operation string, startTime time.Time, err error, labels map[string]string)

	// RecordExecution records a SQL execution operation
	RecordExecution(dbType string, operation string, duration time.Duration, err error, labels map[string]string)

	// RecordConnection records a database connection operation
	RecordConnection(dbType string, operation string, startTime time.Time, err error, labels map[string]string)

	// RecordConnectionPool records connection pool statistics
	RecordConnectionPool(dbType string, activeConnections int, idleConnections int, waitingConnections int, maxConnections int, labels map[string]string)

	// GetMetrics returns collected database metrics
	GetMetrics() ([]types.Metric, error)
}

// SimpleDatabaseCollector is a simple implementation
type SimpleDatabaseCollector struct {
	queries         []queryMetric
	transactions    []transactionMetric
	connections     []connectionMetric
	connectionPools []connectionPoolMetric
}

type queryMetric struct {
	dbType       string
	queryType    string
	rowsAffected int
	startTime    time.Time
	err          error
	labels       map[string]string
}

type transactionMetric struct {
	dbType    string
	operation string
	startTime time.Time
	err       error
	labels    map[string]string
}

type connectionMetric struct {
	dbType    string
	operation string
	startTime time.Time
	err       error
	labels    map[string]string
}

type connectionPoolMetric struct {
	dbType             string
	activeConnections  int
	idleConnections    int
	waitingConnections int
	maxConnections     int
	labels             map[string]string
}

// NewCollector creates a new database collector
func NewCollector() DatabaseCollector {
	return &SimpleDatabaseCollector{
		queries:         make([]queryMetric, 0),
		transactions:    make([]transactionMetric, 0),
		connections:     make([]connectionMetric, 0),
		connectionPools: make([]connectionPoolMetric, 0),
	}
}

// RecordQuery implements DatabaseCollector interface
func (c *SimpleDatabaseCollector) RecordQuery(dbType string, queryType string, rowsAffected int, startTime time.Time, err error, labels map[string]string) {
	c.queries = append(c.queries, queryMetric{
		dbType:       dbType,
		queryType:    queryType,
		rowsAffected: rowsAffected,
		startTime:    startTime,
		err:          err,
		labels:       labels,
	})
}

// RecordTransaction implements DatabaseCollector interface
func (c *SimpleDatabaseCollector) RecordTransaction(dbType string, operation string, startTime time.Time, err error, labels map[string]string) {
	c.transactions = append(c.transactions, transactionMetric{
		dbType:    dbType,
		operation: operation,
		startTime: startTime,
		err:       err,
		labels:    labels,
	})
}

// RecordExecution implements DatabaseCollector interface
func (c *SimpleDatabaseCollector) RecordExecution(dbType string, operation string, duration time.Duration, err error, labels map[string]string) {
	// Simplified implementation
	startTime := time.Now().Add(-duration)
	c.RecordQuery(dbType, operation, 0, startTime, err, labels)
}

// RecordConnection implements DatabaseCollector interface
func (c *SimpleDatabaseCollector) RecordConnection(dbType string, operation string, startTime time.Time, err error, labels map[string]string) {
	c.connections = append(c.connections, connectionMetric{
		dbType:    dbType,
		operation: operation,
		startTime: startTime,
		err:       err,
		labels:    labels,
	})
}

// RecordConnectionPool implements DatabaseCollector interface
func (c *SimpleDatabaseCollector) RecordConnectionPool(dbType string, activeConnections int, idleConnections int, waitingConnections int, maxConnections int, labels map[string]string) {
	c.connectionPools = append(c.connectionPools, connectionPoolMetric{
		dbType:             dbType,
		activeConnections:  activeConnections,
		idleConnections:    idleConnections,
		waitingConnections: waitingConnections,
		maxConnections:     maxConnections,
		labels:             labels,
	})
}

// GetMetrics implements DatabaseCollector interface
func (c *SimpleDatabaseCollector) GetMetrics() ([]types.Metric, error) {
	// Return empty metrics for now
	return []types.Metric{}, nil
}

// Helper method for RecordError (optional)
func (c *SimpleDatabaseCollector) RecordError(dbType string, operation string, errorType string, labels map[string]string) {
	// Simplified implementation
	startTime := time.Now()
	err := &DatabaseError{Type: errorType}
	c.RecordQuery(dbType, operation, 0, startTime, err, labels)
}

// DatabaseError represents a database error
type DatabaseError struct {
	Type string
}

func (e *DatabaseError) Error() string {
	return "database error: " + e.Type
}
