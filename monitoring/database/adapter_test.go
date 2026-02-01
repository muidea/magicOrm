package database

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/monitoring/core"
)

// MockExecutor is a mock database executor for testing
type MockExecutor struct {
	queryCalled         bool
	executeCalled       bool
	executeInsertCalled bool
	beginCalled         bool
	commitCalled        bool
	rollbackCalled      bool
	releaseCalled       bool
	nextCalled          bool
	finishCalled        bool
	getFieldCalled      bool
	checkTableCalled    bool

	queryResult        []string
	queryError         *cd.Error
	executeResult      int64
	executeError       *cd.Error
	executeInsertError *cd.Error
	beginError         *cd.Error
	commitError        *cd.Error
	rollbackError      *cd.Error
	getFieldError      *cd.Error
	checkTableResult   bool
	checkTableError    *cd.Error
}

func (m *MockExecutor) Release() {
	m.releaseCalled = true
}

func (m *MockExecutor) BeginTransaction() *cd.Error {
	m.beginCalled = true
	return m.beginError
}

func (m *MockExecutor) CommitTransaction() *cd.Error {
	m.commitCalled = true
	return m.commitError
}

func (m *MockExecutor) RollbackTransaction() *cd.Error {
	m.rollbackCalled = true
	return m.rollbackError
}

func (m *MockExecutor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
	m.queryCalled = true
	return m.queryResult, m.queryError
}

func (m *MockExecutor) Next() bool {
	m.nextCalled = true
	return false
}

func (m *MockExecutor) Finish() {
	m.finishCalled = true
}

func (m *MockExecutor) GetField(value ...any) *cd.Error {
	m.getFieldCalled = true
	return m.getFieldError
}

func (m *MockExecutor) Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error) {
	m.executeCalled = true
	return m.executeResult, m.executeError
}

func (m *MockExecutor) ExecuteInsert(sql string, pkValOut any, args ...any) (err *cd.Error) {
	m.executeInsertCalled = true
	return m.executeInsertError
}

func (m *MockExecutor) CheckTableExist(tableName string) (bool, *cd.Error) {
	m.checkTableCalled = true
	return m.checkTableResult, m.checkTableError
}

func TestNewDatabaseMonitor(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := core.NewCollector(&config)

	monitor := NewDatabaseMonitor(collector, &config)
	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	if monitor.collector != collector {
		t.Error("Expected collector to be set")
	}

	if monitor.config != &config {
		t.Error("Expected config to be set")
	}
}

func TestDatabaseMonitorRecordConnection(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0 // Sample all operations
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	// Test successful connection
	monitor.RecordConnection("postgresql", "connect", true, 100*time.Millisecond)

	metrics, err := collector.GetMetric("database_connections_total")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Value != 1.0 {
		t.Errorf("Expected value 1.0, got %f", metric.Value)
	}

	if metric.Labels["database_type"] != "postgresql" {
		t.Errorf("Expected database_type=postgresql, got %s", metric.Labels["database_type"])
	}

	if metric.Labels["operation"] != "connect" {
		t.Errorf("Expected operation=connect, got %s", metric.Labels["operation"])
	}

	// Test failed connection
	monitor.RecordConnection("mysql", "connect", false, 50*time.Millisecond)

	metrics, err = collector.GetMetric("database_connections_total")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 2 {
		t.Fatalf("Expected 2 metrics, got %d", len(metrics))
	}
}

