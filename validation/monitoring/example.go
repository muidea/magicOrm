package monitoring

import (
	"fmt"
	"os"
	"time"
)

// Example demonstrates how to use the monitoring package
func Example() {
	fmt.Println("=== MagicORM Validation Monitoring Example ===")

	// 1. Create metrics collector
	metrics := NewMetricsCollector()

	// 2. Create logger with metrics
	logger := NewValidationLogger("info", nil).WithMetrics(metrics)

	// 3. Add some custom fields to logger
	logger.WithFields(map[string]interface{}{
		"service": "validation",
		"version": "1.0.0",
		"env":     "production",
	})

	// 4. Simulate some validation operations
	fmt.Println("\nSimulating validation operations...")

	// Validation 1: Successful validation
	logger.LogValidation(
		"insert",
		"User",
		"insert",
		50*time.Millisecond,
		nil,
		map[string]interface{}{
			"user_id": 123,
			"fields":  []string{"name", "email"},
		},
	)

	// Validation 2: Failed validation
	logger.LogValidation(
		"update",
		"Product",
		"update",
		30*time.Millisecond,
		fmt.Errorf("constraint validation failed: price must be positive"),
		map[string]interface{}{
			"product_id": 456,
			"field":      "price",
			"value":      -10.0,
		},
	)

	// 5. Simulate cache operations
	fmt.Println("\nSimulating cache operations...")

	logger.LogCacheAccess(
		"get",
		"constraint",
		"price|min:0|max:1000",
		true, // hit
		2*time.Millisecond,
		map[string]interface{}{
			"value_type": "float64",
		},
	)

	logger.LogCacheAccess(
		"get",
		"model",
		"User|insert",
		false, // miss
		5*time.Millisecond,
	)

	// 6. Simulate layer performance logging
	fmt.Println("\nSimulating layer performance logging...")

	logger.LogLayerPerformance(
		"type",
		"validate_string",
		10*time.Millisecond,
		true,
		map[string]interface{}{
			"value": "test@example.com",
		},
	)

	logger.LogLayerPerformance(
		"constraint",
		"validate_email",
		15*time.Millisecond,
		false,
		map[string]interface{}{
			"constraint": "email",
			"value":      "invalid-email",
		},
	)

	// 7. Get and display metrics
	fmt.Println("\n=== Current Metrics ===")

	metricsData, err := logger.GetMetrics()
	if err != nil {
		fmt.Printf("Error getting metrics: %v\n", err)
		return
	}

	fmt.Printf("Total Validations: %d\n", metricsData.TotalValidations)
	fmt.Printf("Validation Rate: %.2f validations/sec\n", metricsData.ValidationRate)
	fmt.Printf("Average Validation Time: %v\n", metricsData.AverageValidationTime)
	fmt.Printf("Total Errors: %d (Error Rate: %.2f%%)\n",
		metricsData.TotalErrors, metricsData.ErrorRate*100)
	fmt.Printf("Cache Hits: %d, Misses: %d (Hit Rate: %.2f%%)\n",
		metricsData.CacheHits, metricsData.CacheMisses, metricsData.CacheHitRate*100)
	fmt.Printf("Current Concurrent Validations: %d\n", metricsData.CurrentConcurrentValidations)
	fmt.Printf("Peak Concurrent Validations: %d\n", metricsData.PeakConcurrentValidations)
	fmt.Printf("Memory Usage: %d bytes\n", metricsData.MemoryUsage)
	fmt.Printf("Uptime: %v\n", metricsData.Uptime)

	// 8. Display errors by type
	fmt.Println("\n=== Errors by Type ===")
	for errorType, count := range metricsData.ErrorsByType {
		fmt.Printf("  %s: %d\n", errorType, count)
	}

	// 9. Display layer performance
	fmt.Println("\n=== Layer Performance ===")
	for layer, avgTime := range metricsData.LayerAverages {
		count := metricsData.LayerCounts[layer]
		fmt.Printf("  %s: %v (count: %d)\n", layer, avgTime, count)
	}

	// 10. Start metrics exporter (optional)
	fmt.Println("\n=== Starting Metrics Exporter ===")

	exporterConfig := ExportConfig{
		Enabled:          true,
		Port:             9091, // Different port for example
		Path:             "/metrics",
		HealthCheckPath:  "/health",
		MetricsPath:      "/metrics/json",
		EnablePrometheus: true,
		EnableJSON:       true,
		RefreshInterval:  30 * time.Second,
		Timeout:          10 * time.Second,
	}

	exporter := NewMetricsExporter(metrics, logger, exporterConfig)

	// Add custom labels
	exporter.WithLabels(map[string]string{
		"application": "magicorm",
		"component":   "validation",
		"instance":    "example-1",
	})

	// Start exporter in background
	go func() {
		if err := exporter.Start(); err != nil {
			fmt.Printf("Failed to start exporter: %v\n", err)
		}
	}()

	fmt.Println("Metrics exporter started on port 9091")
	fmt.Println("  - Prometheus metrics: http://localhost:9091/metrics")
	fmt.Println("  - JSON metrics: http://localhost:9091/metrics/json")
	fmt.Println("  - Health check: http://localhost:9091/health")

	// 11. Simulate more operations over time
	fmt.Println("\n=== Simulating more operations (waiting 2 seconds)... ===")

	time.Sleep(2 * time.Second)

	// Perform more validations
	for i := 0; i < 5; i++ {
		logger.LogValidation(
			"query",
			"Order",
			"query",
			time.Duration(20+i*5)*time.Millisecond,
			nil,
			map[string]interface{}{
				"order_id": 1000 + i,
			},
		)

		time.Sleep(100 * time.Millisecond)
	}

	// 12. Display updated metrics
	fmt.Println("\n=== Updated Metrics ===")

	metricsData, _ = logger.GetMetrics()
	fmt.Printf("Total Validations: %d\n", metricsData.TotalValidations)
	fmt.Printf("Validation Rate: %.2f validations/sec\n", metricsData.ValidationRate)
	fmt.Printf("Cache Hit Rate: %.2f%%\n", metricsData.CacheHitRate*100)

	// 13. Reset metrics
	fmt.Println("\n=== Resetting Metrics ===")

	if err := logger.ResetMetrics(); err != nil {
		fmt.Printf("Error resetting metrics: %v\n", err)
	} else {
		fmt.Println("Metrics reset successfully")

		// Verify reset
		metricsData, _ = logger.GetMetrics()
		fmt.Printf("Total Validations after reset: %d\n", metricsData.TotalValidations)
		fmt.Printf("Total Errors after reset: %d\n", metricsData.TotalErrors)
	}

	// 14. Stop exporter (in real application, you'd call this on shutdown)
	fmt.Println("\n=== Stopping Metrics Exporter ===")

	if err := exporter.Stop(); err != nil {
		fmt.Printf("Error stopping exporter: %v\n", err)
	} else {
		fmt.Println("Metrics exporter stopped successfully")
	}

	fmt.Println("\n=== Example Complete ===")
}

