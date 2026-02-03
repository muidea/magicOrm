package database

import (
	"testing"
	"time"
)

func TestNewDatabaseMonitor(t *testing.T) {
	collector := NewCollector()
	monitor := NewDatabaseMonitor(collector)

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	if monitor.collector != collector {
		t.Error("Expected collector to be set")
	}

	if !monitor.enabled {
		t.Error("Expected monitor to be enabled when collector is provided")
	}
}

func TestNewDatabaseMonitorNilCollector(t *testing.T) {
	monitor := NewDatabaseMonitor(nil)

	if monitor == nil {
		t.Fatal("Expected monitor to be created even with nil collector")
	}

	if monitor.collector != nil {
		t.Error("Expected collector to be nil")
	}

	if monitor.enabled {
		t.Error("Expected monitor to be disabled when collector is nil")
	}
}

func TestDatabaseMonitorRecordQuery(t *testing.T) {
	collector := NewCollector()
	monitor := NewDatabaseMonitor(collector)

	// Test successful query
	monitor.RecordQuery("postgresql", "SELECT", true, 100*time.Millisecond, 5, nil)

	// Verify query was recorded
	simpleCollector := collector.(*SimpleDatabaseCollector)
	if len(simpleCollector.queries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(simpleCollector.queries))
	}

	query := simpleCollector.queries[0]
	if query.dbType != "postgresql" {
		t.Errorf("Expected dbType 'postgresql', got '%s'", query.dbType)
	}

	if query.queryType != "SELECT" {
		t.Errorf("Expected queryType 'SELECT', got '%s'", query.queryType)
	}

	if query.rowsAffected != 5 {
		t.Errorf("Expected rowsAffected 5, got %d", query.rowsAffected)
	}

	if query.err != nil {
		t.Errorf("Expected no error for successful query, got %v", query.err)
	}

	// Test failed query
	monitor.RecordQuery("mysql", "INSERT", false, 50*time.Millisecond, 0, map[string]string{"custom": "label"})

	if len(simpleCollector.queries) != 2 {
		t.Fatalf("Expected 2 queries, got %d", len(simpleCollector.queries))
	}

	query2 := simpleCollector.queries[1]
	if query2.dbType != "mysql" {
		t.Errorf("Expected dbType 'mysql', got '%s'", query2.dbType)
	}

	if query2.err == nil {
		t.Error("Expected error for failed query")
	}

	if query2.labels["custom"] != "label" {
		t.Errorf("Expected custom label 'label', got '%s'", query2.labels["custom"])
	}

	if query2.labels["rows_affected"] != "0" {
		t.Errorf("Expected rows_affected label '0', got '%s'", query2.labels["rows_affected"])
	}
}

func TestDatabaseMonitorRecordTransaction(t *testing.T) {
	collector := NewCollector()
	monitor := NewDatabaseMonitor(collector)

	// Test successful transaction
	monitor.RecordTransaction("postgresql", "BEGIN", true, 100*time.Millisecond, nil)

	simpleCollector := collector.(*SimpleDatabaseCollector)
	if len(simpleCollector.transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(simpleCollector.transactions))
	}

	transaction := simpleCollector.transactions[0]
	if transaction.dbType != "postgresql" {
		t.Errorf("Expected dbType 'postgresql', got '%s'", transaction.dbType)
	}

	if transaction.operation != "BEGIN" {
		t.Errorf("Expected operation 'BEGIN', got '%s'", transaction.operation)
	}

	if transaction.err != nil {
		t.Errorf("Expected no error for successful transaction, got %v", transaction.err)
	}

	// Test failed transaction
	monitor.RecordTransaction("mysql", "COMMIT", false, 50*time.Millisecond, map[string]string{"custom": "label"})

	if len(simpleCollector.transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(simpleCollector.transactions))
	}

	transaction2 := simpleCollector.transactions[1]
	if transaction2.dbType != "mysql" {
		t.Errorf("Expected dbType 'mysql', got '%s'", transaction2.dbType)
	}

	if transaction2.err == nil {
		t.Error("Expected error for failed transaction")
	}

	if transaction2.labels["custom"] != "label" {
		t.Errorf("Expected custom label 'label', got '%s'", transaction2.labels["custom"])
	}
}