func TestDatabaseMonitorRecordQuery(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	// Test successful query
	monitor.RecordQuery("postgresql", "select", true, 200*time.Millisecond, 10)

	// Check query count
	metrics, err := collector.GetMetric("database_queries_total")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Value != 1.0 {
		t.Errorf("Expected value 1.0, got %f", metric.Value)
	}

	if metric.Labels["database_type"] != "postgresql" {
		t.Errorf("Expected database_type=postgresql, got %s", metric.Labels["database_type"])
	}

	if metric.Labels["query_type"] != "select" {
		t.Errorf("Expected query_type=select, got %s", metric.Labels["query_type"])
	}

	if metric.Labels["success"] != "true" {
		t.Errorf("Expected success=true, got %s", metric.Labels["success"])
	}

	// Check query duration
	durationMetrics, err := collector.GetMetric("database_query_duration_seconds")
	if err != nil {
		t.Fatalf("Failed to get duration metrics: %v", err)
	}

	if len(durationMetrics) != 1 {
		t.Fatalf("Expected 1 duration metric, got %d", len(durationMetrics))
	}

	// Check rows processed
	rowsMetrics, err := collector.GetMetric("database_query_rows_processed")
	if err != nil {
		t.Fatalf("Failed to get rows metrics: %v", err)
	}

	if len(rowsMetrics) != 1 {
		t.Fatalf("Expected 1 rows metric, got %d", len(rowsMetrics))
	}

	rowsMetric := rowsMetrics[0]
	if rowsMetric.Value != 10.0 {
		t.Errorf("Expected 10 rows, got %f", rowsMetric.Value)
	}
}

func TestDatabaseMonitorRecordTransaction(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	// Test successful transaction
	monitor.RecordTransaction("mysql", "begin", true, 50*time.Millisecond)

	metrics, err := collector.GetMetric("database_transactions_total")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Value != 1.0 {
		t.Errorf("Expected value 1.0, got %f", metric.Value)
	}

	if metric.Labels["database_type"] != "mysql" {
		t.Errorf("Expected database_type=mysql, got %s", metric.Labels["database_type"])
	}

	if metric.Labels["operation"] != "begin" {
		t.Errorf("Expected operation=begin, got %s", metric.Labels["operation"])
	}

	if metric.Labels["success"] != "true" {
		t.Errorf("Expected success=true, got %s", metric.Labels["success"])
	}
}

func TestDatabaseMonitorRecordExecution(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	// Test successful execution
	monitor.RecordExecution("postgresql", "update", true, 150*time.Millisecond, 5)

	// Check execution count
	metrics, err := collector.GetMetric("database_executions_total")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Value != 1.0 {
		t.Errorf("Expected value 1.0, got %f", metric.Value)
	}

	if metric.Labels["database_type"] != "postgresql" {
		t.Errorf("Expected database_type=postgresql, got %s", metric.Labels["database_type"])
	}

	if metric.Labels["operation"] != "update" {
		t.Errorf("Expected operation=update, got %s", metric.Labels["operation"])
	}

	// Check rows affected
	rowsMetrics, err := collector.GetMetric("database_execution_rows_affected")
	if err != nil {
		t.Fatalf("Failed to get rows metrics: %v", err)
	}

	if len(rowsMetrics) != 1 {
		t.Fatalf("Expected 1 rows metric, got %d", len(rowsMetrics))
	}

	rowsMetric := rowsMetrics[0]
	if rowsMetric.Value != 5.0 {
		t.Errorf("Expected 5 rows affected, got %f", rowsMetric.Value)
	}
}

