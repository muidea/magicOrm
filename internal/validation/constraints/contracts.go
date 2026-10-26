package constraints

import (
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
)

// ConstraintValidator validates business constraints
type ConstraintValidator interface {
	// ValidateConstraints validates constraints for a value
	ValidateConstraints(value any, constraints models.Constraints, scenario errors.Scenario) error

	// GetApplicableDirectives returns the directives that are evaluated for a scenario.
	GetApplicableDirectives(constraints models.Constraints, scenario errors.Scenario) []models.Directive

	// GetApplicableConstraints returns constraints applicable for a scenario
	GetApplicableConstraints(scenario errors.Scenario) []models.Key

	// RegisterCustomConstraint registers a custom constraint
	RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error

	// ClearCache clears the constraint cache
	ClearCache()
}
