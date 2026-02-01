package validation

import (
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
)

// ScenarioAdapter orchestrates validation based on operation scenarios
type ScenarioAdapter interface {
	GetValidationStrategy(scenario errors.Scenario) ValidationStrategy
	ShouldValidateConstraint(constraint models.Key, scenario errors.Scenario) bool
}

// ValidationStrategy defines how validation should be performed for a scenario
type ValidationStrategy interface {
	ShouldValidate(constraint models.Key) bool
	ShouldSkipReadOnlyFields() bool
	ShouldSkipWriteOnlyFields() bool
	IsStrictMode() bool
}

// scenarioAdapterImpl implements ScenarioAdapter
type scenarioAdapterImpl struct {
	strategies map[errors.Scenario]ValidationStrategy
}

// NewScenarioAdapter creates a new scenario adapter
func NewScenarioAdapter() ScenarioAdapter {
	adapter := &scenarioAdapterImpl{
		strategies: make(map[errors.Scenario]ValidationStrategy),
	}

	// Initialize built-in strategies
	adapter.strategies[errors.ScenarioInsert] = &insertStrategy{}
	adapter.strategies[errors.ScenarioUpdate] = &updateStrategy{}
	adapter.strategies[errors.ScenarioQuery] = &queryStrategy{}
	adapter.strategies[errors.ScenarioDelete] = &deleteStrategy{}

	return adapter
}

// GetValidationStrategy returns the validation strategy for a scenario
func (a *scenarioAdapterImpl) GetValidationStrategy(scenario errors.Scenario) ValidationStrategy {
	strategy, exists := a.strategies[scenario]
	if !exists {
		// Default to insert strategy
		strategy = a.strategies[errors.ScenarioInsert]
	}
	return strategy
}

// ShouldValidateConstraint checks if a constraint should be validated in a scenario
func (a *scenarioAdapterImpl) ShouldValidateConstraint(constraint models.Key, scenario errors.Scenario) bool {
	strategy := a.GetValidationStrategy(scenario)
	return strategy.ShouldValidate(constraint)
}

// Built-in validation strategies

type insertStrategy struct{}

func (s *insertStrategy) ShouldValidate(constraint models.Key) bool {
	// Validate all constraints in insert scenario
	return true
}

func (s *insertStrategy) ShouldSkipReadOnlyFields() bool {
	return false
}

func (s *insertStrategy) ShouldSkipWriteOnlyFields() bool {
	return false
}

func (s *insertStrategy) IsStrictMode() bool {
	return true
}

type updateStrategy struct{}

func (s *updateStrategy) ShouldValidate(constraint models.Key) bool {
	// Validate read-only constraints in update scenario to prevent modification
	// This ensures ro fields are protected during updates
	return true
}

func (s *updateStrategy) ShouldSkipReadOnlyFields() bool {
	// Don't skip read-only fields - they should be validated to prevent modification
	return false
}

func (s *updateStrategy) ShouldSkipWriteOnlyFields() bool {
	return false
}

func (s *updateStrategy) IsStrictMode() bool {
	return false
}

type queryStrategy struct{}

func (s *queryStrategy) ShouldValidate(constraint models.Key) bool {
	// Skip write-only constraints in query scenario
	if constraint == models.KeyWriteOnly {
		return false
	}
	// Only validate basic constraints for queries
	switch constraint {
	case models.KeyMin, models.KeyMax, models.KeyRange, models.KeyIn, models.KeyRegexp:
		return true
	default:
		return false
	}
}

func (s *queryStrategy) ShouldSkipReadOnlyFields() bool {
	return false
}

func (s *queryStrategy) ShouldSkipWriteOnlyFields() bool {
	return true
}

func (s *queryStrategy) IsStrictMode() bool {
	return false
}

type deleteStrategy struct{}

func (s *deleteStrategy) ShouldValidate(constraint models.Key) bool {
	// Only validate required constraint for deletes
	return constraint == models.KeyRequired
}

func (s *deleteStrategy) ShouldSkipReadOnlyFields() bool {
	return true
}

func (s *deleteStrategy) ShouldSkipWriteOnlyFields() bool {
	return true
}

func (s *deleteStrategy) IsStrictMode() bool {
	return false
}
