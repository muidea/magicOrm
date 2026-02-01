package main

import (
	"fmt"
	"time"

	"github.com/muidea/magicOrm/validation"
)

func ConfigurationExample() {
	fmt.Println("=== Validation Configuration Examples ===")

	// Example 1: Default configuration
	fmt.Println("\n1. Default Configuration:")
	defaultConfig := validation.DefaultConfig()
	fmt.Printf("   - Type Validation: %v\n", defaultConfig.EnableTypeValidation)
	fmt.Printf("   - Constraint Validation: %v\n", defaultConfig.EnableConstraintValidation)
	fmt.Printf("   - Database Validation: %v\n", defaultConfig.EnableDatabaseValidation)
	fmt.Printf("   - Scenario Adaptation: %v\n", defaultConfig.EnableScenarioAdaptation)
	fmt.Printf("   - Caching: %v\n", defaultConfig.EnableCaching)
	fmt.Printf("   - Cache TTL: %v\n", defaultConfig.CacheTTL)
	fmt.Printf("   - Stop on First Error: %v\n", defaultConfig.DefaultOptions.StopOnFirstError)
	fmt.Printf("   - Validate Read-Only Fields: %v\n", defaultConfig.DefaultOptions.ValidateReadOnlyFields)
	fmt.Printf("   - Validate Write-Only Fields: %v\n", defaultConfig.DefaultOptions.ValidateWriteOnlyFields)

	// Example 2: Simple configuration for basic validation
	fmt.Println("\n2. Simple Configuration (Basic Validation):")
	simpleConfig := validation.SimpleConfig()
	fmt.Printf("   - Type Validation: %v\n", simpleConfig.EnableTypeValidation)
	fmt.Printf("   - Constraint Validation: %v\n", simpleConfig.EnableConstraintValidation)
	fmt.Printf("   - Database Validation: %v\n", simpleConfig.EnableDatabaseValidation)
	fmt.Printf("   - Scenario Adaptation: %v\n", simpleConfig.EnableScenarioAdaptation)
	fmt.Printf("   - Caching: %v\n", simpleConfig.EnableCaching)
	fmt.Printf("   - Stop on First Error: %v\n", simpleConfig.DefaultOptions.StopOnFirstError)

	// Example 3: Performance-optimized configuration
	fmt.Println("\n3. Performance-Optimized Configuration:")
	perfConfig := validation.ValidationConfig{
		EnableTypeValidation:       true,
		EnableConstraintValidation: true,
		EnableDatabaseValidation:   false, // Skip database validation for performance
		EnableScenarioAdaptation:   true,
		EnableCaching:              true,
		CacheTTL:                   10 * time.Minute, // Longer TTL
		MaxCacheSize:               2000,             // Larger cache
		DefaultOptions: validation.ValidationOptions{
			StopOnFirstError:        true, // Stop early for performance
			IncludeFieldPathInError: false,
			ValidateReadOnlyFields:  true,
			ValidateWriteOnlyFields: true,
		},
	}
	fmt.Printf("   - Database Validation: %v (disabled for performance)\n", perfConfig.EnableDatabaseValidation)
	fmt.Printf("   - Cache TTL: %v (longer for better hit rate)\n", perfConfig.CacheTTL)
	fmt.Printf("   - Max Cache Size: %d (larger cache)\n", perfConfig.MaxCacheSize)
	fmt.Printf("   - Stop on First Error: %v (early termination)\n", perfConfig.DefaultOptions.StopOnFirstError)

	// Example 4: Strict validation configuration
	fmt.Println("\n4. Strict Validation Configuration:")
	strictConfig := validation.ValidationConfig{
		EnableTypeValidation:       true,
		EnableConstraintValidation: true,
		EnableDatabaseValidation:   true,
		EnableScenarioAdaptation:   true,
		EnableCaching:              false, // No caching for strict validation
		DefaultOptions: validation.ValidationOptions{
			StopOnFirstError:        false, // Collect all errors
			IncludeFieldPathInError: true,
			ValidateReadOnlyFields:  true,
			ValidateWriteOnlyFields: true,
		},
	}
	fmt.Printf("   - Caching: %v (disabled for strict validation)\n", strictConfig.EnableCaching)
	fmt.Printf("   - Stop on First Error: %v (collect all errors)\n", strictConfig.DefaultOptions.StopOnFirstError)
	fmt.Printf("   - Include Field Path: %v (detailed error messages)\n", strictConfig.DefaultOptions.IncludeFieldPathInError)

	// Example 5: Development configuration
	fmt.Println("\n5. Development/Testing Configuration:")
	devConfig := validation.ValidationConfig{
		EnableTypeValidation:       true,
		EnableConstraintValidation: true,
		EnableDatabaseValidation:   true,
		EnableScenarioAdaptation:   true,
		EnableCaching:              false, // Disable cache to see fresh errors
		DefaultOptions: validation.ValidationOptions{
			StopOnFirstError:        false, // See all errors during development
			IncludeFieldPathInError: true,  // Detailed error messages
			ValidateReadOnlyFields:  true,
			ValidateWriteOnlyFields: true,
		},
	}
	fmt.Printf("   - Caching: %v (disabled for development)\n", devConfig.EnableCaching)
	fmt.Printf("   - Stop on First Error: %v (see all errors)\n", devConfig.DefaultOptions.StopOnFirstError)

	// Example 6: Scenario-specific configuration
	fmt.Println("\n6. Scenario-Specific Configuration Examples:")

	// Insert scenario: strict validation
	insertConfig := validation.DefaultConfig()
	insertConfig.DefaultOptions.ValidateReadOnlyFields = true  // Validate read-only on insert
	insertConfig.DefaultOptions.ValidateWriteOnlyFields = true // Validate write-only on insert
	fmt.Println("   - Insert Scenario:")
	fmt.Printf("     * Validate Read-Only: %v\n", insertConfig.DefaultOptions.ValidateReadOnlyFields)
	fmt.Printf("     * Validate Write-Only: %v\n", insertConfig.DefaultOptions.ValidateWriteOnlyFields)

	// Update scenario: relaxed validation
	updateConfig := validation.DefaultConfig()
	updateConfig.DefaultOptions.ValidateReadOnlyFields = false // Skip read-only validation
	updateConfig.DefaultOptions.ValidateWriteOnlyFields = true // Still validate write-only
	fmt.Println("   - Update Scenario:")
	fmt.Printf("     * Validate Read-Only: %v (relaxed)\n", updateConfig.DefaultOptions.ValidateReadOnlyFields)
	fmt.Printf("     * Validate Write-Only: %v\n", updateConfig.DefaultOptions.ValidateWriteOnlyFields)

	// Query scenario: minimal validation
	queryConfig := validation.DefaultConfig()
	queryConfig.DefaultOptions.ValidateReadOnlyFields = true   // Can read read-only fields
	queryConfig.DefaultOptions.ValidateWriteOnlyFields = false // Skip write-only validation
	fmt.Println("   - Query Scenario:")
	fmt.Printf("     * Validate Read-Only: %v\n", queryConfig.DefaultOptions.ValidateReadOnlyFields)
	fmt.Printf("     * Validate Write-Only: %v (skip)\n", queryConfig.DefaultOptions.ValidateWriteOnlyFields)

	// Example 7: Custom configuration based on environment
	fmt.Println("\n7. Environment-Based Configuration:")

	// Production environment
	prodConfig := createConfigForEnvironment("production")
	fmt.Println("   - Production Environment:")
	fmt.Printf("     * Caching: %v\n", prodConfig.EnableCaching)
	fmt.Printf("     * Cache TTL: %v\n", prodConfig.CacheTTL)
	fmt.Printf("     * Max Cache Size: %d\n", prodConfig.MaxCacheSize)

	// Staging environment
	stageConfig := createConfigForEnvironment("staging")
	fmt.Println("   - Staging Environment:")
	fmt.Printf("     * Caching: %v\n", stageConfig.EnableCaching)
	fmt.Printf("     * Stop on First Error: %v\n", stageConfig.DefaultOptions.StopOnFirstError)

	// Development environment
	devEnvConfig := createConfigForEnvironment("development")
	fmt.Println("   - Development Environment:")
	fmt.Printf("     * Caching: %v\n", devEnvConfig.EnableCaching)
	fmt.Printf("     * Include Field Path: %v\n", devEnvConfig.DefaultOptions.IncludeFieldPathInError)

	fmt.Println("\n=== Configuration Usage Tips ===")
	fmt.Println("1. Use DefaultConfig() for most use cases")
	fmt.Println("2. Use SimpleConfig() for basic validation needs")
	fmt.Println("3. Adjust caching based on expected load and memory constraints")
	fmt.Println("4. Configure scenario-specific options based on operation type")
	fmt.Println("5. Disable caching during development for easier debugging")
	fmt.Println("6. Enable all error collection during testing")
	fmt.Println("7. Consider database validation overhead in performance-critical paths")
}

