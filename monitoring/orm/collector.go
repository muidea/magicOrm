// Package orm provides ORM-specific metric collectors.
// This package focuses on collecting ORM operation metrics only.
package orm

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/monitoring"
)

// Use types from the monitoring package
type OperationType = monitoring.OperationType
type QueryType = monitoring.QueryType

// ORMCollector defines the interface for ORM-specific metric collectors.
type ORMCollector interface {
	// RecordOperation records an ORM operation
	RecordOperation(operation OperationType, modelName string, startTime time.Time, err error, labels map[string]string)

	// RecordQuery records a query operation with additional details
	RecordQuery(modelName string, queryType QueryType, rowsReturned int, startTime time.Time, err error, labels map[string]string)

	// RecordTransaction records a transaction operation
	RecordTransaction(operation string, startTime time.Time, err error, labels map[string]string)

	// RecordCacheAccess records cache access for ORM
	RecordCacheAccess(cacheType string, operation string, hit bool, duration time.Duration, labels map[string]string)

	// RecordDatabaseOperation records a database-level operation
	RecordDatabaseOperation(dbType string, operation string, startTime time.Time, err error, labels map[string]string)

	// GetMetrics returns collected ORM metrics
	GetMetrics() ([]types.Metric, error)
}

// SimpleORMCollector is a simplified collector that only records metrics without complex logic
type SimpleORMCollector struct {
	operations   []operationMetric
	queries      []queryMetric
	transactions []transactionMetric
}

type operationMetric struct {
	operation OperationType
	modelName string
	startTime time.Time
	err       error
	labels    map[string]string
}

type queryMetric struct {
	modelName    string
	queryType    QueryType
	rowsReturned int
	startTime    time.Time
	err          error
	labels       map[string]string
}

type transactionMetric struct {
	operation string
	startTime time.Time
	err       error
	labels    map[string]string
}

// NewCollector creates a new ORM collector
func NewCollector() ORMCollector {
	return &SimpleORMCollector{
		operations:   make([]operationMetric, 0),
		queries:      make([]queryMetric, 0),
		transactions: make([]transactionMetric, 0),
	}
}

// RecordOperation implements ORMCollector interface
func (c *SimpleORMCollector) RecordOperation(operation OperationType, modelName string, startTime time.Time, err error, labels map[string]string) {
	c.operations = append(c.operations, operationMetric{
		operation: operation,
		modelName: modelName,
		startTime: startTime,
		err:       err,
		labels:    labels,
	})
}

// RecordQuery implements ORMCollector interface
func (c *SimpleORMCollector) RecordQuery(modelName string, queryType QueryType, rowsReturned int, startTime time.Time, err error, labels map[string]string) {
	c.queries = append(c.queries, queryMetric{
		modelName:    modelName,
		queryType:    queryType,
		rowsReturned: rowsReturned,
		startTime:    startTime,
		err:          err,
		labels:       labels,
	})
}

// RecordTransaction implements ORMCollector interface
func (c *SimpleORMCollector) RecordTransaction(operation string, startTime time.Time, err error, labels map[string]string) {
	c.transactions = append(c.transactions, transactionMetric{
		operation: operation,
		startTime: startTime,
		err:       err,
		labels:    labels,
	})
}

// RecordCacheAccess implements ORMCollector interface
func (c *SimpleORMCollector) RecordCacheAccess(cacheType string, operation string, hit bool, duration time.Duration, labels map[string]string) {
	// Simplified implementation - just store the data
	// In a real implementation, this would update cache statistics
}

// RecordDatabaseOperation implements ORMCollector interface
func (c *SimpleORMCollector) RecordDatabaseOperation(dbType string, operation string, startTime time.Time, err error, labels map[string]string) {
	// This would typically be handled by the database collector
}

// GetMetrics implements ORMCollector interface
func (c *SimpleORMCollector) GetMetrics() ([]types.Metric, error) {
	// Return empty metrics for now
	// In a real implementation, this would convert stored metrics to types.Metric format
	return []types.Metric{}, nil
}
