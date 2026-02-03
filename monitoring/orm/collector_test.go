package orm

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/monitoring"
	"github.com/muidea/magicOrm/monitoring/database"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("Expected collector to be created")
	}

	// Verify it's the right type
	_, ok := collector.(*SimpleORMCollector)
	if !ok {
		t.Error("Expected SimpleORMCollector type")
	}
}

func TestSimpleORMCollectorRecordOperation(t *testing.T) {
	collector := NewCollector().(*SimpleORMCollector)
	startTime := time.Now()

	// Test successful operation
	collector.RecordOperation(monitoring.OperationInsert, "User", startTime, nil, map[string]string{"test": "label"})

	if len(collector.operations) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(collector.operations))
	}

	op := collector.operations[0]
	if op.operation != monitoring.OperationInsert {
		t.Errorf("Expected operation 'insert', got '%s'", op.operation)
	}

	if op.modelName != "User" {
		t.Errorf("Expected modelName 'User', got '%s'", op.modelName)
	}

	if op.err != nil {
		t.Errorf("Expected no error, got %v", op.err)
	}

	if op.labels["test"] != "label" {
		t.Errorf("Expected label 'label', got '%s'", op.labels["test"])
	}

	// Test operation with error
	testErr := &database.QueryError{Message: "test error"}
	collector.RecordOperation(monitoring.OperationUpdate, "Product", startTime, testErr, nil)

	if len(collector.operations) != 2 {
		t.Fatalf("Expected 2 operations, got %d", len(collector.operations))
	}

	op2 := collector.operations[1]
	if op2.operation != monitoring.OperationUpdate {
		t.Errorf("Expected operation 'update', got '%s'", op2.operation)
	}

	if op2.err == nil {
		t.Error("Expected error for failed operation")
	}
}

func TestSimpleORMCollectorRecordQuery(t *testing.T) {
	collector := NewCollector().(*SimpleORMCollector)
	startTime := time.Now()

	// Test successful query
	collector.RecordQuery("User", monitoring.QueryTypeSimple, 10, startTime, nil, map[string]string{"test": "label"})

	if len(collector.queries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(collector.queries))
	}

	query := collector.queries[0]
	if query.modelName != "User" {
		t.Errorf("Expected modelName 'User', got '%s'", query.modelName)
	}

	if query.queryType != monitoring.QueryTypeSimple {
		t.Errorf("Expected queryType 'simple', got '%s'", query.queryType)
	}

	if query.rowsReturned != 10 {
		t.Errorf("Expected rowsReturned 10, got %d", query.rowsReturned)
	}

	if query.err != nil {
		t.Errorf("Expected no error, got %v", query.err)
	}

	if query.labels["test"] != "label" {
		t.Errorf("Expected label 'label', got '%s'", query.labels["test"])
	}

	// Test query with error
	testErr := &database.QueryError{Message: "query failed"}
	collector.RecordQuery("Product", monitoring.QueryTypeFilter, 0, startTime, testErr, nil)

	if len(collector.queries) != 2 {
		t.Fatalf("Expected 2 queries, got %d", len(collector.queries))
	}

	query2 := collector.queries[1]
	if query2.modelName != "Product" {
		t.Errorf("Expected modelName 'Product', got '%s'", query2.modelName)
	}

	if query2.queryType != monitoring.QueryTypeFilter {
		t.Errorf("Expected queryType 'filter', got '%s'", query2.queryType)
	}

	if query2.err == nil {
		t.Error("Expected error for failed query")
	}
}

func TestSimpleORMCollectorRecordTransaction(t *testing.T) {
	collector := NewCollector().(*SimpleORMCollector)
	startTime := time.Now()

	// Test successful transaction
	collector.RecordTransaction("BEGIN", startTime, nil, map[string]string{"test": "label"})

	if len(collector.transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(collector.transactions))
	}

	transaction := collector.transactions[0]
	if transaction.operation != "BEGIN" {
		t.Errorf("Expected operation 'BEGIN', got '%s'", transaction.operation)
	}

	if transaction.err != nil {
		t.Errorf("Expected no error, got %v", transaction.err)
	}

	if transaction.labels["test"] != "label" {
		t.Errorf("Expected label 'label', got '%s'", transaction.labels["test"])
	}

	// Test transaction with error
	testErr := &database.TransactionError{Message: "transaction failed"}
	collector.RecordTransaction("COMMIT", startTime, testErr, nil)

	if len(collector.transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(collector.transactions))
	}

	transaction2 := collector.transactions[1]
	if transaction2.operation != "COMMIT" {
		t.Errorf("Expected operation 'COMMIT', got '%s'", transaction2.operation)
	}

	if transaction2.err == nil {
		t.Error("Expected error for failed transaction")
	}
}