func TestDatabaseMonitorRecordError(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitor.RecordError("mysql", "connection_error", "connect")

	metrics, metricErr := collector.GetMetric("database_errors_total")
	if metricErr != nil {
		t.Fatalf("Failed to get metrics: %v", metricErr)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Value != 1.0 {
		t.Errorf("Expected value 1.0, got %f", metric.Value)
	}

	if metric.Labels["database_type"] != "mysql" {
		t.Errorf("Expected database_type=mysql, got %s", metric.Labels["database_type"])
	}

	if metric.Labels["error_type"] != "connection_error" {
		t.Errorf("Expected error_type=connection_error, got %s", metric.Labels["error_type"])
	}

	if metric.Labels["operation"] != "connect" {
		t.Errorf("Expected operation=connect, got %s", metric.Labels["operation"])
	}
}

func TestDatabaseMonitorUpdateConnectionPool(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitor.UpdateConnectionPool("postgresql", 10, 5)

	// Check active connections
	activeMetrics, err := collector.GetMetric("database_connections_active")
	if err != nil {
		t.Fatalf("Failed to get active metrics: %v", err)
	}

	if len(activeMetrics) != 1 {
		t.Fatalf("Expected 1 active metric, got %d", len(activeMetrics))
	}

	activeMetric := activeMetrics[0]
	if activeMetric.Value != 10.0 {
		t.Errorf("Expected 10 active connections, got %f", activeMetric.Value)
	}

	// Check idle connections
	idleMetrics, err := collector.GetMetric("database_connections_idle")
	if err != nil {
		t.Fatalf("Failed to get idle metrics: %v", err)
	}

	if len(idleMetrics) != 1 {
		t.Fatalf("Expected 1 idle metric, got %d", len(idleMetrics))
	}

	idleMetric := idleMetrics[0]
	if idleMetric.Value != 5.0 {
		t.Errorf("Expected 5 idle connections, got %f", idleMetric.Value)
	}
}

func TestDatabaseMonitorSampling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 0.0 // Sample nothing
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	// This should not be recorded due to sampling
	monitor.RecordConnection("postgresql", "connect", true, 100*time.Millisecond)

	metrics, err := collector.GetMetric("database_connections_total")
	if err == nil {
		t.Fatal("Expected error when getting metrics (none should be recorded)")
	}

	if metrics != nil {
		t.Errorf("Expected no metrics, got %d", len(metrics))
	}
}

func TestNewMonitoredExecutor(t *testing.T) {
	mockExecutor := &MockExecutor{}
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitoredExecutor := NewMonitoredExecutor(mockExecutor, monitor, "postgresql")
	if monitoredExecutor == nil {
		t.Fatal("Expected monitored executor to be created")
	}

	if monitoredExecutor.Executor != mockExecutor {
		t.Error("Expected executor to be wrapped")
	}

	if monitoredExecutor.monitor != monitor {
		t.Error("Expected monitor to be set")
	}

	if monitoredExecutor.dbType != "postgresql" {
		t.Errorf("Expected dbType=postgresql, got %s", monitoredExecutor.dbType)
	}
}

func TestMonitoredExecutorQuery(t *testing.T) {
	mockExecutor := &MockExecutor{
		queryResult: []string{"id", "name"},
		queryError:  nil,
	}

	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitoredExecutor := NewMonitoredExecutor(mockExecutor, monitor, "mysql")

	result, err := monitoredExecutor.Query("SELECT * FROM users", true)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if !mockExecutor.queryCalled {
		t.Error("Expected Query to be called on mock executor")
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result))
	}

	// Check metrics were recorded
	metrics, metricErr := collector.GetMetric("database_queries_total")
	if metricErr != nil {
		t.Fatalf("Failed to get metrics: %v", metricErr)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Labels["database_type"] != "mysql" {
		t.Errorf("Expected database_type=mysql, got %s", metric.Labels["database_type"])
	}

	if metric.Labels["query_type"] != "query" {
		t.Errorf("Expected query_type=query, got %s", metric.Labels["query_type"])
	}
}

func TestMonitoredExecutorQueryError(t *testing.T) {
	mockExecutor := &MockExecutor{
		queryError: cd.NewError(cd.Unexpected, "query failed"),
	}

	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitoredExecutor := NewMonitoredExecutor(mockExecutor, monitor, "postgresql")

	_, err := monitoredExecutor.Query("SELECT * FROM users", true)
	if err == nil {
		t.Fatal("Expected query to fail")
	}

	if !mockExecutor.queryCalled {
		t.Error("Expected Query to be called on mock executor")
	}

	// Check error was recorded
	errorMetrics, metricErr := collector.GetMetric("database_errors_total")
	if metricErr != nil {
		t.Fatalf("Failed to get error metrics: %v", metricErr)
	}

	if len(errorMetrics) != 1 {
		t.Fatalf("Expected 1 error metric, got %d", len(errorMetrics))
	}

	errorMetric := errorMetrics[0]
	if errorMetric.Labels["error_type"] != "Unexpected" {
		t.Errorf("Expected error_type=Unexpected, got %s", errorMetric.Labels["error_type"])
	}
}

