package local

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation"
	"github.com/muidea/magicOrm/validation/errors"
)

// ValidationExtension provides scenario-aware validation for local provider
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

	valImplPtr, valImplOK := vVal.(*ValueImpl)
	if !valImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		log.Errorf("SetModelValueWithScenario failed, err:%s", err.Error())
		return
	}

	valueModel, valueModelErr := getValueModel(valImplPtr.value, models.OriginView)
	if valueModelErr != nil {
		err = valueModelErr
		log.Errorf("SetModelValueWithScenario failed, err:%s", err.Error())
		return
	}

	vModelImplPtr := vModel.(*objectImpl)
	fields := valueModel.GetFields()

	// Validate each field with scenario-aware validation
	for _, field := range fields {
		if !models.IsValidField(field) {
			continue
		}

		fieldValue := field.GetValue().Get()
		err = e.ValidateFieldWithScenario(field, fieldValue, scenario, false)
		if err != nil {
			log.Errorf("SetModelValueWithScenario failed, validate field:%s value err:%s", field.GetName(), err.Error())
			return
		}

		// Set the value after validation
		setErr := vModelImplPtr.innerSetFieldValue(field.GetName(), fieldValue, true)
		if setErr != nil {
			err = setErr
			log.Errorf("SetModelValueWithScenario failed, set field:%s value err:%s", field.GetName(), err.Error())
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
		"", // database type not specified for local validation
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
		"", // database type not specified for local validation
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
	// Get field constraints
	var constraints models.Constraints
	if spec := field.GetSpec(); spec != nil {
		constraints = spec.GetConstraints()
	}

	// Create field adapter
	return validation.NewFieldAdapter(
		field.GetName(),
		nil, // field type - would need conversion from models.Type to reflect.Type
		constraints,
		value,
	)
}

// createModelAdapter creates a model adapter for validation
func (e *validationExtensionImpl) createModelAdapter(model models.Model) validation.ModelAdapter {
	modelImpl, ok := model.(*objectImpl)
	if !ok {
		return nil
	}

	// Collect all field adapters
	fieldAdapters := make([]validation.FieldAdapter, 0, len(modelImpl.fields))
	for _, field := range modelImpl.fields {
		fieldValue := field.GetValue().Get()
		fieldAdapter := e.createFieldAdapter(field, fieldValue)
		if fieldAdapter != nil {
			fieldAdapters = append(fieldAdapters, fieldAdapter)
		}
	}

	// Create model adapter
	return validation.NewModelAdapter(fieldAdapters)
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

// Extended object method with scenario support
func (s *objectImpl) innerSetFieldValueWithScenario(name string, val any, scenario errors.Scenario, disableValidator bool) (err *cd.Error) {
	for _, sf := range s.fields {
		sf.valueValidator = s.valueValidator
		if sf.GetName() == name {
			// Try to use extended method if available
			// Note: We can't use type assertion to *field since it's private
			// For now, use the original method
			err = sf.innerSetValue(val, disableValidator)
			return
		}
	}

	log.Warnf("innerSetFieldValueWithScenario failed, field:%s not found", name)
	return
}