func TestSimpleORMCollectorRecordCacheAccess(t *testing.T) {
	collector := NewCollector().(*SimpleORMCollector)

	// This method is currently a no-op in the simplified implementation
	// Just test that it doesn't panic
	collector.RecordCacheAccess("query", "get", true, 10*time.Millisecond, nil)
	collector.RecordCacheAccess("model", "set", false, 5*time.Millisecond, map[string]string{"test": "label"})

	// If we get here without panic, the test passes
}

func TestSimpleORMCollectorRecordDatabaseOperation(t *testing.T) {
	collector := NewCollector().(*SimpleORMCollector)
	startTime := time.Now()

	// This method is currently a no-op in the simplified implementation
	// Just test that it doesn't panic
	collector.RecordDatabaseOperation("postgresql", "execute", startTime, nil, nil)
	collector.RecordDatabaseOperation("mysql", "query", startTime, &database.QueryError{Message: "failed"}, map[string]string{"test": "label"})

	// If we get here without panic, the test passes
}

func TestSimpleORMCollectorGetMetrics(t *testing.T) {
	collector := NewCollector()

	// Record some metrics
	startTime := time.Now()
	collector.RecordOperation(monitoring.OperationInsert, "User", startTime, nil, nil)
	collector.RecordQuery("User", monitoring.QueryTypeSimple, 5, startTime, nil, nil)
	collector.RecordTransaction("BEGIN", startTime, nil, nil)

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

func TestNewMonitoredOrm(t *testing.T) {
	// Create a mock ORM
	mockOrm := &mockOrm{}
	collector := NewCollector()

	monitoredOrm := NewMonitoredOrm(mockOrm, collector)

	if monitoredOrm == nil {
		t.Fatal("Expected monitored ORM to be created")
	}

	if monitoredOrm.orm != mockOrm {
		t.Error("Expected ORM to be set")
	}

	if monitoredOrm.collector != collector {
		t.Error("Expected collector to be set")
	}

	if !monitoredOrm.enabled {
		t.Error("Expected monitor to be enabled when collector is provided")
	}
}

func TestNewMonitoredOrmNilCollector(t *testing.T) {
	mockOrm := &mockOrm{}
	monitoredOrm := NewMonitoredOrm(mockOrm, nil)

	if monitoredOrm == nil {
		t.Fatal("Expected monitored ORM to be created even with nil collector")
	}

	if monitoredOrm.collector != nil {
		t.Error("Expected collector to be nil")
	}

	if monitoredOrm.enabled {
		t.Error("Expected monitor to be disabled when collector is nil")
	}
}

// Mock implementation for testing
type mockOrm struct{}

func (m *mockOrm) Create(entity models.Model) *cd.Error                        { return nil }
func (m *mockOrm) Drop(entity models.Model) *cd.Error                          { return nil }
func (m *mockOrm) Insert(entity models.Model) (models.Model, *cd.Error)        { return entity, nil }
func (m *mockOrm) Update(entity models.Model) (models.Model, *cd.Error)        { return entity, nil }
func (m *mockOrm) Delete(entity models.Model) (models.Model, *cd.Error)        { return entity, nil }
func (m *mockOrm) Query(entity models.Model) (models.Model, *cd.Error)         { return entity, nil }
func (m *mockOrm) Count(filter models.Filter) (int64, *cd.Error)               { return 0, nil }
func (m *mockOrm) BatchQuery(filter models.Filter) ([]models.Model, *cd.Error) { return nil, nil }
func (m *mockOrm) BeginTransaction() *cd.Error                                 { return nil }
func (m *mockOrm) CommitTransaction() *cd.Error                                { return nil }
func (m *mockOrm) RollbackTransaction() *cd.Error                              { return nil }
func (m *mockOrm) Release()                                                    {}
