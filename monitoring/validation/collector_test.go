package validation

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/monitoring/database"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("Expected collector to be created")
	}

	// Verify it's the right type
	_, ok := collector.(*SimpleValidationCollector)
	if !ok {
		t.Error("Expected SimpleValidationCollector type")
	}
}

func TestSimpleValidationCollectorRecordValidation(t *testing.T) {
	collector := NewCollector().(*SimpleValidationCollector)
	startTime := time.Now()

	// Test successful validation
	collector.RecordValidation("validate", "User", "insert", startTime, nil, map[string]string{"test": "label"})

	// The BaseCollector stores validations internally, but we can't access them directly
	// Just verify the method doesn't panic

	// Test validation with error
	testErr := &database.QueryError{Message: "validation failed"}
	collector.RecordValidation("constraint", "Product", "update", startTime, testErr, nil)

	// Verify no panic
}

func TestSimpleValidationCollectorRecordCacheAccess(t *testing.T) {
	collector := NewCollector().(*SimpleValidationCollector)

	// Test cache access recording
	collector.RecordCacheAccess("constraint", "get", true, 10*time.Millisecond, map[string]string{"test": "label"})
	collector.RecordCacheAccess("rule", "set", false, 5*time.Millisecond, nil)

	// Verify no panic
}

func TestSimpleValidationCollectorRecordConstraintCheck(t *testing.T) {
	collector := NewCollector().(*SimpleValidationCollector)

	// Test successful constraint check
	collector.RecordConstraintCheck("required", "email", true, 1*time.Millisecond, map[string]string{"test": "label"})

	if len(collector.constraintChecks) != 1 {
		t.Fatalf("Expected 1 constraint check, got %d", len(collector.constraintChecks))
	}

	constraintCheck := collector.constraintChecks[0]
	if constraintCheck.ConstraintType != "required" {
		t.Errorf("Expected ConstraintType 'required', got '%s'", constraintCheck.ConstraintType)
	}

	if constraintCheck.FieldName != "email" {
		t.Errorf("Expected FieldName 'email', got '%s'", constraintCheck.FieldName)
	}

	if !constraintCheck.Passed {
		t.Error("Expected Passed to be true")
	}

	if constraintCheck.Duration != 1*time.Millisecond {
		t.Errorf("Expected Duration 1ms, got %v", constraintCheck.Duration)
	}

	if constraintCheck.Labels["test"] != "label" {
		t.Errorf("Expected label 'label', got '%s'", constraintCheck.Labels["test"])
	}

	// Test failed constraint check
	collector.RecordConstraintCheck("min", "age", false, 2*time.Millisecond, nil)

	if len(collector.constraintChecks) != 2 {
		t.Fatalf("Expected 2 constraint checks, got %d", len(collector.constraintChecks))
	}

	constraintCheck2 := collector.constraintChecks[1]
	if constraintCheck2.Passed {
		t.Error("Expected Passed to be false for failed constraint")
	}
}

func TestSimpleValidationCollectorGetMetrics(t *testing.T) {
	collector := NewCollector()

	// Record some metrics
	startTime := time.Now()
	collector.RecordValidation("validate", "User", "insert", startTime, nil, nil)
	collector.RecordCacheAccess("constraint", "get", true, 10*time.Millisecond, nil)
	collector.RecordConstraintCheck("required", "email", true, 1*time.Millisecond, nil)

	// Get metrics
	metrics, err := collector.GetMetrics()
	if err != nil {
		t.Errorf("GetMetrics should not return error: %v", err)
	}

	// In the current implementation, GetMetrics returns empty slice
	// This is expected as it's a simplified implementation
	if metrics == nil {
		t.Error("GetMetrics should not return nil")
	}
}

func TestNewValidationMonitor(t *testing.T) {
	collector := NewCollector()
	monitor := NewValidationMonitor(collector)

	if monitor == nil {
		t.Fatal("Expected validation monitor to be created")
	}

	if monitor.collector != collector {
		t.Error("Expected collector to be set")
	}

	if !monitor.enabled {
		t.Error("Expected monitor to be enabled when collector is provided")
	}
}

func TestNewValidationMonitorNilCollector(t *testing.T) {
	monitor := NewValidationMonitor(nil)

	if monitor == nil {
		t.Fatal("Expected validation monitor to be created even with nil collector")
	}

	if monitor.collector != nil {
		t.Error("Expected collector to be nil")
	}

	if monitor.enabled {
		t.Error("Expected monitor to be disabled when collector is nil")
	}
}

func TestValidationMonitorRecordValidation(t *testing.T) {
	collector := NewCollector()
	monitor := NewValidationMonitor(collector)

	// Test successful validation
	monitor.RecordValidation("validate", "User", "insert", 100*time.Millisecond, nil, nil)

	// Test with disabled monitor
	disabledMonitor := NewValidationMonitor(nil)
	disabledMonitor.RecordValidation("validate", "User", "insert", 100*time.Millisecond, nil, nil)

	// Should not panic when disabled
}

func TestValidationMonitorRecordCacheAccess(t *testing.T) {
	collector := NewCollector()
	monitor := NewValidationMonitor(collector)

	// Test cache access
	monitor.RecordCacheAccess("constraint", "get", true, 10*time.Millisecond, nil)

	// Test with disabled monitor
	disabledMonitor := NewValidationMonitor(nil)
	disabledMonitor.RecordCacheAccess("constraint", "get", true, 10*time.Millisecond, nil)

	// Should not panic when disabled
}

func TestValidationMonitorRecordConstraintCheck(t *testing.T) {
	collector := NewCollector().(*SimpleValidationCollector)
	monitor := NewValidationMonitor(collector)

	// Test constraint check
	monitor.RecordConstraintCheck("required", "email", true, 1*time.Millisecond, nil)

	if len(collector.constraintChecks) != 1 {
		t.Fatalf("Expected 1 constraint check, got %d", len(collector.constraintChecks))
	}

	// Test with disabled monitor
	disabledMonitor := NewValidationMonitor(nil)
	disabledMonitor.RecordConstraintCheck("required", "email", true, 1*time.Millisecond, nil)

	// Should not record anything when disabled
	if len(collector.constraintChecks) != 1 {
		t.Errorf("Disabled monitor should not record constraint checks")
	}
}

func TestValidationMonitorIntegration(t *testing.T) {
	// Test that the monitor integrates with the monitoring package
	collector := NewCollector()
	monitor := NewValidationMonitor(collector)

	// Test all monitor methods

	monitor.RecordValidation("test", "Model", "scenario", 100*time.Millisecond, nil, map[string]string{"test": "label"})
	monitor.RecordCacheAccess("type", "op", true, 100*time.Millisecond, map[string]string{"test": "label"})
	monitor.RecordConstraintCheck("type", "field", true, 50*time.Millisecond, map[string]string{"test": "label"})

	// Verify no panic
	if monitor == nil {
		t.Error("Monitor should still be valid")
	}
}
