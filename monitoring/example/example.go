// Package main provides a simple example of using MagicORM monitoring collectors.
package main

import (
	"fmt"
	"time"

	"github.com/muidea/magicOrm/monitoring"
	"github.com/muidea/magicOrm/monitoring/database"
	"github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/validation"
)

func main() {
	fmt.Println("=== MagicORM Monitoring Collectors Example ===")

	// Example 1: Simple ORM Collector
	fmt.Println("\n1. Simple ORM Collector:")
	simpleORMExample()

	// Example 2: Validation Collector
	fmt.Println("\n2. Validation Collector:")
	validationExample()

	// Example 3: Database Collector
	fmt.Println("\n3. Database Collector:")
	databaseExample()

	fmt.Println("\n=== Example Complete ===")
}

func simpleORMExample() {
	// Create a simple ORM collector
	collector := orm.NewCollector()

	// Simulate ORM operations
	startTime := time.Now()

	// Record an insert operation
	time.Sleep(50 * time.Millisecond)
	collector.RecordOperation(
		monitoring.OperationInsert,
		"User",
		startTime,
		nil,
		map[string]string{"source": "example"},
	)

	// Record a query operation
	startTime = time.Now()
	time.Sleep(30 * time.Millisecond)
	collector.RecordQuery(
		"Product",
		monitoring.QueryTypeSimple,
		10,
		startTime,
		nil,
		map[string]string{"source": "example"},
	)

	// Record a transaction
	startTime = time.Now()
	time.Sleep(20 * time.Millisecond)
	collector.RecordTransaction(
		"begin",
		startTime,
		nil,
		map[string]string{"source": "example"},
	)

	fmt.Println("  Recorded 3 ORM operations")
}

func validationExample() {
	// Create a simple validation collector
	collector := validation.NewCollector()

	// Simulate validation operations
	startTime := time.Now()

	// Record a validation operation
	time.Sleep(10 * time.Millisecond)
	collector.RecordValidation(
		"validate_model",
		"User",
		"insert",
		startTime,
		nil,
		map[string]string{"source": "example"},
	)

	// Record a constraint check
	collector.RecordConstraintCheck(
		"required",
		"email",
		true,
		5*time.Millisecond,
		map[string]string{"source": "example"},
	)

	fmt.Println("  Recorded validation and constraint check")
}

func databaseExample() {
	// Create a simple database collector
	collector := database.NewCollector()

	// Simulate database operations
	startTime := time.Now()

	// Record a query
	time.Sleep(100 * time.Millisecond)
	collector.RecordQuery(
		"postgresql",
		"select",
		25,
		startTime,
		nil,
		map[string]string{"source": "example"},
	)

	// Record connection pool stats
	collector.RecordConnectionPool(
		"postgresql",
		5,  // active connections
		10, // idle connections
		2,  // waiting connections
		20, // max connections
		map[string]string{"source": "example"},
	)

	fmt.Println("  Recorded database query and connection stats")
}
