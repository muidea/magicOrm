package validation_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
	"github.com/muidea/magicOrm/validation/errors"
)

// TestSimpleValidation tests basic validation functionality
func TestSimpleValidation(t *testing.T) {
	config := validation.DefaultConfig()
	manager := validation.NewValidationManager(config)

	// Test 1: Basic type validation
	t.Run("TypeValidation", func(t *testing.T) {
		ctx := validation.NewContext(
			errors.ScenarioInsert,
			validation.OperationCreate,
			nil,
			"postgresql",
		)

		testCases := []struct {
			name  string
			value any
			valid bool
		}{
			{"String", "test", true},
			{"Integer", 42, true},
			{"Float", 3.14, true},
			{"Boolean", true, true},
			{"Time", time.Now(), true},
			{"Nil", nil, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := manager.Validate(tc.value, ctx)
				if tc.valid && err != nil {
					t.Errorf("%s should be valid but got error: %v", tc.name, err)
				}
			})
		}
	})

	// Test 2: Error collection
	t.Run("ErrorCollection", func(t *testing.T) {
		collector := errors.NewErrorCollector()
		_ = validation.NewContextWithCollector(errors.ScenarioInsert, collector)

		// Add some test errors
		collector.AddError(errors.NewValidationError("Test error 1"))
		collector.AddError(errors.NewValidationError("Test error 2"))

		if !collector.HasErrors() {
			t.Error("Collector should have errors")
		}

		if len(collector.GetErrors()) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(collector.GetErrors()))
		}
	})
}

// TestFieldValidation tests field-level validation
func TestFieldValidation(t *testing.T) {
	config := validation.DefaultConfig()
	manager := validation.NewValidationManager(config)

	// Create a mock field adapter
	field := validation.NewFieldAdapter(
		"testField",
		reflect.TypeOf(""),
		&mockConstraints{
			directives: []models.Directive{
				&mockDirective{key: models.KeyRequired},
				&mockDirective{key: models.KeyMin, args: []string{"3"}},
			},
		},
		nil,
	)

	ctx := validation.NewContext(
		errors.ScenarioInsert,
		validation.OperationCreate,
		nil,
		"",
	)

	testCases := []struct {
		name     string
		value    any
		expected bool
	}{
		{"Valid", "valid", true},
		{"TooShort", "ab", false},
		{"Empty", "", false},
		{"Nil", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create field context
			fieldCtx := ctx
			fieldCtx.Field = field

			err := manager.Validate(tc.value, fieldCtx)
			hasError := err != nil

			if hasError != !tc.expected {
				t.Errorf("%s: expected valid=%v, got error=%v (err: %v)",
					tc.name, tc.expected, hasError, err)
			}
		})
	}
}

// TestScenarioValidation tests scenario-aware validation
func TestScenarioValidation(t *testing.T) {
	scenarios := []errors.Scenario{
		errors.ScenarioInsert,
		errors.ScenarioUpdate,
		errors.ScenarioQuery,
		errors.ScenarioDelete,
	}

	for _, scenario := range scenarios {
		t.Run(string(scenario), func(t *testing.T) {
			ctx := validation.NewContext(
				scenario,
				validation.OperationCreate,
				nil,
				"",
			)

			// Just test that context creation works
			if ctx.Scenario != scenario {
				t.Errorf("Expected scenario %s, got %s", scenario, ctx.Scenario)
			}
		})
	}
}

// Mock implementations

type mockConstraints struct {
	directives []models.Directive
}

func (c *mockConstraints) Has(key models.Key) bool {
	for _, d := range c.directives {
		if d.Key() == key {
			return true
		}
	}
	return false
}

func (c *mockConstraints) Get(key models.Key) (models.Directive, bool) {
	for _, d := range c.directives {
		if d.Key() == key {
			return d, true
		}
	}
	return nil, false
}

func (c *mockConstraints) Directives() []models.Directive {
	return c.directives
}

type mockDirective struct {
	key  models.Key
	args []string
}

func (d *mockDirective) Key() models.Key {
	return d.key
}

func (d *mockDirective) Args() []string {
	return d.args
}

func (d *mockDirective) HasArgs() bool {
	return len(d.args) > 0
}
