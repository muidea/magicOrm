package metrics

import (
	"testing"
)

func TestMetricsTypeDefinitions(t *testing.T) {
	// Test that the metrics package provides the expected type definitions

	// Test that operation types are defined
	if OperationInsert != "insert" {
		t.Errorf("OperationInsert should be 'insert', got '%s'", OperationInsert)
	}

	if OperationQuery != "query" {
		t.Errorf("OperationQuery should be 'query', got '%s'", OperationQuery)
	}

	// Test that query types are defined
	if QueryTypeSimple != "simple" {
		t.Errorf("QueryTypeSimple should be 'simple', got '%s'", QueryTypeSimple)
	}

	// Test that error types are defined
	if ErrorTypeDatabase != "database" {
		t.Errorf("ErrorTypeDatabase should be 'database', got '%s'", ErrorTypeDatabase)
	}

	// Test DefaultLabels
	labels := DefaultLabels()
	if labels["component"] != "magicorm" {
		t.Errorf("Expected component label 'magicorm', got '%s'", labels["component"])
	}
}

func TestCrossComponentMetrics(t *testing.T) {
	// Test that different metrics components work together

	// Test MergeLabels function
	labels1 := map[string]string{"a": "1"}
	labels2 := map[string]string{"b": "2"}
	merged := MergeLabels(labels1, labels2)

	if merged["a"] != "1" {
		t.Errorf("Expected a=1, got %s", merged["a"])
	}

	if merged["b"] != "2" {
		t.Errorf("Expected b=2, got %s", merged["b"])
	}

	// Test that all type constants are accessible
	if OperationCreate != "create" {
		t.Errorf("OperationCreate should be 'create'")
	}

	if QueryTypeFilter != "filter" {
		t.Errorf("QueryTypeFilter should be 'filter'")
	}

	if ErrorTypeValidation != "validation" {
		t.Errorf("ErrorTypeValidation should be 'validation'")
	}
}

func TestMetricsErrorHandling(t *testing.T) {
	// Test error type constants in metrics

	// Test that all error types are defined
	errorTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeDatabase,
		ErrorTypeConnection,
		ErrorTypeTimeout,
		ErrorTypeConstraint,
		ErrorTypeTransaction,
		ErrorTypeUnknown,
	}

	for _, errType := range errorTypes {
		if string(errType) == "" {
			t.Error("Error type should not be empty string")
		}
	}

	// Test MergeLabels with error cases
	// nil maps should be handled gracefully
	result := MergeLabels(nil, map[string]string{"a": "1"}, nil)
	if result["a"] != "1" {
		t.Errorf("MergeLabels should handle nil maps")
	}

	// Empty map
	empty := MergeLabels()
	if len(empty) != 0 {
		t.Errorf("MergeLabels with no arguments should return empty map")
	}
}

func TestMetricsConfiguration(t *testing.T) {
	// Test metrics configuration
	// Providers are auto-registered via init() in subpackages

	// Test that DefaultLabels returns expected values
	labels := DefaultLabels()
	if labels == nil {
		t.Error("DefaultLabels should not return nil")
	}
	if labels["component"] != "magicorm" {
		t.Errorf("Expected component label 'magicorm', got '%s'", labels["component"])
	}
}

func TestOperationTypeConstants(t *testing.T) {
	// Test that operation type constants are properly defined
	tests := []struct {
		name     string
		actual   OperationType
		expected string
	}{
		{"OperationInsert", OperationInsert, "insert"},
		{"OperationUpdate", OperationUpdate, "update"},
		{"OperationQuery", OperationQuery, "query"},
		{"OperationDelete", OperationDelete, "delete"},
		{"OperationCreate", OperationCreate, "create"},
		{"OperationDrop", OperationDrop, "drop"},
		{"OperationCount", OperationCount, "count"},
		{"OperationBatch", OperationBatch, "batch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.actual) != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestQueryTypeConstants(t *testing.T) {
	// Test that query type constants are properly defined
	tests := []struct {
		name     string
		actual   QueryType
		expected string
	}{
		{"QueryTypeSimple", QueryTypeSimple, "simple"},
		{"QueryTypeFilter", QueryTypeFilter, "filter"},
		{"QueryTypeRelation", QueryTypeRelation, "relation"},
		{"QueryTypeBatch", QueryTypeBatch, "batch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.actual) != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestErrorTypeConstants(t *testing.T) {
	// Test that error type constants are properly defined
	tests := []struct {
		name     string
		actual   ErrorType
		expected string
	}{
		{"ErrorTypeValidation", ErrorTypeValidation, "validation"},
		{"ErrorTypeDatabase", ErrorTypeDatabase, "database"},
		{"ErrorTypeConnection", ErrorTypeConnection, "connection"},
		{"ErrorTypeTimeout", ErrorTypeTimeout, "timeout"},
		{"ErrorTypeConstraint", ErrorTypeConstraint, "constraint"},
		{"ErrorTypeTransaction", ErrorTypeTransaction, "transaction"},
		{"ErrorTypeUnknown", ErrorTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.actual) != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestDefaultLabels(t *testing.T) {
	labels := DefaultLabels()

	if labels["component"] != "magicorm" {
		t.Errorf("Expected component label 'magicorm', got '%s'", labels["component"])
	}

	// Version might vary, but it should be set
	if labels["version"] == "" {
		t.Error("Version label should not be empty")
	}
}

func TestMergeLabels(t *testing.T) {
	// Test merging multiple label maps
	labels1 := map[string]string{"a": "1", "b": "2"}
	labels2 := map[string]string{"b": "overridden", "c": "3"}
	labels3 := map[string]string{"d": "4"}

	result := MergeLabels(labels1, labels2, labels3)

	if result["a"] != "1" {
		t.Errorf("Expected a=1, got %s", result["a"])
	}

	// labels2 should override labels1
	if result["b"] != "overridden" {
		t.Errorf("Expected b=overridden, got %s", result["b"])
	}

	if result["c"] != "3" {
		t.Errorf("Expected c=3, got %s", result["c"])
	}

	if result["d"] != "4" {
		t.Errorf("Expected d=4, got %s", result["d"])
	}

	// Test with nil maps
	result2 := MergeLabels(nil, labels1, nil)
	if result2["a"] != "1" {
		t.Errorf("Expected a=1 with nil maps, got %s", result2["a"])
	}

	// Test with empty map
	result3 := MergeLabels(map[string]string{})
	if len(result3) != 0 {
		t.Errorf("Expected empty map, got %v", result3)
	}
}
