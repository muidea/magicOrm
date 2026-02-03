// Package monitoring provides lightweight collectors for MagicORM-specific metrics.
// This package only provides data collection functionality.
// Export and management are handled by github.com/muidea/magicCommon/monitoring.
package monitoring

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// Collector defines the interface for MagicORM-specific metric collectors.
// Implementations should focus on collecting domain-specific metrics only.
type Collector interface {
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

	// RecordValidation records a validation operation
	RecordValidation(operation string, modelName string, scenario string, startTime time.Time, err error, labels map[string]string)

	// GetMetrics returns all collected metrics as MetricProvider format
	GetMetrics() ([]types.Metric, error)
}

// OperationType represents the type of ORM operation
type OperationType string

const (
	OperationInsert OperationType = "insert"
	OperationUpdate OperationType = "update"
	OperationQuery  OperationType = "query"
	OperationDelete OperationType = "delete"
	OperationCreate OperationType = "create"
	OperationDrop   OperationType = "drop"
	OperationCount  OperationType = "count"
	OperationBatch  OperationType = "batch"
)

// QueryType represents the type of query
type QueryType string

const (
	QueryTypeSimple   QueryType = "simple"
	QueryTypeFilter   QueryType = "filter"
	QueryTypeRelation QueryType = "relation"
	QueryTypeBatch    QueryType = "batch"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeDatabase    ErrorType = "database"
	ErrorTypeConnection  ErrorType = "connection"
	ErrorTypeTimeout     ErrorType = "timeout"
	ErrorTypeConstraint  ErrorType = "constraint"
	ErrorTypeTransaction ErrorType = "transaction"
	ErrorTypeUnknown     ErrorType = "unknown"
)

// BaseCollector provides common functionality for all collectors
type BaseCollector struct {
	// metrics storage
	operations    []OperationMetric
	queries       []QueryMetric
	transactions  []TransactionMetric
	cacheAccesses []CacheAccessMetric
	validations   []ValidationMetric
	databaseOps   []DatabaseOperationMetric
}

// OperationMetric represents a recorded ORM operation
type OperationMetric struct {
	Operation OperationType
	ModelName string
	StartTime time.Time
	Duration  time.Duration
	Error     error
	Labels    map[string]string
}

// QueryMetric represents a recorded query operation
type QueryMetric struct {
	ModelName    string
	QueryType    QueryType
	RowsReturned int
	StartTime    time.Time
	Duration     time.Duration
	Error        error
	Labels       map[string]string
}

// TransactionMetric represents a recorded transaction
type TransactionMetric struct {
	Operation string
	StartTime time.Time
	Duration  time.Duration
	Error     error
	Labels    map[string]string
}

// CacheAccessMetric represents a recorded cache access
type CacheAccessMetric struct {
	CacheType string
	Operation string
	Hit       bool
	Duration  time.Duration
	Labels    map[string]string
}

// ValidationMetric represents a recorded validation operation
type ValidationMetric struct {
	Operation string
	ModelName string
	Scenario  string
	StartTime time.Time
	Duration  time.Duration
	Error     error
	Labels    map[string]string
}

// DatabaseOperationMetric represents a recorded database operation
type DatabaseOperationMetric struct {
	DBType    string
	Operation string
	StartTime time.Time
	Duration  time.Duration
	Error     error
	Labels    map[string]string
}

// NewBaseCollector creates a new base collector
func NewBaseCollector() *BaseCollector {
	return &BaseCollector{
		operations:    make([]OperationMetric, 0),
		queries:       make([]QueryMetric, 0),
		transactions:  make([]TransactionMetric, 0),
		cacheAccesses: make([]CacheAccessMetric, 0),
		validations:   make([]ValidationMetric, 0),
		databaseOps:   make([]DatabaseOperationMetric, 0),
	}
}

// classifyError classifies an error into ErrorType
func (c *BaseCollector) classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()

	switch {
	case contains(errStr, "validation"):
		return ErrorTypeValidation
	case contains(errStr, "database"):
		return ErrorTypeDatabase
	case contains(errStr, "connection"):
		return ErrorTypeConnection
	case contains(errStr, "timeout"):
		return ErrorTypeTimeout
	case contains(errStr, "constraint"):
		return ErrorTypeConstraint
	case contains(errStr, "transaction"):
		return ErrorTypeTransaction
	default:
		return ErrorTypeUnknown
	}
}

// contains helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// DefaultLabels returns default labels for metrics
func DefaultLabels() map[string]string {
	return map[string]string{
		"component": "magicorm",
		"version":   "1.0.0",
	}
}

// MergeLabels merges multiple label maps
func MergeLabels(labels ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, labelMap := range labels {
		for k, v := range labelMap {
			result[k] = v
		}
	}

	return result
}

