package types

import (
	"fmt"
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

// TypeValidator validates basic type compatibility and handles type conversions
type TypeValidator interface {
	// ValidateType validates if a value is compatible with the given type
	ValidateType(value any, fieldType reflect.Type) error

	// Convert converts a value to the target type
	Convert(value any, targetType reflect.Type) (any, error)

	// GetSupportedTypes returns all types supported by this validator
	GetSupportedTypes() []reflect.Type

	// RegisterTypeHandler registers a custom type handler
	RegisterTypeHandler(typeName string, handler TypeHandler) error

	// GetZeroValue returns the zero value for a type
	GetZeroValue(fieldType reflect.Type) any
}

// TypeHandler handles validation and conversion for a specific type
type TypeHandler interface {
	// Validate validates if a value is valid for this type
	Validate(value any) error

	// Convert converts a value to this type
	Convert(value any) (any, error)

	// GetZeroValue returns the zero value for this type
	GetZeroValue() any

	// GetType returns the reflect.Type this handler handles
	GetType() reflect.Type
}

// typeValidatorImpl implements TypeValidator
type typeValidatorImpl struct {
	typeHandlers map[reflect.Type]TypeHandler
	typeRegistry map[string]reflect.Type
}

// NewTypeValidator creates a new TypeValidator with built-in type support
func NewTypeValidator() TypeValidator {
	v := &typeValidatorImpl{
		typeHandlers: make(map[reflect.Type]TypeHandler),
		typeRegistry: make(map[string]reflect.Type),
	}

	// Register built-in type handlers
	v.registerBuiltinHandlers()

	return v
}

// ValidateType validates if a value is compatible with the given type
func (v *typeValidatorImpl) ValidateType(value any, fieldType reflect.Type) error {
	if value == nil {
		// nil is valid for pointer types
		if fieldType.Kind() == reflect.Ptr {
			return nil
		}
		return cd.NewError(cd.IllegalParam, fmt.Sprintf("nil value not allowed for non-pointer type %v", fieldType))
	}

	valType := reflect.TypeOf(value)

	// Check if types match directly
	if valType.AssignableTo(fieldType) {
		return nil
	}

	// Check if we can convert between types
	if valType.ConvertibleTo(fieldType) {
		return nil
	}

	// Check for pointer/value compatibility
	if fieldType.Kind() == reflect.Ptr && valType.AssignableTo(fieldType.Elem()) {
		return nil
	}

	// Check if we have a handler for this type
	if handler, ok := v.typeHandlers[fieldType]; ok {
		return handler.Validate(value)
	}

	return cd.NewError(cd.IllegalParam,
		fmt.Sprintf("value type %v is not compatible with field type %v", valType, fieldType))
}

// Convert converts a value to the target type
func (v *typeValidatorImpl) Convert(value any, targetType reflect.Type) (any, error) {
	if value == nil {
		// Return zero value for nil
		return v.GetZeroValue(targetType), nil
	}

	valType := reflect.TypeOf(value)

	// Direct assignment
	if valType.AssignableTo(targetType) {
		return value, nil
	}

	// Standard conversion
	if valType.ConvertibleTo(targetType) {
		val := reflect.ValueOf(value)
		converted := val.Convert(targetType)
		return converted.Interface(), nil
	}

	// Pointer to value conversion
	if targetType.Kind() == reflect.Ptr && valType.AssignableTo(targetType.Elem()) {
		val := reflect.ValueOf(value)
		ptr := reflect.New(targetType.Elem())
		ptr.Elem().Set(val)
		return ptr.Interface(), nil
	}

	// Value to pointer conversion
	if valType.Kind() == reflect.Ptr && valType.Elem().AssignableTo(targetType) {
		val := reflect.ValueOf(value)
		if val.IsNil() {
			return v.GetZeroValue(targetType), nil
		}
		return val.Elem().Interface(), nil
	}

	// Use type handler if available
	if handler, ok := v.typeHandlers[targetType]; ok {
		return handler.Convert(value)
	}

	return nil, cd.NewError(cd.IllegalParam,
		fmt.Sprintf("cannot convert %v to %v", valType, targetType))
}

// GetSupportedTypes returns all types supported by this validator
func (v *typeValidatorImpl) GetSupportedTypes() []reflect.Type {
	types := make([]reflect.Type, 0, len(v.typeHandlers))
	for t := range v.typeHandlers {
		types = append(types, t)
	}
	return types
}

// RegisterTypeHandler registers a custom type handler
func (v *typeValidatorImpl) RegisterTypeHandler(typeName string, handler TypeHandler) error {
	t := handler.GetType()
	v.typeHandlers[t] = handler
	v.typeRegistry[typeName] = t
	return nil
}

// GetZeroValue returns the zero value for a type
func (v *typeValidatorImpl) GetZeroValue(fieldType reflect.Type) any {
	if handler, ok := v.typeHandlers[fieldType]; ok {
		return handler.GetZeroValue()
	}

	// Return reflect zero value
	return reflect.Zero(fieldType).Interface()
}

// registerBuiltinHandlers registers built-in type handlers
func (v *typeValidatorImpl) registerBuiltinHandlers() {
	// Basic types
	v.registerBasicType("int", reflect.TypeOf(int(0)))
	v.registerBasicType("int8", reflect.TypeOf(int8(0)))
	v.registerBasicType("int16", reflect.TypeOf(int16(0)))
	v.registerBasicType("int32", reflect.TypeOf(int32(0)))
	v.registerBasicType("int64", reflect.TypeOf(int64(0)))
	v.registerBasicType("uint", reflect.TypeOf(uint(0)))
	v.registerBasicType("uint8", reflect.TypeOf(uint8(0)))
	v.registerBasicType("uint16", reflect.TypeOf(uint16(0)))
	v.registerBasicType("uint32", reflect.TypeOf(uint32(0)))
	v.registerBasicType("uint64", reflect.TypeOf(uint64(0)))
	v.registerBasicType("float32", reflect.TypeOf(float32(0)))
	v.registerBasicType("float64", reflect.TypeOf(float64(0)))
	v.registerBasicType("bool", reflect.TypeOf(false))
	v.registerBasicType("string", reflect.TypeOf(""))

	// Time type
	v.registerTimeType()

	// Slice types
	v.registerSliceType("[]int", reflect.TypeOf([]int{}))
	v.registerSliceType("[]string", reflect.TypeOf([]string{}))
	v.registerSliceType("[]float64", reflect.TypeOf([]float64{}))

	// Pointer types
	v.registerPointerType("*int", reflect.TypeOf((*int)(nil)))
	v.registerPointerType("*string", reflect.TypeOf((*string)(nil)))
	v.registerPointerType("*bool", reflect.TypeOf((*bool)(nil)))
}

// registerBasicType registers a basic type handler
func (v *typeValidatorImpl) registerBasicType(typeName string, t reflect.Type) {
	handler := &basicTypeHandler{typ: t}
	v.typeHandlers[t] = handler
	v.typeRegistry[typeName] = t
}

// registerTimeType registers time.Time type handler
func (v *typeValidatorImpl) registerTimeType() {
	t := reflect.TypeOf(time.Time{})
	handler := &timeTypeHandler{typ: t}
	v.typeHandlers[t] = handler
	v.typeRegistry["time.Time"] = t
}

// registerSliceType registers a slice type handler
func (v *typeValidatorImpl) registerSliceType(typeName string, t reflect.Type) {
	handler := &sliceTypeHandler{typ: t}
	v.typeHandlers[t] = handler
	v.typeRegistry[typeName] = t
}

// registerPointerType registers a pointer type handler
func (v *typeValidatorImpl) registerPointerType(typeName string, t reflect.Type) {
	handler := &pointerTypeHandler{typ: t}
	v.typeHandlers[t] = handler
	v.typeRegistry[typeName] = t
}

// basicTypeHandler handles basic Go types
type basicTypeHandler struct {
	typ reflect.Type
}

func (h *basicTypeHandler) Validate(value any) error {
	valType := reflect.TypeOf(value)
	if !valType.AssignableTo(h.typ) && !valType.ConvertibleTo(h.typ) {
		return cd.NewError(cd.IllegalParam,
			fmt.Sprintf("value type %v is not compatible with %v", valType, h.typ))
	}
	return nil
}

func (h *basicTypeHandler) Convert(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Type().ConvertibleTo(h.typ) {
		return val.Convert(h.typ).Interface(), nil
	}
	return nil, cd.NewError(cd.IllegalParam,
		fmt.Sprintf("cannot convert %v to %v", val.Type(), h.typ))
}

func (h *basicTypeHandler) GetZeroValue() any {
	return reflect.Zero(h.typ).Interface()
}

func (h *basicTypeHandler) GetType() reflect.Type {
	return h.typ
}

// timeTypeHandler handles time.Time type
type timeTypeHandler struct {
	typ reflect.Type
}

func (h *timeTypeHandler) Validate(value any) error {
	switch v := value.(type) {
	case time.Time:
		return nil
	case string:
		// Try to parse as time
		_, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return nil
		}
		// Try other common formats
		_, err = time.Parse("2006-01-02", v)
		if err == nil {
			return nil
		}
		return cd.NewError(cd.IllegalParam, "invalid time format")
	default:
		return cd.NewError(cd.IllegalParam,
			fmt.Sprintf("value type %T is not compatible with time.Time", value))
	}
}

