package validation

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/models"
)

// FieldAdapter adapts models.Field to validation system
type FieldAdapter interface {
	GetName() string
	GetType() reflect.Type
	GetConstraints() models.Constraints
	HasConstraint(key models.Key) bool
	GetValue() any
}

// ModelAdapter adapts models.Model to validation system
type ModelAdapter interface {
	GetFields() []FieldAdapter
	GetField(name string) (FieldAdapter, error)
}

// fieldAdapterImpl implements FieldAdapter
type fieldAdapterImpl struct {
	name        string
	fieldType   reflect.Type
	constraints models.Constraints
	value       any
}

// NewFieldAdapter creates a new field adapter
func NewFieldAdapter(name string, fieldType reflect.Type, constraints models.Constraints, value any) FieldAdapter {
	return &fieldAdapterImpl{
		name:        name,
		fieldType:   fieldType,
		constraints: constraints,
		value:       value,
	}
}

// GetName returns the field name
func (a *fieldAdapterImpl) GetName() string {
	return a.name
}

// GetType returns the field type as reflect.Type
func (a *fieldAdapterImpl) GetType() reflect.Type {
	return a.fieldType
}

// GetConstraints returns field constraints
func (a *fieldAdapterImpl) GetConstraints() models.Constraints {
	return a.constraints
}

// HasConstraint checks if field has a specific constraint
func (a *fieldAdapterImpl) HasConstraint(key models.Key) bool {
	if a.constraints != nil {
		return a.constraints.Has(key)
	}
	return false
}

// GetValue returns the field value
func (a *fieldAdapterImpl) GetValue() any {
	return a.value
}

// modelAdapterImpl implements ModelAdapter
type modelAdapterImpl struct {
	fields map[string]FieldAdapter
}

// NewModelAdapter creates a new model adapter
func NewModelAdapter(fields []FieldAdapter) ModelAdapter {
	fieldMap := make(map[string]FieldAdapter)
	for _, field := range fields {
		fieldMap[field.GetName()] = field
	}

	return &modelAdapterImpl{
		fields: fieldMap,
	}
}

// GetFields returns all field adapters
func (a *modelAdapterImpl) GetFields() []FieldAdapter {
	fields := make([]FieldAdapter, 0, len(a.fields))
	for _, field := range a.fields {
		fields = append(fields, field)
	}
	return fields
}

// GetField returns a specific field adapter
func (a *modelAdapterImpl) GetField(name string) (FieldAdapter, error) {
	field, exists := a.fields[name]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found", name)
	}
	return field, nil
}

// Helper functions for working with models

// GetFieldSpec gets the spec for a field
func GetFieldSpec(field models.Field) models.Spec {
	// Field has GetSpec() method
	return field.GetSpec()
}

// HasFieldConstraint checks if a field has a specific constraint
func HasFieldConstraint(field models.Field, key models.Key) bool {
	spec := GetFieldSpec(field)
	if spec != nil {
		constraints := spec.GetConstraints()
		if constraints != nil {
			return constraints.Has(key)
		}
	}
	return false
}

// GetFieldConstraints gets constraints for a field
func GetFieldConstraints(field models.Field) models.Constraints {
	spec := GetFieldSpec(field)
	if spec != nil {
		return spec.GetConstraints()
	}
	return nil
}

// GetFieldTypeName gets the type name for a field
func GetFieldTypeName(field models.Field) string {
	fieldType := field.GetType()
	if fieldType != nil {
		return fieldType.GetName()
	}
	return ""
}

// IsFieldRequired checks if a field is required
func IsFieldRequired(field models.Field) bool {
	return HasFieldConstraint(field, models.KeyRequired)
}

// IsFieldReadOnly checks if a field is read-only
func IsFieldReadOnly(field models.Field) bool {
	return HasFieldConstraint(field, models.KeyReadOnly)
}

// IsFieldWriteOnly checks if a field is write-only
func IsFieldWriteOnly(field models.Field) bool {
	return HasFieldConstraint(field, models.KeyWriteOnly)
}