func TestDatabaseMonitorRecordExecution(t *testing.T) {
	collector := NewCollector()
	monitor := NewDatabaseMonitor(collector)

	// Test successful execution
	monitor.RecordExecution("postgresql", "CREATE TABLE", true, 200*time.Millisecond, nil)

	simpleCollector := collector.(*SimpleDatabaseCollector)
	// RecordExecution calls RecordQuery internally
	if len(simpleCollector.queries) != 1 {
		t.Fatalf("Expected 1 query from execution, got %d", len(simpleCollector.queries))
	}

	execution := simpleCollector.queries[0]
	if execution.dbType != "postgresql" {
		t.Errorf("Expected dbType 'postgresql', got '%s'", execution.dbType)
	}

	if execution.queryType != "CREATE TABLE" {
		t.Errorf("Expected queryType 'CREATE TABLE', got '%s'", execution.queryType)
	}

	if execution.err != nil {
		t.Errorf("Expected no error for successful execution, got %v", execution.err)
	}
}

func TestDatabaseMonitorRecordConnection(t *testing.T) {
	collector := NewCollector()
	monitor := NewDatabaseMonitor(collector)

	// Test successful connection
	monitor.RecordConnection("postgresql", "connect", true, 100*time.Millisecond, nil)

	simpleCollector := collector.(*SimpleDatabaseCollector)
	if len(simpleCollector.connections) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(simpleCollector.connections))
	}

	connection := simpleCollector.connections[0]
	if connection.dbType != "postgresql" {
		t.Errorf("Expected dbType 'postgresql', got '%s'", connection.dbType)
	}

	if connection.operation != "connect" {
		t.Errorf("Expected operation 'connect', got '%s'", connection.operation)
	}

	if connection.err != nil {
		t.Errorf("Expected no error for successful connection, got %v", connection.err)
	}
}

func TestDatabaseMonitorRecordConnectionPool(t *testing.T) {
	collector := NewCollector()
	monitor := NewDatabaseMonitor(collector)

	// Test connection pool recording
	monitor.RecordConnectionPool("postgresql", 10, 5, 2, 20, map[string]string{"pool": "main"})

	simpleCollector := collector.(*SimpleDatabaseCollector)
	if len(simpleCollector.connectionPools) != 1 {
		t.Fatalf("Expected 1 connection pool metric, got %d", len(simpleCollector.connectionPools))
	}

	pool := simpleCollector.connectionPools[0]
	if pool.dbType != "postgresql" {
		t.Errorf("Expected dbType 'postgresql', got '%s'", pool.dbType)
	}

	if pool.activeConnections != 10 {
		t.Errorf("Expected activeConnections 10, got %d", pool.activeConnections)
	}

	if pool.idleConnections != 5 {
		t.Errorf("Expected idleConnections 5, got %d", pool.idleConnections)
	}

	if pool.waitingConnections != 2 {
		t.Errorf("Expected waitingConnections 2, got %d", pool.waitingConnections)
	}

	if pool.maxConnections != 20 {
		t.Errorf("Expected maxConnections 20, got %d", pool.maxConnections)
	}

	if pool.labels["pool"] != "main" {
		t.Errorf("Expected pool label 'main', got '%s'", pool.labels["pool"])
	}
}

func TestDatabaseMonitorDisabled(t *testing.T) {
	// Test with nil collector (disabled monitor)
	monitor := NewDatabaseMonitor(nil)

	// These calls should not panic
	monitor.RecordQuery("postgresql", "SELECT", true, 100*time.Millisecond, 5, nil)
	monitor.RecordTransaction("postgresql", "BEGIN", true, 100*time.Millisecond, nil)
	monitor.RecordExecution("postgresql", "CREATE", true, 200*time.Millisecond, nil)
	monitor.RecordConnection("postgresql", "connect", true, 100*time.Millisecond, nil)
	monitor.RecordConnectionPool("postgresql", 10, 5, 2, 20, nil)

	// If we get here without panic, the test passes
}

func TestSimpleDatabaseCollectorGetMetrics(t *testing.T) {
	collector := NewCollector()

	// Record some metrics
	collector.RecordQuery("postgresql", "SELECT", 5, time.Now(), nil, nil)
	collector.RecordTransaction("postgresql", "BEGIN", time.Now(), nil, nil)

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

func TestQueryError(t *testing.T) {
	err := &QueryError{Message: "query failed"}

	expected := "query failed"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestTransactionError(t *testing.T) {
	err := &TransactionError{Message: "transaction failed"}

	expected := "transaction failed"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}