func (h *timeTypeHandler) Convert(value any) (any, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try common time formats
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t, nil
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t, nil
		}
		return nil, cd.NewError(cd.IllegalParam, "cannot parse time string")
	default:
		return nil, cd.NewError(cd.IllegalParam,
			fmt.Sprintf("cannot convert %T to time.Time", value))
	}
}

func (h *timeTypeHandler) GetZeroValue() any {
	return time.Time{}
}

func (h *timeTypeHandler) GetType() reflect.Type {
	return h.typ
}

// sliceTypeHandler handles slice types
type sliceTypeHandler struct {
	typ reflect.Type
}

func (h *sliceTypeHandler) Validate(value any) error {
	valType := reflect.TypeOf(value)
	if valType.Kind() != reflect.Slice && valType.Kind() != reflect.Array {
		return cd.NewError(cd.IllegalParam,
			fmt.Sprintf("value type %v is not a slice or array", valType))
	}

	// Check element type compatibility
	elemType := h.typ.Elem()
	valElemType := valType.Elem()

	if !valElemType.AssignableTo(elemType) && !valElemType.ConvertibleTo(elemType) {
		return cd.NewError(cd.IllegalParam,
			fmt.Sprintf("slice element type %v is not compatible with %v",
				valElemType, elemType))
	}

	return nil
}

