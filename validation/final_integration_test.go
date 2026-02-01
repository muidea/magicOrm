package validation

import (
	"testing"

	"github.com/muidea/magicOrm/validation/errors"
)

// TestFinalIntegration tests the final integration of all validation components
func TestFinalIntegration(t *testing.T) {
	t.Run("FactoryCreation", func(t *testing.T) {
		factory := NewValidationFactory()
		if factory == nil {
			t.Fatal("Validation factory is nil")
		}

		// Test default config
		defaultConfig := factory.GetDefaultConfig()
		if !defaultConfig.EnableTypeValidation {
			t.Error("Default config should have type validation enabled")
		}

		// Test simple config
		simpleConfig := factory.GetSimpleConfig()
		if simpleConfig.EnableDatabaseValidation {
			t.Error("Simple config should not have database validation enabled")
		}
	})

	t.Run("ManagerCreation", func(t *testing.T) {
		factory := NewValidationFactory()
		config := DefaultConfig()

		manager := factory.CreateValidationManager(config)
		if manager == nil {
			t.Fatal("Validation manager is nil")
		}

		// Test layer management
		err := manager.DisableLayer(LayerType)
		if err != nil {
			t.Errorf("Failed to disable type layer: %s", err.Error())
		}

		err = manager.EnableLayer(LayerType)
		if err != nil {
			t.Errorf("Failed to enable type layer: %s", err.Error())
		}

		// Test scenario setting
		manager.SetScenario(errors.ScenarioUpdate)
	})

	t.Run("CacheIntegration", func(t *testing.T) {
		// Test with caching enabled
		config := DefaultConfig()
		config.EnableCaching = true
		config.CacheTTL = 60 // 60 seconds

		factory := NewValidationFactory()
		manager := factory.CreateValidationManager(config)

		// Get initial stats
		initialStats := manager.GetValidationStats()

		// Perform some validations
		ctx := NewContext(
			errors.ScenarioInsert,
			OperationCreate,
			nil,
			"postgresql",
		)

		// Validate multiple times
		for i := 0; i < 5; i++ {
			err := manager.Validate("test value", ctx)
			if err != nil {
				t.Errorf("Validation failed: %s", err.Error())
			}
		}

		// Get final stats
		finalStats := manager.GetValidationStats()

		// Check that validations increased
		if finalStats.TotalValidations <= initialStats.TotalValidations {
			t.Error("Total validations should have increased")
		}

		t.Logf("Initial stats: %+v", initialStats)
		t.Logf("Final stats: %+v", finalStats)
	})

	t.Run("Customization", func(t *testing.T) {
		factory := NewValidationFactory()

		// Test custom constraint registration
		customErr := factory.RegisterCustomConstraint("custom", func(val any, args []string) error {
			// Always succeed for this test
			return nil
		})

		if customErr != nil {
			t.Errorf("Failed to register custom constraint: %s", customErr.Error())
		}

		// Create manager with custom constraint
		manager := factory.CreateValidationManager(DefaultConfig())

		// Test that manager can be used
		ctx := NewContext(
			errors.ScenarioInsert,
			OperationCreate,
			nil,
			"postgresql",
		)

		err := manager.Validate("test", ctx)
		if err != nil {
			t.Errorf("Validation with custom constraint failed: %s", err.Error())
		}
	})
}

// TestProviderCompatibility tests that validation system is compatible with providers
func TestProviderCompatibility(t *testing.T) {
	t.Run("ScenarioSupport", func(t *testing.T) {
		// Test all scenarios are supported
		scenarios := []errors.Scenario{
			errors.ScenarioInsert,
			errors.ScenarioUpdate,
			errors.ScenarioQuery,
			errors.ScenarioDelete,
		}

		factory := NewValidationFactory()
		manager := factory.CreateValidationManager(DefaultConfig())

		for _, scenario := range scenarios {
			t.Run(string(scenario), func(t *testing.T) {
				manager.SetScenario(scenario)

				// Create context with scenario
				ctx := NewContext(
					scenario,
					getOperationTypeForScenario(scenario),
					nil,
					"postgresql",
				)

				// Simple validation test
				err := manager.Validate("test", ctx)
				if err != nil {
					t.Errorf("Validation failed for scenario %s: %s", scenario, err.Error())
				}
			})
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		factory := NewValidationFactory()
		manager := factory.CreateValidationManager(DefaultConfig())

		// Test error collector
		collector := errors.NewErrorCollector()
		ctx := NewContextWithCollector(
			errors.ScenarioInsert,
			collector,
		)

		// Validate with error collector
		err := manager.Validate("test", ctx)
		if err != nil {
			t.Errorf("Validation failed: %s", err.Error())
		}

		// Error collector should be available in context
		if ctx.Collector == nil {
			t.Error("Error collector should be set in context")
		}
	})

	t.Run("Performance", func(t *testing.T) {
		// Test that validation is performant
		factory := NewValidationFactory()

		// Test with caching
		cacheConfig := DefaultConfig()
		cacheConfig.EnableCaching = true
		cacheManager := factory.CreateValidationManager(cacheConfig)

		ctx := NewContext(
			errors.ScenarioInsert,
			OperationCreate,
			nil,
			"postgresql",
		)

		// Run multiple validations
		const iterations = 100
		for i := 0; i < iterations; i++ {
			err := cacheManager.Validate("performance test", ctx)
			if err != nil {
				t.Errorf("Validation failed at iteration %d: %s", i, err.Error())
				break
			}
		}

		stats := cacheManager.GetValidationStats()
		t.Logf("Performance test stats: %+v", stats)

		// Check cache efficiency
		if stats.CacheHits+stats.CacheMisses > 0 {
			hitRate := float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses)
			t.Logf("Cache hit rate: %.2f%%", hitRate*100)
		}
	})
}

// Helper function to get operation type for scenario
func getOperationTypeForScenario(scenario errors.Scenario) OperationType {
	switch scenario {
	case errors.ScenarioInsert:
		return OperationCreate
	case errors.ScenarioUpdate:
		return OperationUpdate
	case errors.ScenarioQuery:
		return OperationRead
	case errors.ScenarioDelete:
		return OperationDelete
	default:
		return OperationCreate
	}
}