func TestMonitoredExecutorExecute(t *testing.T) {
	mockExecutor := &MockExecutor{
		executeResult: 5,
		executeError:  nil,
	}

	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitoredExecutor := NewMonitoredExecutor(mockExecutor, monitor, "mysql")

	rowsAffected, err := monitoredExecutor.Execute("UPDATE users SET name = ? WHERE id = ?", "John", 1)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !mockExecutor.executeCalled {
		t.Error("Expected Execute to be called on mock executor")
	}

	if rowsAffected != 5 {
		t.Errorf("Expected 5 rows affected, got %d", rowsAffected)
	}

	// Check metrics were recorded
	metrics, metricErr := collector.GetMetric("database_executions_total")
	if metricErr != nil {
		t.Fatalf("Failed to get metrics: %v", metricErr)
	}

	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %d", len(metrics))
	}
}

func TestMonitoredExecutorTransaction(t *testing.T) {
	mockExecutor := &MockExecutor{
		beginError:    nil,
		commitError:   nil,
		rollbackError: nil,
	}

	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0
	collector := core.NewCollector(&config)
	monitor := NewDatabaseMonitor(collector, &config)

	monitoredExecutor := NewMonitoredExecutor(mockExecutor, monitor, "postgresql")

	// Test begin transaction
	err := monitoredExecutor.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction failed: %v", err)
	}

	if !mockExecutor.beginCalled {
		t.Error("Expected BeginTransaction to be called")
	}

	// Test commit transaction
	err = monitoredExecutor.CommitTransaction()
	if err != nil {
		t.Fatalf("CommitTransaction failed: %v", err)
	}

	if !mockExecutor.commitCalled {
		t.Error("Expected CommitTransaction to be called")
	}

	// Test rollback transaction
	err = monitoredExecutor.RollbackTransaction()
	if err != nil {
		t.Fatalf("RollbackTransaction failed: %v", err)
	}

	if !mockExecutor.rollbackCalled {
		t.Error("Expected RollbackTransaction to be called")
	}

	// Check transaction metrics
	metrics, metricErr := collector.GetMetric("database_transactions_total")
	if metricErr != nil {
		t.Fatalf("Failed to get metrics: %v", metricErr)
	}

	if len(metrics) != 3 {
		t.Fatalf("Expected 3 transaction metrics, got %d", len(metrics))
	}
}

func TestDatabaseMonitoringFactory(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0

	factory := NewSimpleDatabaseMonitoringFactory(&config)
	if factory == nil {
		t.Fatal("Expected factory to be created")
	}

	monitor := factory.GetMonitor()
	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	// Test wrapping executor
	mockExecutor := &MockExecutor{}
	monitoredExecutor := factory.WrapExecutor(mockExecutor, "mysql")
	if monitoredExecutor == nil {
		t.Fatal("Expected monitored executor to be created")
	}

	if monitoredExecutor.Executor != mockExecutor {
		t.Error("Expected executor to be wrapped")
	}
}

func TestDatabaseMonitorShouldSample(t *testing.T) {
	// Test with nil config (should always sample)
	monitor := &DatabaseMonitor{config: nil}
	if !monitor.shouldSample() {
		t.Error("Expected shouldSample to return true with nil config")
	}

	// Test with sampling rate 1.0
	config := &core.MonitoringConfig{SamplingRate: 1.0}
	monitor.config = config
	if !monitor.shouldSample() {
		t.Error("Expected shouldSample to return true with sampling rate 1.0")
	}

	// Test with sampling rate 0.0
	config.SamplingRate = 0.0
	if monitor.shouldSample() {
		t.Error("Expected shouldSample to return false with sampling rate 0.0")
	}
}
