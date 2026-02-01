package validation_test

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
	"github.com/muidea/magicOrm/validation/errors"
)

// TestIntegration tests the complete validation system
func TestIntegration(t *testing.T) {
	// Create validation manager with all layers enabled
	config := validation.DefaultConfig()
	config.EnableTypeValidation = true
	config.EnableConstraintValidation = true
	config.EnableDatabaseValidation = true
	config.EnableScenarioAdaptation = true

	manager := validation.NewValidationManager(config)

	// Test 1: Complete validation flow
	t.Run("CompleteValidationFlow", func(t *testing.T) {
		// Create a field with constraints
		field := validation.NewFieldAdapter(
			"username",
			reflect.TypeOf(""),
			&mockConstraints{
				directives: []models.Directive{
					&mockDirective{key: models.KeyRequired},
					&mockDirective{key: models.KeyMin, args: []string{"3"}},
					&mockDirective{key: models.KeyMax, args: []string{"20"}},
				},
			},
			"testuser",
		)

		// Create context for insert scenario
		ctx := validation.NewContext(
			errors.ScenarioInsert,
			validation.OperationCreate,
			nil,
			"postgresql",
		)
		ctx.Field = field

		// Validate
		err := manager.Validate("testuser", ctx)
		if err != nil {
			t.Errorf("Valid value should pass: %v", err)
		}
	})

	// Test 2: Error handling with all layers
	t.Run("MultiLayerErrorHandling", func(t *testing.T) {
		collector := errors.NewErrorCollector()
		ctx := validation.NewContextWithCollector(errors.ScenarioInsert, collector)

		// Create a field that will fail multiple validations
		field := validation.NewFieldAdapter(
			"age",
			reflect.TypeOf(0),
			&mockConstraints{
				directives: []models.Directive{
					&mockDirective{key: models.KeyRequired},
					&mockDirective{key: models.KeyMin, args: []string{"18"}},
				},
			},
			nil, // This will fail required constraint
		)
		ctx.Field = field

		// This should generate errors
		_ = manager.Validate(nil, ctx)

		if !collector.HasErrors() {
			t.Error("Should have collected validation errors")
		}

		// Check that we have error details
		errorList := collector.GetErrors()
		if len(errorList) == 0 {
			t.Error("Should have at least one error")
		}
	})

	// Test 3: Scenario-specific validation
	t.Run("ScenarioSpecificValidation", func(t *testing.T) {
		// Test insert scenario (strict)
		field := validation.NewFieldAdapter(
			"readonly_field",
			reflect.TypeOf(""),
			&mockConstraints{
				directives: []models.Directive{
					&mockDirective{key: models.KeyReadOnly},
					&mockDirective{key: models.KeyRequired},
				},
			},
			"value",
		)

		// Insert scenario should validate read-only fields
		insertCtx := validation.NewContext(
			errors.ScenarioInsert,
			validation.OperationCreate,
			nil,
			"",
		)
		insertCtx.Field = field

		err := manager.Validate("value", insertCtx)
		if err != nil {
			t.Errorf("Insert should validate read-only fields: %v", err)
		}

		// Update scenario should skip read-only fields
		updateCtx := validation.NewContext(
			errors.ScenarioUpdate,
			validation.OperationUpdate,
			nil,
			"",
		)
		updateCtx.Field = field
		updateCtx.Options.ValidateReadOnlyFields = false

		err = manager.Validate("", updateCtx)
		if err != nil {
			t.Errorf("Update should skip read-only fields: %v", err)
		}
	})

	// Test 4: Database validation
	t.Run("DatabaseValidation", func(t *testing.T) {
		field := validation.NewFieldAdapter(
			"required_field",
			reflect.TypeOf(""),
			&mockConstraints{
				directives: []models.Directive{
					&mockDirective{key: models.KeyRequired},
				},
			},
			nil,
		)

		ctx := validation.NewContext(
			errors.ScenarioInsert,
			validation.OperationCreate,
			nil,
			"postgresql",
		)
		ctx.Field = field

		// This should fail database NOT NULL constraint
		err := manager.Validate(nil, ctx)
		if err == nil {
			t.Error("Database validation should fail for null required field")
		}
	})
}

// TestPerformance tests validation performance
func TestPerformance(t *testing.T) {
	config := validation.DefaultConfig()
	config.EnableCaching = true
	manager := validation.NewValidationManager(config)

	// Create a field with constraints
	field := validation.NewFieldAdapter(
		"test_field",
		reflect.TypeOf(""),
		&mockConstraints{
			directives: []models.Directive{
				&mockDirective{key: models.KeyRequired},
				&mockDirective{key: models.KeyMin, args: []string{"1"}},
				&mockDirective{key: models.KeyMax, args: []string{"100"}},
			},
		},
		"test",
	)

	ctx := validation.NewContext(
		errors.ScenarioInsert,
		validation.OperationCreate,
		nil,
		"",
	)
	ctx.Field = field

	// Run multiple validations to test caching
	for i := 0; i < 10; i++ {
		err := manager.Validate("test", ctx)
		if err != nil {
			t.Errorf("Validation failed on iteration %d: %v", i, err)
		}
	}

	// Get stats
	stats := manager.GetValidationStats()
	if stats.TotalValidations < 10 {
		t.Errorf("Expected at least 10 validations, got %d", stats.TotalValidations)
	}
}

// TestCustomization tests custom constraint registration
func TestCustomization(t *testing.T) {
	config := validation.DefaultConfig()
	manager := validation.NewValidationManager(config)

	// Register a custom constraint
	customCalled := false
	err := manager.RegisterCustomConstraint("custom", func(value any, args []string) error {
		customCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Failed to register custom constraint: %v", err)
	}

	// Test with custom constraint
	field := validation.NewFieldAdapter(
		"custom_field",
		reflect.TypeOf(""),
		&mockConstraints{
			directives: []models.Directive{
				&mockDirective{key: models.Key("custom")},
			},
		},
		"test",
	)

	ctx := validation.NewContext(
		errors.ScenarioInsert,
		validation.OperationCreate,
		nil,
		"",
	)
	ctx.Field = field

	err = manager.Validate("test", ctx)
	if err != nil {
		t.Errorf("Custom constraint validation failed: %v", err)
	}

	if !customCalled {
		t.Error("Custom constraint handler was not called")
	}
}
