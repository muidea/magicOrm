// Package metrics provides MagicORM-specific metric types and constants.
// This package provides only type definitions for integration with magicCommon/monitoring.
// Metric collection and export are handled by github.com/muidea/magicCommon/monitoring.
package metrics

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
