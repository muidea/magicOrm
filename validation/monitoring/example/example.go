package main

import (
	"fmt"
	"time"

	"github.com/muidea/magicOrm/validation/monitoring"
)

func main() {
	fmt.Println("=== MagicORM Validation Monitoring Example ===")

	// 1. Create metrics collector with default config
	config := monitoring.DefaultMonitoringConfig()
	metrics := monitoring.NewMetricsCollector(config)

	// 2. Create logger
	logger := monitoring.NewValidationLogger(metrics, config)

	// 3. Record some validation operations
	fmt.Println("\nRecording validation operations...")

	// Successful validation
	logger.RecordValidation(
		"validate_user",
		"User",
		"insert",
		50*time.Millisecond,
		nil,
		map[string]interface{}{
			"field_count": 5,
			"constraints": 3,
		},
	)

	// Failed validation
	logger.RecordValidation(
		"validate_product",
		"Product",
		"update",
		30*time.Millisecond,
		fmt.Errorf("price must be positive"),
		map[string]interface{}{
			"field_count": 8,
		},
	)

	// 4. Create and start exporter (optional)
	exportConfig := monitoring.ExportConfig{
		Enabled: true,
		Port:    9090,
	}

	exporter := monitoring.NewMetricsExporter(metrics, logger, exportConfig)
	if err := exporter.Start(); err != nil {
		fmt.Printf("Failed to start exporter: %v\n", err)
	}
	defer exporter.Stop()

	// 5. Get metrics
	fmt.Println("\nMetrics collected:")
	collectedMetrics := metrics.GetMetrics()
	fmt.Printf("Total metric types: %d\n", len(collectedMetrics))

	// 6. Simulate some operations
	fmt.Println("\nSimulating more operations...")
	for i := 0; i < 5; i++ {
		logger.RecordValidation(
			fmt.Sprintf("validate_%d", i),
			"TestModel",
			"insert",
			time.Duration(20+i*10)*time.Millisecond,
			nil,
			map[string]interface{}{
				"iteration": i,
			},
		)
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Metrics available at: http://localhost:9090/metrics")
}
