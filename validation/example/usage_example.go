package main

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
	"github.com/muidea/magicOrm/validation/errors"
)

// User represents a user model with validation constraints
type User struct {
	ID       int    `orm:"id key auto" constraint:"req"`
	Username string `orm:"username" constraint:"req,min=3,max=20"`
	Email    string `orm:"email" constraint:"req,re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
	Age      int    `orm:"age" constraint:"min=18,max=120"`
	Status   string `orm:"status" constraint:"in=active,inactive,suspended"`
}

func main() {
	fmt.Println("=== MagicORM Validation System - Usage Example ===")

	// Example 1: Basic validation setup
	exampleBasicValidation()

	// Example 2: Scenario-aware validation
	exampleScenarioValidation()

	// Example 3: Error handling and collection
	exampleErrorHandling()

	// Example 4: Custom constraints
	exampleCustomConstraints()

	fmt.Println("\n=== All Examples Completed Successfully ===")
}

func exampleBasicValidation() {
	fmt.Println("\n1. Basic Validation Setup:")

	// Create validation configuration
	config := validation.DefaultConfig()

	// Create validation manager
	manager := validation.NewValidationManager(config)

	// Create a mock field adapter (in real usage, this would come from models)
	usernameField := validation.NewFieldAdapter(
		"username",
		reflect.TypeOf(""),
		createMockConstraints([]string{"req", "min=3", "max=20"}),
		"john_doe",
	)

	// Create validation context
	ctx := validation.NewContext(
		errors.ScenarioInsert,
		validation.OperationCreate,
		nil,
		"postgresql",
	)
	ctx.Field = usernameField

	// Test valid value
	err := manager.Validate("john_doe", ctx)
	if err != nil {
		fmt.Printf("  ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("  ✓ Valid username passed")
	}

	// Test invalid value
	err = manager.Validate("jo", ctx)
	if err != nil {
		fmt.Printf("  ✓ Invalid username correctly rejected: %v\n", err)
	} else {
		fmt.Println("  ❌ Invalid username should have failed")
	}
}

func exampleScenarioValidation() {
	fmt.Println("\n2. Scenario-Aware Validation:")

	config := validation.DefaultConfig()
	manager := validation.NewValidationManager(config)

	// Create a read-only field
	createdAtField := validation.NewFieldAdapter(
		"created_at",
		reflect.TypeOf(""),
		createMockConstraints([]string{"ro", "req"}),
		"2024-01-01",
	)

	// Test insert scenario (should validate read-only)
	insertCtx := validation.NewContext(
		errors.ScenarioInsert,
		validation.OperationCreate,
		nil,
		"",
	)
	insertCtx.Field = createdAtField

	err := manager.Validate("2024-01-01", insertCtx)
	fmt.Printf("  Insert scenario (read-only field): %v\n",
		map[bool]string{true: "✓ validated", false: "❌ error"}[err == nil])

	// Test update scenario (should skip read-only)
	updateCtx := validation.NewContext(
		errors.ScenarioUpdate,
		validation.OperationUpdate,
		nil,
		"",
	)
	updateCtx.Field = createdAtField
	updateCtx.Options.ValidateReadOnlyFields = false

	err = manager.Validate("", updateCtx)
	fmt.Printf("  Update scenario (read-only field): %v\n",
		map[bool]string{true: "✓ skipped", false: "❌ error"}[err == nil])
}

func exampleErrorHandling() {
	fmt.Println("\n3. Error Handling and Collection:")

	config := validation.DefaultConfig()
	manager := validation.NewValidationManager(config)

	// Create error collector
	collector := errors.NewErrorCollector()
	ctx := validation.NewContextWithCollector(errors.ScenarioInsert, collector)

	// Create multiple fields with potential errors
	fields := []struct {
		name        string
		value       any
		constraints []string
	}{
		{"username", "jo", []string{"req", "min=3"}},
		{"email", "invalid", []string{"req", "email"}},
		{"age", 15, []string{"min=18"}},
	}

	// Validate each field and collect errors
	for _, field := range fields {
		fieldAdapter := validation.NewFieldAdapter(
			field.name,
			getType(field.value),
			createMockConstraints(field.constraints),
			field.value,
		)

		fieldCtx := ctx
		fieldCtx.Field = fieldAdapter

		_ = manager.Validate(field.value, fieldCtx)
	}

	// Display collected errors
	if collector.HasErrors() {
		fmt.Printf("  Collected %d error(s):\n", len(collector.GetErrors()))
		for i, err := range collector.GetErrors() {
			fmt.Printf("    %d. %s: %s\n", i+1, err.GetField(), err.Error())
		}

		// Show error summary
		fmt.Printf("  Error summary:\n%s\n", collector.GetErrorSummary())

		// Convert to combined error
		combinedErr := collector.ToRichError()
		fmt.Printf("  Combined error: %v\n", combinedErr)
	}
}

func exampleCustomConstraints() {
	fmt.Println("\n4. Custom Constraints:")

	config := validation.DefaultConfig()
	manager := validation.NewValidationManager(config)

	// Register custom constraint
	err := manager.RegisterCustomConstraint("even", func(value any, args []string) error {
		if num, ok := value.(int); ok {
			if num%2 != 0 {
				return fmt.Errorf("value must be even")
			}
			return nil
		}
		return fmt.Errorf("even constraint only applies to integers")
	})

	if err != nil {
		fmt.Printf("  ❌ Failed to register custom constraint: %v\n", err)
		return
	}

	fmt.Println("  ✓ Custom 'even' constraint registered")

	// Test custom constraint
	evenField := validation.NewFieldAdapter(
		"even_number",
		reflect.TypeOf(0),
		createMockConstraints([]string{"even"}),
		4,
	)

	ctx := validation.NewContext(
		errors.ScenarioInsert,
		validation.OperationCreate,
		nil,
		"",
	)
	ctx.Field = evenField

	// Test even number
	err = manager.Validate(4, ctx)
	fmt.Printf("  Even number (4): %v\n",
		map[bool]string{true: "✓ passed", false: "❌ failed"}[err == nil])

	// Test odd number
	err = manager.Validate(3, ctx)
	fmt.Printf("  Odd number (3): %v\n",
		map[bool]string{true: "❌ should fail", false: "✓ correctly rejected"}[err == nil])
}

// Helper functions

func createMockConstraints(constraintStrs []string) models.Constraints {
	directives := make([]models.Directive, 0, len(constraintStrs))

	for _, c := range constraintStrs {
		// Simple parsing - in real code, use proper constraint parsing
		directives = append(directives, &mockDirective{
			key: models.Key(c),
		})
	}

	return &mockConstraints{
		directives: directives,
	}
}

func getType(value any) reflect.Type {
	if value == nil {
		return nil
	}
	return reflect.TypeOf(value)
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
	key models.Key
}

func (d *mockDirective) Key() models.Key {
	return d.key
}

func (d *mockDirective) Args() []string {
	return []string{}
}

func (d *mockDirective) HasArgs() bool {
	return false
}
