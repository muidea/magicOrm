package validation

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/models"
)

var interfaceType = reflect.TypeOf((*any)(nil)).Elem()

func exactFallbackType(fallback any) (reflect.Type, bool) {
	if fallback == nil {
		return nil, false
	}
	return reflect.TypeOf(fallback), true
}

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

// ReflectTypeFromModelType converts models.Type to a best-effort reflect.Type.
// Complex relation types fall back to interface{} when the concrete type cannot be inferred.
func ReflectTypeFromModelType(fieldType models.Type, fallback any) reflect.Type {
	if fieldType == nil {
		if fallback != nil {
			return reflect.TypeOf(fallback)
		}
		return interfaceType
	}

	var ret reflect.Type
	usedExactFallback := false
	switch fieldType.GetValue() {
	case models.TypeBooleanValue:
		ret = reflect.TypeOf(false)
	case models.TypeByteValue:
		ret = reflect.TypeOf(int8(0))
	case models.TypeSmallIntegerValue:
		ret = reflect.TypeOf(int16(0))
	case models.TypeInteger32Value:
		ret = reflect.TypeOf(int32(0))
	case models.TypeIntegerValue:
		ret = reflect.TypeOf(int(0))
	case models.TypeBigIntegerValue:
		ret = reflect.TypeOf(int64(0))
	case models.TypePositiveByteValue:
		ret = reflect.TypeOf(uint8(0))
	case models.TypePositiveSmallIntegerValue:
		ret = reflect.TypeOf(uint16(0))
	case models.TypePositiveInteger32Value:
		ret = reflect.TypeOf(uint32(0))
	case models.TypePositiveIntegerValue:
		ret = reflect.TypeOf(uint(0))
	case models.TypePositiveBigIntegerValue:
		ret = reflect.TypeOf(uint64(0))
	case models.TypeFloatValue:
		ret = reflect.TypeOf(float32(0))
	case models.TypeDoubleValue:
		ret = reflect.TypeOf(float64(0))
	case models.TypeStringValue:
		ret = reflect.TypeOf("")
	case models.TypeDateTimeValue:
		if fallbackType, ok := exactFallbackType(fallback); ok {
			ret = fallbackType
			usedExactFallback = true
		} else {
			ret = reflect.TypeOf(time.Time{})
		}
	case models.TypeSliceValue:
		if fallbackType, ok := exactFallbackType(fallback); ok {
			ret = fallbackType
			usedExactFallback = true
		} else {
			elemType := ReflectTypeFromModelType(fieldType.Elem(), nil)
			if elemType == nil {
				elemType = interfaceType
			}
			ret = reflect.SliceOf(elemType)
		}
	case models.TypeStructValue:
		if fallbackType, ok := exactFallbackType(fallback); ok {
			ret = fallbackType
			usedExactFallback = true
		} else {
			ret = interfaceType
		}
	default:
		if fallback != nil {
			ret = reflect.TypeOf(fallback)
		} else {
			ret = interfaceType
		}
	}

	if !usedExactFallback && fieldType.IsPtrType() && ret != nil && ret.Kind() != reflect.Ptr {
		ret = reflect.PointerTo(ret)
	}

	return ret
}

// AdaptField converts models.Field to FieldAdapter.
func AdaptField(field models.Field, value any) FieldAdapter {
	if field == nil {
		return nil
	}

	var constraints models.Constraints
	if spec := field.GetSpec(); spec != nil {
		constraints = spec.GetConstraints()
	}

	return NewFieldAdapter(
		field.GetName(),
		ReflectTypeFromModelType(field.GetType(), value),
		constraints,
		value,
	)
}

// AdaptModel converts models.Model to ModelAdapter.
func AdaptModel(model models.Model) ModelAdapter {
	if model == nil {
		return NewModelAdapter(nil)
	}

	fields := model.GetFields()
	adapters := make([]FieldAdapter, 0, len(fields))
	for _, field := range fields {
		var value any
		if fieldValue := field.GetValue(); fieldValue != nil {
			value = fieldValue.Get()
		}
		adapters = append(adapters, AdaptField(field, value))
	}

	return NewModelAdapter(adapters)
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