// SimpleCollector is a simple implementation of Collector interface
// This is for backward compatibility during migration
type SimpleCollector struct {
	*BaseCollector
}

// NewSimpleCollector creates a new simple collector
func NewSimpleCollector() *SimpleCollector {
	return &SimpleCollector{
		BaseCollector: NewBaseCollector(),
	}
}

// RecordOperation implements Collector interface
func (c *SimpleCollector) RecordOperation(operation OperationType, modelName string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.operations = append(c.operations, OperationMetric{
		Operation: operation,
		ModelName: modelName,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordQuery implements Collector interface
func (c *SimpleCollector) RecordQuery(modelName string, queryType QueryType, rowsReturned int, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.queries = append(c.queries, QueryMetric{
		ModelName:    modelName,
		QueryType:    queryType,
		RowsReturned: rowsReturned,
		StartTime:    startTime,
		Duration:     duration,
		Error:        err,
		Labels:       MergeLabels(DefaultLabels(), labels),
	})
}

// RecordTransaction implements Collector interface
func (c *SimpleCollector) RecordTransaction(operation string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.transactions = append(c.transactions, TransactionMetric{
		Operation: operation,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordCacheAccess implements Collector interface
func (c *SimpleCollector) RecordCacheAccess(cacheType string, operation string, hit bool, duration time.Duration, labels map[string]string) {
	c.cacheAccesses = append(c.cacheAccesses, CacheAccessMetric{
		CacheType: cacheType,
		Operation: operation,
		Hit:       hit,
		Duration:  duration,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordDatabaseOperation implements Collector interface
func (c *SimpleCollector) RecordDatabaseOperation(dbType string, operation string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.databaseOps = append(c.databaseOps, DatabaseOperationMetric{
		DBType:    dbType,
		Operation: operation,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordValidation implements Collector interface
func (c *SimpleCollector) RecordValidation(operation string, modelName string, scenario string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.validations = append(c.validations, ValidationMetric{
		Operation: operation,
		ModelName: modelName,
		Scenario:  scenario,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// GetMetrics implements Collector interface
func (c *SimpleCollector) GetMetrics() ([]types.Metric, error) {
	// For now, return empty metrics
	// In real implementation, this would convert collected metrics to types.Metric format
	return []types.Metric{}, nil
}

// RecordOperation implements Collector interface for BaseCollector
func (c *BaseCollector) RecordOperation(operation OperationType, modelName string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.operations = append(c.operations, OperationMetric{
		Operation: operation,
		ModelName: modelName,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordQuery implements Collector interface for BaseCollector
func (c *BaseCollector) RecordQuery(modelName string, queryType QueryType, rowsReturned int, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.queries = append(c.queries, QueryMetric{
		ModelName:    modelName,
		QueryType:    queryType,
		RowsReturned: rowsReturned,
		StartTime:    startTime,
		Duration:     duration,
		Error:        err,
		Labels:       MergeLabels(DefaultLabels(), labels),
	})
}

// RecordTransaction implements Collector interface for BaseCollector
func (c *BaseCollector) RecordTransaction(operation string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.transactions = append(c.transactions, TransactionMetric{
		Operation: operation,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordCacheAccess implements Collector interface for BaseCollector
func (c *BaseCollector) RecordCacheAccess(cacheType string, operation string, hit bool, duration time.Duration, labels map[string]string) {
	c.cacheAccesses = append(c.cacheAccesses, CacheAccessMetric{
		CacheType: cacheType,
		Operation: operation,
		Hit:       hit,
		Duration:  duration,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordDatabaseOperation implements Collector interface for BaseCollector
func (c *BaseCollector) RecordDatabaseOperation(dbType string, operation string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.databaseOps = append(c.databaseOps, DatabaseOperationMetric{
		DBType:    dbType,
		Operation: operation,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// RecordValidation implements Collector interface for BaseCollector
func (c *BaseCollector) RecordValidation(operation string, modelName string, scenario string, startTime time.Time, err error, labels map[string]string) {
	duration := time.Since(startTime)

	c.validations = append(c.validations, ValidationMetric{
		Operation: operation,
		ModelName: modelName,
		Scenario:  scenario,
		StartTime: startTime,
		Duration:  duration,
		Error:     err,
		Labels:    MergeLabels(DefaultLabels(), labels),
	})
}

// GetMetrics implements Collector interface for BaseCollector
func (c *BaseCollector) GetMetrics() ([]types.Metric, error) {
	// For now, return empty metrics
	// In real implementation, this would convert collected metrics to types.Metric format
	return []types.Metric{}, nil
}
