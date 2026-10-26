package database

import (
	"github.com/muidea/magicOrm/models"
)

// DatabaseValidator validates database-specific constraints
type DatabaseValidator interface {
	ValidateDatabaseConstraints(value any, fieldName string, constraints models.Constraints, dbType string) error
	GetDatabaseConstraints(constraints models.Constraints) []string
	ConvertToDatabaseValue(value any) (any, error)
}
