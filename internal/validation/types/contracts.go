package types

import (
	"reflect"

	"github.com/muidea/magicOrm/models"
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
	RegisterTypeHandler(typeName string, handler models.TypeHandler) error

	// GetZeroValue returns the zero value for a type
	GetZeroValue(fieldType reflect.Type) any
}

// TypeHandler names the shared extension contract inside the implementation.
type TypeHandler = models.TypeHandler
