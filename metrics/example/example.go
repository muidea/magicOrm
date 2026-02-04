// Package main provides a simple example of using MagicORM metric providers.
package main

import (
	"fmt"

	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/metrics/database"
	"github.com/muidea/magicOrm/metrics/orm"
	"github.com/muidea/magicOrm/metrics/validation"
)

func main() {
	fmt.Println("=== MagicORM Metric Providers Example ===")
	fmt.Println("Note: MagicORM now only provides metric definitions.")
	fmt.Println("      Data collection is handled by magicCommon/monitoring system.")
	fmt.Println()

	// Example 1: Using ORM Provider
	fmt.Println("1. ORM Metric Provider:")
	ormProviderExample()

	// Example 2: Using Validation Provider
	fmt.Println("\n2. Validation Metric Provider:")
	validationProviderExample()

	// Example 3: Using Database Provider
	fmt.Println("\n3. Database Metric Provider:")
	databaseProviderExample()

	// Example 4: Type Definitions
	fmt.Println("\n4. Type Definitions:")
	typeDefinitionsExample()

	fmt.Println("\n=== Example Complete ===")
}

func ormProviderExample() {
	// Create an ORM metric provider
	provider := orm.NewORMMetricProvider()

	// Get metric definitions
	definitions := provider.Metrics()
	fmt.Printf("  Provider Name: %s\n", provider.Name())
	fmt.Printf("  Metric Definitions: %d\n", len(definitions))

	// List some metric definitions
	for i, def := range definitions {
		if i < 3 { // Show first 3 definitions
			fmt.Printf("    - %s: %s\n", def.Name, def.Help)
		}
	}
	fmt.Println("  ... (more definitions available)")

	// Note: MagicORM no longer collects data
	fmt.Println("  Note: Data collection is handled by magicCommon/monitoring system")
}

func validationProviderExample() {
	// Create a validation metric provider
	provider := validation.NewValidationMetricProvider()

	// Get metric definitions
	definitions := provider.Metrics()
	fmt.Printf("  Provider Name: %s\n", provider.Name())
	fmt.Printf("  Metric Definitions: %d\n", len(definitions))

	// List some metric definitions
	for i, def := range definitions {
		if i < 3 { // Show first 3 definitions
			fmt.Printf("    - %s: %s\n", def.Name, def.Help)
		}
	}
	fmt.Println("  ... (more definitions available)")

	// Note: MagicORM no longer collects data
	fmt.Println("  Note: Data collection is handled by magicCommon/monitoring system")
}

func databaseProviderExample() {
	// Create a database metric provider
	provider := database.NewDatabaseMetricProvider()

	// Get metric definitions
	definitions := provider.Metrics()
	fmt.Printf("  Provider Name: %s\n", provider.Name())
	fmt.Printf("  Metric Definitions: %d\n", len(definitions))

	// List some metric definitions
	for i, def := range definitions {
		if i < 3 { // Show first 3 definitions
			fmt.Printf("    - %s: %s\n", def.Name, def.Help)
		}
	}
	fmt.Println("  ... (more definitions available)")

	// Note: MagicORM no longer collects data
	fmt.Println("  Note: Data collection is handled by magicCommon/monitoring system")
}

func typeDefinitionsExample() {
	// Demonstrate using type definitions from metrics package
	fmt.Println("  Operation Types:")
	fmt.Printf("    Insert: %s\n", metrics.OperationInsert)
	fmt.Printf("    Update: %s\n", metrics.OperationUpdate)
	fmt.Printf("    Query: %s\n", metrics.OperationQuery)
	fmt.Printf("    Delete: %s\n", metrics.OperationDelete)

	fmt.Println("  Query Types:")
	fmt.Printf("    Simple: %s\n", metrics.QueryTypeSimple)
	fmt.Printf("    Filter: %s\n", metrics.QueryTypeFilter)
	fmt.Printf("    Relation: %s\n", metrics.QueryTypeRelation)

	fmt.Println("  Error Types:")
	fmt.Printf("    Database: %s\n", metrics.ErrorTypeDatabase)
	fmt.Printf("    Validation: %s\n", metrics.ErrorTypeValidation)
	fmt.Printf("    Timeout: %s\n", metrics.ErrorTypeTimeout)

	// Demonstrate label utilities
	fmt.Println("  Label Utilities:")
	defaultLabels := metrics.DefaultLabels()
	fmt.Printf("    Default Labels: %v\n", defaultLabels)

	labels1 := map[string]string{"environment": "production"}
	labels2 := map[string]string{"service": "user-api"}
	merged := metrics.MergeLabels(labels1, labels2)
	fmt.Printf("    Merged Labels: %v\n", merged)
}