func (h *sliceTypeHandler) Convert(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Type().AssignableTo(h.typ) {
		return val.Interface(), nil
	}

	// Create new slice and convert elements
	elemType := h.typ.Elem()
	length := val.Len()
	result := reflect.MakeSlice(h.typ, length, length)

	for i := 0; i < length; i++ {
		elem := val.Index(i)
		if elem.Type().ConvertibleTo(elemType) {
			result.Index(i).Set(elem.Convert(elemType))
		} else {
			return nil, cd.NewError(cd.IllegalParam,
				fmt.Sprintf("cannot convert slice element %v to %v",
					elem.Type(), elemType))
		}
	}

	return result.Interface(), nil
}

func (h *sliceTypeHandler) GetZeroValue() any {
	return reflect.Zero(h.typ).Interface()
}

func (h *sliceTypeHandler) GetType() reflect.Type {
	return h.typ
}

// pointerTypeHandler handles pointer types
type pointerTypeHandler struct {
	typ reflect.Type
}

func (h *pointerTypeHandler) Validate(value any) error {
	if value == nil {
		return nil // nil is valid for pointers
	}

	valType := reflect.TypeOf(value)
	elemType := h.typ.Elem()

	// Check if value is assignable to pointer element type
	if valType.AssignableTo(elemType) || valType.ConvertibleTo(elemType) {
		return nil
	}

	// Check if value is a pointer to compatible type
	if valType.Kind() == reflect.Ptr {
		valElemType := valType.Elem()
		if valElemType.AssignableTo(elemType) || valElemType.ConvertibleTo(elemType) {
			return nil
		}
	}

	return cd.NewError(cd.IllegalParam,
		fmt.Sprintf("value type %v is not compatible with pointer type %v",
			valType, h.typ))
}

func (h *pointerTypeHandler) Convert(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	val := reflect.ValueOf(value)
	elemType := h.typ.Elem()

	// If value is already the right pointer type
	if val.Type().AssignableTo(h.typ) {
		return val.Interface(), nil
	}

	// If value is assignable to element type
	if val.Type().AssignableTo(elemType) {
		ptr := reflect.New(elemType)
		ptr.Elem().Set(val)
		return ptr.Interface(), nil
	}

	// If value is convertible to element type
	if val.Type().ConvertibleTo(elemType) {
		ptr := reflect.New(elemType)
		ptr.Elem().Set(val.Convert(elemType))
		return ptr.Interface(), nil
	}

	// If value is a pointer to convertible type
	if val.Type().Kind() == reflect.Ptr && !val.IsNil() {
		elem := val.Elem()
		if elem.Type().ConvertibleTo(elemType) {
			ptr := reflect.New(elemType)
			ptr.Elem().Set(elem.Convert(elemType))
			return ptr.Interface(), nil
		}
	}

	return nil, cd.NewError(cd.IllegalParam,
		fmt.Sprintf("cannot convert %v to pointer type %v", val.Type(), h.typ))
}

func (h *pointerTypeHandler) GetZeroValue() any {
	return nil
}

func (h *pointerTypeHandler) GetType() reflect.Type {
	return h.typ
}
