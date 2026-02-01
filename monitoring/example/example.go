package main

import (
	"fmt"
	"time"

	"github.com/muidea/magicOrm/monitoring/core"
	"github.com/muidea/magicOrm/monitoring/database"
	"github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/unified"
	"github.com/muidea/magicOrm/monitoring/validation"
)

func main() {
	fmt.Println("=== MagicORM Monitoring System Example ===")

	// Example 1: Quick Start with Default Configuration
	fmt.Println("\n1. Quick Start with Default Configuration:")
	quickStartExample()

	// Example 2: Unified Monitoring Manager
	fmt.Println("\n2. Unified Monitoring Manager:")
	unifiedManagerExample()

	// Example 3: Database Monitoring
	fmt.Println("\n3. Database Monitoring:")
	databaseMonitoringExample()

	// Example 4: ORM Monitoring
	fmt.Println("\n4. ORM Monitoring:")
	ormMonitoringExample()

	// Example 5: Validation Monitoring
	fmt.Println("\n5. Validation Monitoring:")
	validationMonitoringExample()

	// Example 6: Custom Configuration
	fmt.Println("\n6. Custom Configuration:")
	customConfigurationExample()

	fmt.Println("\n=== Example Complete ===")
}

func quickStartExample() {
	// Create a default monitoring manager
	manager := unified.DefaultMonitoringManager()

	// Start monitoring
	if err := manager.Start(); err != nil {
		fmt.Printf("Failed to start monitoring: %v\n", err)
		return
	}
	defer manager.Stop()

	// Record some metrics
	collector := manager.GetCollector()
	if collector != nil {
		// Register a metric
		collector.RegisterDefinition(core.MetricDefinition{
			Name:       "example_operations_total",
			Type:       core.CounterMetric,
			Help:       "Total number of example operations",
			LabelNames: []string{"operation_type", "status"},
		})

		// Record some operations
		collector.Increment("example_operations_total", map[string]string{
			"operation_type": "query",
			"status":         "success",
		})

		collector.Increment("example_operations_total", map[string]string{
			"operation_type": "insert",
			"status":         "error",
		})

		// Get metrics
		metrics := manager.GetMetrics()
		fmt.Printf("Collected %d metric types\n", len(metrics))
	}
}

func unifiedManagerExample() {
	// Create manager with production configuration
	manager := unified.ProductionMonitoringManager()

	// Enable monitoring
	manager.Enable()

	// Get statistics
	stats := manager.GetStats()
	fmt.Printf("Manager uptime: %v\n", stats.Uptime)
	fmt.Printf("Metrics collected: %d\n", stats.MetricsCollected)

	// Get comprehensive stats
	comprehensiveStats := manager.GetManagerStats()
	fmt.Printf("Comprehensive stats available: %v\n", len(comprehensiveStats) > 0)

	// Add custom labels for all metrics
	manager.AddCustomLabel("environment", "production")
	manager.AddCustomLabel("service", "magicorm")

	// Cleanup old metrics
	manager.Cleanup()
}

func databaseMonitoringExample() {
	// Create database monitor
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false // Synchronous for example
	config.SamplingRate = 1.0      // Sample all operations

	collector := core.NewCollector(&config)
	dbMonitor := database.NewDatabaseMonitor(collector, &config)

	// Simulate database operations
	fmt.Println("Simulating database operations...")

	// Record connections
	dbMonitor.RecordConnection("postgresql", "connect", true, 100*time.Millisecond)
	dbMonitor.RecordConnection("mysql", "connect", false, 50*time.Millisecond)

	// Record queries
	dbMonitor.RecordQuery("postgresql", "select", true, 200*time.Millisecond, 10)
	dbMonitor.RecordQuery("mysql", "update", false, 150*time.Millisecond, 0)

	// Record transactions
	dbMonitor.RecordTransaction("postgresql", "begin", true, 50*time.Millisecond)
	dbMonitor.RecordTransaction("postgresql", "commit", true, 30*time.Millisecond)

	// Record executions
	dbMonitor.RecordExecution("mysql", "insert", true, 120*time.Millisecond, 1)
	dbMonitor.RecordExecution("postgresql", "delete", false, 80*time.Millisecond, 0)

	// Record errors
	dbMonitor.RecordError("mysql", "connection_error", "connect")
	dbMonitor.RecordError("postgresql", "timeout", "query")

	// Update connection pool status
	dbMonitor.UpdateConnectionPool("postgresql", 10, 5)
	dbMonitor.UpdateConnectionPool("mysql", 5, 2)

	// Get metrics
	metrics := collector.GetMetrics()
	fmt.Printf("Database metrics collected: %d types\n", len(metrics))

	// Show some metrics
	for name, metricList := range metrics {
		if len(metricList) > 0 {
			fmt.Printf("  %s: %d data points\n", name, len(metricList))
		}
	}
}

