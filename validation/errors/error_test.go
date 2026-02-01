package errors

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
)

func TestValidationError(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("test error")
		if err.Error() != "test error" {
			t.Errorf("Expected error message 'test error', got '%s'", err.Error())
		}
	})

	t.Run("NewTypeError", func(t *testing.T) {
		err := NewTypeError("age", 25, "int")
		expectedMsg := "type mismatch for field 'age': got int, expected int"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
		if err.GetField() != "age" {
			t.Errorf("Expected field 'age', got '%s'", err.GetField())
		}
		if err.GetValue() != 25 {
			t.Errorf("Expected value 25, got %v", err.GetValue())
		}
		if err.GetExpected() != "int" {
			t.Errorf("Expected type 'int', got '%v'", err.GetExpected())
		}
		if err.GetLayer() != LayerType {
			t.Errorf("Expected layer 'type', got '%s'", err.GetLayer())
		}
	})

	t.Run("NewConstraintError", func(t *testing.T) {
		err := NewConstraintError("name", "min", "Jo", 3)
		expectedMsg := "constraint 'min' violation for field 'name': got Jo"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
		if err.GetField() != "name" {
			t.Errorf("Expected field 'name', got '%s'", err.GetField())
		}
		if err.GetConstraint() != "min" {
			t.Errorf("Expected constraint 'min', got '%s'", err.GetConstraint())
		}
		if err.GetValue() != "Jo" {
			t.Errorf("Expected value 'Jo', got '%v'", err.GetValue())
		}
		if err.GetExpected() != 3 {
			t.Errorf("Expected expected value 3, got %v", err.GetExpected())
		}
		if err.GetLayer() != LayerConstraint {
			t.Errorf("Expected layer 'constraint', got '%s'", err.GetLayer())
		}
	})

	t.Run("NewDatabaseError", func(t *testing.T) {
		err := NewDatabaseError("email", "invalid@", "unique")
		expectedMsg := "database constraint 'unique' violation for field 'email': invalid@"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
		if err.GetField() != "email" {
			t.Errorf("Expected field 'email', got '%s'", err.GetField())
		}
		if err.GetConstraint() != "unique" {
			t.Errorf("Expected constraint 'unique', got '%s'", err.GetConstraint())
		}
		if err.GetValue() != "invalid@" {
			t.Errorf("Expected value 'invalid@', got '%v'", err.GetValue())
		}
		if err.GetLayer() != LayerDatabase {
			t.Errorf("Expected layer 'database', got '%s'", err.GetLayer())
		}
	})

	t.Run("WithField", func(t *testing.T) {
		err := NewValidationError("test error").WithField("username")
		if err.GetField() != "username" {
			t.Errorf("Expected field 'username', got '%s'", err.GetField())
		}
	})

	t.Run("WithConstraint", func(t *testing.T) {
		err := NewValidationError("test error").WithConstraint("max")
		if err.GetConstraint() != "max" {
			t.Errorf("Expected constraint 'max', got '%s'", err.GetConstraint())
		}
	})

	t.Run("WithScenario", func(t *testing.T) {
		err := NewValidationError("test error").WithScenario(ScenarioInsert)
		if err.GetScenario() != ScenarioInsert {
			t.Errorf("Expected scenario 'insert', got '%s'", err.GetScenario())
		}
	})

	t.Run("ToRichError", func(t *testing.T) {
		// Test type error
		typeErr := NewTypeError("age", "twenty", "int")
		richErr := typeErr.ToRichError()
		if richErr.Code != cd.IllegalParam {
			t.Errorf("Expected error code %d, got %d", cd.IllegalParam, richErr.Code)
		}

		// Test database error
		dbErr := NewDatabaseError("email", "invalid@", "unique")
		richErr = dbErr.ToRichError()
		if richErr.Code != cd.DatabaseError {
			t.Errorf("Expected error code %d, got %d", cd.DatabaseError, richErr.Code)
		}
	})
}

func TestErrorCollector(t *testing.T) {
	t.Run("NewErrorCollector", func(t *testing.T) {
		collector := NewErrorCollector()
		if collector == nil {
			t.Error("Expected error collector, got nil")
		}
		if collector.HasErrors() {
			t.Error("New collector should not have errors")
		}
	})

	t.Run("AddError", func(t *testing.T) {
		collector := NewErrorCollector()
		err := NewValidationError("test error")
		collector.AddError(err)

		if !collector.HasErrors() {
			t.Error("Collector should have errors after adding one")
		}

		errors := collector.GetErrors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(errors))
		}
		if errors[0].Error() != "test error" {
			t.Errorf("Expected error 'test error', got '%s'", errors[0].Error())
		}
	})

	t.Run("GetErrorsByField", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.AddError(NewValidationError("error 1").WithField("name"))
		collector.AddError(NewValidationError("error 2").WithField("email"))
		collector.AddError(NewValidationError("error 3").WithField("name"))

		nameErrors := collector.GetErrorsByField("name")
		if len(nameErrors) != 2 {
			t.Errorf("Expected 2 errors for field 'name', got %d", len(nameErrors))
		}

		emailErrors := collector.GetErrorsByField("email")
		if len(emailErrors) != 1 {
			t.Errorf("Expected 1 error for field 'email', got %d", len(emailErrors))
		}

		noErrors := collector.GetErrorsByField("nonexistent")
		if len(noErrors) != 0 {
			t.Errorf("Expected 0 errors for field 'nonexistent', got %d", len(noErrors))
		}
	})

	t.Run("GetErrorsByLayer", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.AddError(NewTypeError("age", "twenty", "int"))
		collector.AddError(NewConstraintError("name", "min", "Jo", 3))
		collector.AddError(NewDatabaseError("email", "invalid@", "unique"))
		collector.AddError(NewTypeError("score", "high", "float"))

		typeErrors := collector.GetErrorsByLayer(LayerType)
		if len(typeErrors) != 2 {
			t.Errorf("Expected 2 type errors, got %d", len(typeErrors))
		}

		constraintErrors := collector.GetErrorsByLayer(LayerConstraint)
		if len(constraintErrors) != 1 {
			t.Errorf("Expected 1 constraint error, got %d", len(constraintErrors))
		}

		databaseErrors := collector.GetErrorsByLayer(LayerDatabase)
		if len(databaseErrors) != 1 {
			t.Errorf("Expected 1 database error, got %d", len(databaseErrors))
		}
	})

	t.Run("GetErrorSummary", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.AddError(NewValidationError("first error").WithField("field1"))
		collector.AddError(NewValidationError("second error").WithField("field2"))

		summary := collector.GetErrorSummary()
		if summary == "" {
			t.Error("Expected non-empty error summary")
		}
		t.Logf("Error summary: %s", summary)
	})

	t.Run("Clear", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.AddError(NewValidationError("test error"))

		if !collector.HasErrors() {
			t.Error("Collector should have errors before clear")
		}

		collector.Clear()

		if collector.HasErrors() {
			t.Error("Collector should not have errors after clear")
		}
	})

	t.Run("ToRichError", func(t *testing.T) {
		collector := NewErrorCollector()
		collector.AddError(NewValidationError("first error"))
		collector.AddError(NewValidationError("second error"))

		richErr := collector.ToRichError()
		if richErr == nil {
			t.Error("Expected rich error, got nil")
		}
		if richErr.Code != cd.IllegalParam {
			t.Errorf("Expected error code %d, got %d", cd.IllegalParam, richErr.Code)
		}
	})
}
