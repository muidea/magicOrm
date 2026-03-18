package remote

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
	"github.com/muidea/magicOrm/validation/errors"
	"log/slog"
)

// ValidationExtension provides scenario-aware validation for remote provider
type ValidationExtension interface {
	// SetModelValueWithScenario sets model value with scenario-aware validation
	SetModelValueWithScenario(vModel models.Model, vVal models.Value, scenario errors.Scenario, disableValidator bool) (ret models.Model, err *cd.Error)

	// ValidateFieldWithScenario validates a field with scenario-aware validation
	ValidateFieldWithScenario(field models.Field, value any, scenario errors.Scenario, disableValidator bool) (err *cd.Error)

	// ValidateModelWithScenario validates entire model with scenario-aware validation
	ValidateModelWithScenario(model models.Model, scenario errors.Scenario, disableValidator bool) (err *cd.Error)

	// GetValidationManager returns the validation manager
	GetValidationManager() validation.ValidationManager

	// ConfigureValidation configures validation settings
	ConfigureValidation(config validation.ValidationConfig) error
}

// validationExtensionImpl implements ValidationExtension
type validationExtensionImpl struct {
	validationManager validation.ValidationManager
	valueValidator    models.ValueValidator
}

// NewValidationExtension creates a new validation extension
func NewValidationExtension(valueValidator models.ValueValidator) ValidationExtension {
	// Create default validation configuration
	config := validation.DefaultConfig()
	config.EnableCaching = false

	// Create validation factory
	factory := validation.NewValidationFactory()

	// Create validation manager
	validationManager := factory.CreateValidationManager(config)

	return &validationExtensionImpl{
		validationManager: validationManager,
		valueValidator:    valueValidator,
	}
}

// SetModelValueWithScenario sets model value with scenario-aware validation
func (e *validationExtensionImpl) SetModelValueWithScenario(vModel models.Model, vVal models.Value, scenario errors.Scenario, disableValidator bool) (ret models.Model, err *cd.Error) {
	if disableValidator {
		// Use original implementation when validation is disabled
		return SetModelValue(vModel, vVal, true)
	}

	vObjectPtr := vModel.(*Object)
	switch val := vVal.Get().(type) {
	case *ObjectValue:
		// Validate each field in ObjectValue
		for idx := range val.Fields {
			fieldVal := val.Fields[idx]
			if !fieldVal.Assigned && fieldVal.GetValue().IsZero() {
				continue
			}

			// Find the corresponding Field in the Object
			var targetField models.Field
			for _, field := range vObjectPtr.Fields {
				if field.GetName() == fieldVal.Name {
					targetField = field
					break
				}
			}

			if targetField == nil {
				err = cd.NewError(cd.Unexpected, fmt.Sprintf("field not found: %s", fieldVal.Name))
				slog.Error("validation field not found", "field", fieldVal.Name)
				return
			}
			err = e.ValidateFieldWithScenario(targetField, fieldVal.Value, scenario, false)
			if err != nil {
				slog.Error("ValidateFieldWithScenario failed", "field", fieldVal.Name, "error", err.Error())
				return
			}

			// Set the value after validation
			setErr := vObjectPtr.innerSetFieldValue(fieldVal.Name, fieldVal.Value, true)
			if setErr != nil {
				err = setErr
				slog.Error("innerSetFieldValue failed", "field", fieldVal.Name, "error", setErr.Error())
				return
			}
		}
	default:
		// Handle primary field value
		if vVal.IsValid() {
			// For primary field, we need to find the primary field and validate it
			primaryField := vObjectPtr.GetPrimaryField()
			if primaryField != nil {
				err = e.ValidateFieldWithScenario(primaryField, val, scenario, false)
				if err != nil {
					slog.Error("ValidateFieldWithScenario primary failed", "error", err.Error())
					return
				}
			}

			// Set the value after validation
			err = vObjectPtr.innerSetPrimaryFieldValue(val, true)
			if err != nil {
				slog.Error("innerSetPrimaryFieldValue failed", "error", err.Error())
				return
			}
		} else {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal model value, val:%v", val))
			slog.Error("illegal model value", "val", val)
			return
		}
	}

	ret = vModel
	return
}

// ValidateFieldWithScenario validates a field with scenario-aware validation
func (e *validationExtensionImpl) ValidateFieldWithScenario(field models.Field, value any, scenario errors.Scenario, disableValidator bool) (err *cd.Error) {
	if disableValidator {
		return nil
	}

	// Convert field to validation field adapter
	fieldAdapter := e.createFieldAdapter(field, value)
	if fieldAdapter == nil {
		return nil
	}

	// Create validation context
	ctx := validation.NewContext(
		scenario,
		e.getOperationType(scenario),
		nil,
		"", // database type not specified for remote validation
	)
	ctx.Field = fieldAdapter

	// Perform validation
	validationErr := e.validationManager.Validate(value, ctx)
	if validationErr != nil {
		return cd.NewError(cd.IllegalParam, validationErr.Error())
	}

	return nil
}

// ValidateModelWithScenario validates entire model with scenario-aware validation
func (e *validationExtensionImpl) ValidateModelWithScenario(model models.Model, scenario errors.Scenario, disableValidator bool) (err *cd.Error) {
	if disableValidator {
		return nil
	}

	// Convert model to validation model adapter
	modelAdapter := e.createModelAdapter(model)
	if modelAdapter == nil {
		return nil
	}

	// Create validation context
	ctx := validation.NewContext(
		scenario,
		e.getOperationType(scenario),
		modelAdapter,
		"", // database type not specified for remote validation
	)

	// Perform validation
	validationErr := e.validationManager.ValidateModel(model, ctx)
	if validationErr != nil {
		return cd.NewError(cd.IllegalParam, validationErr.Error())
	}

	return nil
}

// GetValidationManager returns the validation manager
func (e *validationExtensionImpl) GetValidationManager() validation.ValidationManager {
	return e.validationManager
}

// ConfigureValidation configures validation settings
func (e *validationExtensionImpl) ConfigureValidation(config validation.ValidationConfig) error {
	// Recreate validation manager with new configuration
	factory := validation.NewValidationFactory()
	e.validationManager = factory.CreateValidationManager(config)
	return nil
}

// Helper methods

// createFieldAdapter creates a field adapter for validation
func (e *validationExtensionImpl) createFieldAdapter(field models.Field, value any) validation.FieldAdapter {
	return validation.AdaptField(field, value)
}

// createModelAdapter creates a model adapter for validation
func (e *validationExtensionImpl) createModelAdapter(model models.Model) validation.ModelAdapter {
	return validation.AdaptModel(model)
}

// getOperationType maps scenario to operation type
func (e *validationExtensionImpl) getOperationType(scenario errors.Scenario) validation.OperationType {
	switch scenario {
	case errors.ScenarioInsert:
		return validation.OperationCreate
	case errors.ScenarioUpdate:
		return validation.OperationUpdate
	case errors.ScenarioQuery:
		return validation.OperationRead
	case errors.ScenarioDelete:
		return validation.OperationDelete
	default:
		return validation.OperationCreate
	}
}