// FileLoggerExample demonstrates file-based logging
func FileLoggerExample() {
	fmt.Println("=== File Logger Example ===")

	// Create file logger
	fileLogger, err := FileLogger("debug", "validation.log")
	if err != nil {
		fmt.Printf("Error creating file logger: %v\n", err)
		return
	}

	// Add metrics
	metrics := NewMetricsCollector()
	fileLogger.WithMetrics(metrics)

	// Log some operations
	fileLogger.Info("Starting validation system")
	fileLogger.WithField("batch_id", "batch-123")

	fileLogger.LogValidation(
		"batch_insert",
		"UserBatch",
		"insert",
		150*time.Millisecond,
		nil,
		map[string]interface{}{
			"batch_size": 100,
			"successful": 98,
			"failed":     2,
		},
	)

	fileLogger.Error("Batch processing error",
		fmt.Errorf("database connection timeout"),
		map[string]interface{}{
			"retry_count": 3,
			"timeout":     "30s",
		},
	)

	fmt.Println("Log entries written to validation.log")

	// Export metrics to file
	go ExportMetricsToFile(metrics, fileLogger, "metrics.json", 30*time.Second)

	fmt.Println("Metrics will be exported to metrics.json every 30 seconds")
}

// MultiLoggerExample demonstrates multi-output logging
func MultiLoggerExample() {
	fmt.Println("=== Multi Logger Example ===")

	// Create console and file writers
	consoleWriter := os.Stdout
	fileWriter, err := os.OpenFile("multi.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error creating log file: %v\n", err)
		return
	}
	defer fileWriter.Close()

	// Create multi logger
	multiLogger := MultiLogger("info", consoleWriter, fileWriter)

	// Log to both outputs
	multiLogger.Info("This message goes to console and file")
	multiLogger.WithField("multi_output", true)

	multiLogger.LogValidation(
		"multi_test",
		"TestModel",
		"update",
		25*time.Millisecond,
		nil,
	)

	fmt.Println("Check console output and multi.log file")
}

// IntegrationExample shows integration with validation system
func IntegrationExample() {
	fmt.Println("=== Integration with Validation System ===")

	// This would typically be integrated with the actual validation manager
	// For demonstration, we show the pattern

	/*
		// In your validation manager initialization:
		metrics := monitoring.NewMetricsCollector()
		logger := monitoring.NewValidationLogger("info", nil).WithMetrics(metrics)

		config := validation.ValidationConfig{
			EnableMetrics: true,
			Logger:        logger, // Pass logger to validation config
		}

		manager := validation.NewValidationManager(config)

		// Start metrics exporter
		exporter, err := monitoring.StartDefaultExporter(metrics, logger)
		if err != nil {
			logger.Error("Failed to start metrics exporter", err)
		}

		// Use manager for validations
		ctx := validation.NewContext(...)
		err := manager.ValidateModel(model, ctx)

		if err != nil {
			logger.Error("Validation failed", err, map[string]interface{}{
				"model":    model.GetName(),
				"scenario": ctx.Scenario,
			})
		}

		// On shutdown
		exporter.Stop()
	*/

	fmt.Println("Integration pattern shown above")
	fmt.Println("See PRODUCTION_GUIDE.md for complete integration instructions")
}
