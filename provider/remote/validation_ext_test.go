package remote

import (
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

func buildRemoteValidationTestObject() *Object {
	return &Object{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*Field{
			{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{PrimaryKey: true, ValueDeclare: models.AutoIncrement},
			},
			{
				Name: "name",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{Constraint: "req"},
			},
		},
	}
}

func TestRemoteValidationExtensionFieldAndModelValidation(t *testing.T) {
	ext := NewValidationExtension(nil)
	object := buildRemoteValidationTestObject()
	nameField := object.GetField("name")

	if ext.GetValidationManager() == nil {
		t.Fatal("GetValidationManager should not return nil")
	}

	if err := ext.ValidateFieldWithScenario(nameField, nil, verrors.ScenarioInsert, false); err == nil {
		t.Fatal("ValidateFieldWithScenario should reject nil for required field")
	}
	if err := ext.ValidateFieldWithScenario(nameField, "apple", verrors.ScenarioInsert, false); err != nil {
		t.Fatalf("ValidateFieldWithScenario should accept valid value, got %v", err)
	}
	if err := ext.ValidateFieldWithScenario(nameField, nil, verrors.ScenarioInsert, true); err != nil {
		t.Fatalf("disableValidator should skip field validation, got %v", err)
	}

	if err := ext.ValidateModelWithScenario(object, verrors.ScenarioInsert, false); err == nil {
		t.Fatal("ValidateModelWithScenario should reject model with missing required field")
	}
	if err := object.SetFieldValue("id", int64(1)); err != nil {
		t.Fatalf("SetFieldValue(id) failed: %v", err)
	}
	if err := object.SetFieldValue("name", "apple"); err != nil {
		t.Fatalf("SetFieldValue(name) failed: %v", err)
	}
	if err := ext.ValidateModelWithScenario(object, verrors.ScenarioInsert, false); err != nil {
		t.Fatalf("ValidateModelWithScenario should accept populated model, got %v", err)
	}
	if err := ext.ValidateModelWithScenario(object, verrors.ScenarioInsert, true); err != nil {
		t.Fatalf("disableValidator should skip model validation, got %v", err)
	}
}

func TestRemoteValidationExtensionSetModelValueWithScenario(t *testing.T) {
	ext := NewValidationExtension(nil)

	invalidObject := buildRemoteValidationTestObject()
	invalidValue := &ObjectValue{
		Name:    invalidObject.GetName(),
		PkgPath: invalidObject.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "name", Value: nil, Assigned: true},
		},
	}
	if _, err := ext.SetModelValueWithScenario(invalidObject, &ValueImpl{value: invalidValue}, verrors.ScenarioInsert, false); err == nil {
		t.Fatal("SetModelValueWithScenario should reject invalid required field")
	}

	validObject := buildRemoteValidationTestObject()
	validValue := &ObjectValue{
		Name:    validObject.GetName(),
		PkgPath: validObject.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "name", Value: "apple"},
		},
	}
	ret, err := ext.SetModelValueWithScenario(validObject, &ValueImpl{value: validValue}, verrors.ScenarioInsert, false)
	if err != nil {
		t.Fatalf("SetModelValueWithScenario(valid) failed: %v", err)
	}
	if got := ret.(*Object).GetField("name").GetValue().Get(); got != "apple" {
		t.Fatalf("SetModelValueWithScenario(valid) mismatch, got %#v", got)
	}

	disabledObject := buildRemoteValidationTestObject()
	if _, err := ext.SetModelValueWithScenario(disabledObject, &ValueImpl{value: invalidValue}, verrors.ScenarioInsert, true); err != nil {
		t.Fatalf("disableValidator should bypass scenario validation, got %v", err)
	}

	unknownFieldObject := buildRemoteValidationTestObject()
	unknownFieldValue := &ObjectValue{
		Name:    unknownFieldObject.GetName(),
		PkgPath: unknownFieldObject.GetPkgPath(),
		Fields: []*FieldValue{
			{Name: "unknown", Value: "x"},
		},
	}
	if _, err := ext.SetModelValueWithScenario(unknownFieldObject, &ValueImpl{value: unknownFieldValue}, verrors.ScenarioInsert, false); err == nil {
		t.Fatal("SetModelValueWithScenario should reject unknown field")
	}
}

func TestRemoteValidationExtensionConfigureValidation(t *testing.T) {
	ext := NewValidationExtension(nil)
	object := buildRemoteValidationTestObject()
	nameField := object.GetField("name")

	config := validation.DefaultConfig()
	config.EnableTypeValidation = false
	config.EnableConstraintValidation = false
	config.EnableScenarioAdaptation = false
	if err := ext.ConfigureValidation(config); err != nil {
		t.Fatalf("ConfigureValidation failed: %v", err)
	}

	if ext.GetValidationManager() == nil {
		t.Fatal("GetValidationManager should stay non-nil after ConfigureValidation")
	}
	if err := ext.ValidateFieldWithScenario(nameField, nil, verrors.ScenarioInsert, false); err != nil {
		t.Fatalf("validation should be disabled by config, got %v", err)
	}
	if _, err := ext.SetModelValueWithScenario(buildRemoteValidationTestObject(), &ValueImpl{value: &ObjectValue{
		Name:    "product",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "name", Value: nil, Assigned: true},
		},
	}}, verrors.ScenarioInsert, false); err != nil {
		t.Fatalf("SetModelValueWithScenario should succeed after disabling validation, got %v", err)
	}
}

func TestRemoteValidationExtensionPrimaryValueAndScenarioMapping(t *testing.T) {
	ext := NewValidationExtension(nil)

	object := buildRemoteValidationTestObject()
	ret, err := ext.SetModelValueWithScenario(object, &ValueImpl{value: int64(9)}, verrors.ScenarioUpdate, false)
	if err != nil {
		t.Fatalf("SetModelValueWithScenario(primary) failed: %v", err)
	}
	if got := ret.(*Object).GetPrimaryField().GetValue().Get(); got != int64(9) {
		t.Fatalf("SetModelValueWithScenario(primary) mismatch, got %#v", got)
	}

	if _, err := ext.SetModelValueWithScenario(buildRemoteValidationTestObject(), &ValueImpl{}, verrors.ScenarioUpdate, false); err == nil {
		t.Fatal("SetModelValueWithScenario should reject invalid primary value")
	}

	impl, ok := ext.(*validationExtensionImpl)
	if !ok {
		t.Fatalf("expected *validationExtensionImpl, got %T", ext)
	}
	if got := impl.getOperationType(verrors.ScenarioInsert); got != validation.OperationCreate {
		t.Fatalf("getOperationType(insert) mismatch, got %v", got)
	}
	if got := impl.getOperationType(verrors.ScenarioUpdate); got != validation.OperationUpdate {
		t.Fatalf("getOperationType(update) mismatch, got %v", got)
	}
	if got := impl.getOperationType(verrors.ScenarioQuery); got != validation.OperationRead {
		t.Fatalf("getOperationType(query) mismatch, got %v", got)
	}
	if got := impl.getOperationType(verrors.ScenarioDelete); got != validation.OperationDelete {
		t.Fatalf("getOperationType(delete) mismatch, got %v", got)
	}
	if got := impl.getOperationType(verrors.Scenario("custom")); got != validation.OperationCreate {
		t.Fatalf("getOperationType(default) mismatch, got %v", got)
	}
}