func ormMonitoringExample() {
	// Create ORM monitor
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0

	collector := core.NewCollector(&config)
	ormMonitor := orm.NewORMMonitor(collector, &config)

	// Simulate ORM operations
	fmt.Println("Simulating ORM operations...")

	// Record operations using available methods
	operationStart := time.Now().Add(-150 * time.Millisecond)
	ormMonitor.RecordOperation(
		orm.OperationInsert,
		"User",
		operationStart,
		nil,
		map[string]string{
			"entity_type": "User",
			"operation":   "insert",
		},
	)

	// Record query operation
	queryStart := time.Now().Add(-200 * time.Millisecond)
	ormMonitor.RecordQuery(
		"Product",
		orm.QueryTypeSimple,
		10,
		queryStart,
		nil,
		map[string]string{
			"filter": "category=electronics",
		},
	)

	// Record batch operation
	batchStart := time.Now().Add(-500 * time.Millisecond)
	ormMonitor.RecordBatchOperation(
		orm.OperationQuery,
		"Product",
		50,
		batchStart,
		45,
		5,
		nil,
		map[string]string{
			"entity_type": "Product",
		},
	)

	// Record transaction
	transactionStart := time.Now().Add(-300 * time.Millisecond)
	ormMonitor.RecordTransaction(
		"complex_update",
		transactionStart,
		nil,
		map[string]string{
			"description": "complex_update",
		},
	)

	fmt.Println("ORM operations recorded")
}

func validationMonitoringExample() {
	// Create validation monitor
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.SamplingRate = 1.0

	collector := core.NewCollector(&config)
	validationMonitor := validation.NewValidationMonitor(collector, &config)

	// Simulate validation operations
	fmt.Println("Simulating validation operations...")

	// Record validation operations
	validationMonitor.RecordValidation(
		"validate_user",
		"User",
		validation.ScenarioInsert,
		50*time.Millisecond,
		nil,
		map[string]string{
			"field_count": "5",
			"constraints": "3",
		},
	)

	validationMonitor.RecordValidation(
		"validate_product",
		"Product",
		validation.ScenarioUpdate,
		30*time.Millisecond,
		fmt.Errorf("price must be positive"),
		map[string]string{
			"field_count": "8",
		},
	)

	// Record cache operations
	validationMonitor.RecordCacheAccess(
		"validate",
		"type_cache",
		"User",
		true,
		5*time.Millisecond,
		map[string]string{
			"entity": "User",
		},
	)

	validationMonitor.RecordCacheAccess(
		"validate",
		"constraint_cache",
		"Product",
		false,
		10*time.Millisecond,
		map[string]string{
			"entity": "Product",
		},
	)

	// Record layer performance
	validationMonitor.RecordLayerPerformance(
		"type_validation",
		"validate_type",
		20*time.Millisecond,
		true,
		map[string]string{
			"entity": "Order",
		},
	)

	validationMonitor.RecordLayerPerformance(
		"constraint_validation",
		"validate_constraints",
		15*time.Millisecond,
		false,
		map[string]string{
			"entity": "Order",
		},
	)

	// Get validation statistics
	stats := validationMonitor.GetStats()
	fmt.Printf("Validation statistics collected: %v\n", len(stats) > 0)
}

func customConfigurationExample() {
	fmt.Println("Creating custom monitoring configuration...")

	// Create custom configuration
	config := core.MonitoringConfig{
		Enabled:            true,
		SamplingRate:       0.5, // Sample 50% of operations
		DetailLevel:        core.DetailLevelDetailed,
		EnableORM:          true,
		EnableValidation:   true,
		EnableDatabase:     true,
		EnableCache:        true,
		AsyncCollection:    true,
		CollectionInterval: 30 * time.Second,
		RetentionPeriod:    24 * time.Hour,
		BatchSize:          100,
		BufferSize:         1000,
		MaxConcurrentTasks: 10,
		Timeout:            5 * time.Second,
		ExportConfig: core.ExportConfig{
			Enabled:         true,
			Port:            9090,
			Path:            "/metrics",
			HealthCheckPath: "/health",
			InfoPath:        "/info",
			RefreshInterval: 15 * time.Second,
			ScrapeTimeout:   10 * time.Second,
			EnableTLS:       false,
			TLSCertPath:     "",
			TLSKeyPath:      "",
			EnableAuth:      false,
			AuthToken:       "",
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		return
	}

	// Create manager with custom configuration
	manager := unified.NewMonitoringManager(&config)

	// Check configuration
	fmt.Printf("Monitoring enabled: %v\n", manager.IsEnabled())
	fmt.Printf("Sampling rate: %.2f\n", config.SamplingRate)
	fmt.Printf("Export enabled: %v\n", config.IsExportEnabled())

	// Environment-specific configurations
	fmt.Println("\nEnvironment-specific configurations:")
	fmt.Println("  Development:", core.DevelopmentConfig().SamplingRate)
	fmt.Println("  Production:", core.ProductionConfig().SamplingRate)
	fmt.Println("  High Load:", core.HighLoadConfig().SamplingRate)
}

// Helper function to print metrics
func printMetrics(metrics map[string][]core.Metric) {
	for name, metricList := range metrics {
		fmt.Printf("Metric: %s\n", name)
		for i, metric := range metricList {
			fmt.Printf("  [%d] Value: %.2f, Labels: %v\n", i, metric.Value, metric.Labels)
		}
	}
}