// Helper function to create configuration based on environment
func createConfigForEnvironment(env string) validation.ValidationConfig {
	switch env {
	case "production":
		return validation.ValidationConfig{
			EnableTypeValidation:       true,
			EnableConstraintValidation: true,
			EnableDatabaseValidation:   true,
			EnableScenarioAdaptation:   true,
			EnableCaching:              true,
			CacheTTL:                   10 * time.Minute,
			MaxCacheSize:               5000,
			DefaultOptions: validation.ValidationOptions{
				StopOnFirstError:        true, // Fail fast in production
				IncludeFieldPathInError: false,
				ValidateReadOnlyFields:  true,
				ValidateWriteOnlyFields: true,
			},
		}
	case "staging":
		return validation.ValidationConfig{
			EnableTypeValidation:       true,
			EnableConstraintValidation: true,
			EnableDatabaseValidation:   true,
			EnableScenarioAdaptation:   true,
			EnableCaching:              true,
			CacheTTL:                   5 * time.Minute,
			MaxCacheSize:               1000,
			DefaultOptions: validation.ValidationOptions{
				StopOnFirstError:        false, // Collect all errors in staging
				IncludeFieldPathInError: true,
				ValidateReadOnlyFields:  true,
				ValidateWriteOnlyFields: true,
			},
		}
	default: // development
		return validation.ValidationConfig{
			EnableTypeValidation:       true,
			EnableConstraintValidation: true,
			EnableDatabaseValidation:   true,
			EnableScenarioAdaptation:   true,
			EnableCaching:              false, // No cache for development
			DefaultOptions: validation.ValidationOptions{
				StopOnFirstError:        false, // See all errors
				IncludeFieldPathInError: true,  // Detailed errors
				ValidateReadOnlyFields:  true,
				ValidateWriteOnlyFields: true,
			},
		}
	}
}
