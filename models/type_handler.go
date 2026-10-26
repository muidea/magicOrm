package models

import "reflect"

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
