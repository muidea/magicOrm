package database

import (
	"fmt"

	"github.com/muidea/magicOrm/models"
)

// DatabaseValidator validates database-specific constraints
type DatabaseValidator interface {
	ValidateDatabaseConstraints(value any, fieldName string, constraints models.Constraints, dbType string) error
	GetDatabaseConstraints(constraints models.Constraints) []string
	ConvertToDatabaseValue(value any) (any, error)
}

// databaseValidatorImpl implements DatabaseValidator
type databaseValidatorImpl struct{}

// NewDatabaseValidator creates a new database validator
func NewDatabaseValidator() DatabaseValidator {
	return &databaseValidatorImpl{}
}

// ValidateDatabaseConstraints validates database constraints for a value
func (v *databaseValidatorImpl) ValidateDatabaseConstraints(value any, fieldName string, constraints models.Constraints, dbType string) error {
	if dbType == "" {
		return nil
	}

	// Simple validation for now
	// Check NOT NULL constraint
	if constraints != nil && constraints.Has(models.KeyRequired) && value == nil {
		return fmt.Errorf("field '%s' cannot be null", fieldName)
	}

	return nil
}

// GetDatabaseConstraints returns database constraints
func (v *databaseValidatorImpl) GetDatabaseConstraints(constraints models.Constraints) []string {
	if constraints == nil {
		return []string{}
	}

	result := make([]string, 0)
	if constraints.Has(models.KeyRequired) {
		result = append(result, "NOT NULL")
	}

	return result
}

// ConvertToDatabaseValue converts a value for database storage
func (v *databaseValidatorImpl) ConvertToDatabaseValue(value any) (any, error) {
	return value, nil
}
