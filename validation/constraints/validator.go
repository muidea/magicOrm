package constraints

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
	"github.com/muidea/magicOrm/validation/cache"
	"github.com/muidea/magicOrm/validation/errors"
)

// ConstraintValidator validates business constraints
type ConstraintValidator interface {
	// ValidateConstraints validates constraints for a value
	ValidateConstraints(value any, constraints models.Constraints, scenario errors.Scenario) error

	// GetApplicableConstraints returns constraints applicable for a scenario
	GetApplicableConstraints(scenario errors.Scenario) []models.Key

	// RegisterCustomConstraint registers a custom constraint
	RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error

	// ClearCache clears the constraint cache
	ClearCache()
}

// constraintValidatorImpl implements ConstraintValidator
type constraintValidatorImpl struct {
	baseValidator   models.ValueValidator
	scenarioRules   map[errors.Scenario]scenarioRule
	validationCache *cache.ValidationCache
	customHandlers  map[models.Key]models.ValidatorFunc
	enableCaching   bool
}

// scenarioRule defines which constraints apply to which scenarios
type scenarioRule struct {
	required []models.Key
	optional []models.Key
	skipped  []models.Key
	strict   bool
}

// NewConstraintValidator creates a new constraint validator
func NewConstraintValidator(enableCaching bool) ConstraintValidator {
	v := &constraintValidatorImpl{
		baseValidator:  utils.NewValueValidator(),
		scenarioRules:  make(map[errors.Scenario]scenarioRule),
		customHandlers: make(map[models.Key]models.ValidatorFunc),
		enableCaching:  enableCaching,
	}

	if enableCaching {
		// Create validation cache with default configuration
		cacheConfig := cache.DefaultCacheConfig()
		cacheConfig.Enabled = enableCaching
		v.validationCache = cache.NewValidationCache(cacheConfig)
	}

	// Initialize scenario rules
	v.initScenarioRules()

	return v
}

// ValidateConstraints validates constraints for a value
func (v *constraintValidatorImpl) ValidateConstraints(value any, constraints models.Constraints, scenario errors.Scenario) error {
	if constraints == nil {
		return nil
	}

	// Check cache first if enabled
	if v.enableCaching && v.validationCache != nil {
		if cachedResult, found := v.validationCache.GetConstraintResult(value, constraints, scenario); found {
			return cachedResult
		}
	}

	// Get applicable directives for this scenario
	directives := v.getApplicableDirectives(constraints, scenario)
	if len(directives) == 0 {
		return nil
	}

	// Validate using base validator
	err := v.baseValidator.ValidateValue(value, directives)

	// Cache the result if enabled
	if v.enableCaching && v.validationCache != nil && err == nil {
		// Only cache successful validations to avoid caching errors
		v.validationCache.SetConstraintResult(value, constraints, scenario, err)
	}

	return err
}

// GetApplicableConstraints returns constraints applicable for a scenario
func (v *constraintValidatorImpl) GetApplicableConstraints(scenario errors.Scenario) []models.Key {
	rule, exists := v.scenarioRules[scenario]
	if !exists {
		return []models.Key{}
	}

	// Combine required and optional constraints
	result := make([]models.Key, 0, len(rule.required)+len(rule.optional))
	result = append(result, rule.required...)
	result = append(result, rule.optional...)

	return result
}

// RegisterCustomConstraint registers a custom constraint
func (v *constraintValidatorImpl) RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error {
	v.customHandlers[key] = validator
	v.baseValidator.Register(key, validator)
	return nil
}

// ClearCache clears the constraint cache
func (v *constraintValidatorImpl) ClearCache() {
	if v.validationCache != nil {
		v.validationCache.Clear()
	}
}

// getApplicableDirectives returns directives applicable for the given scenario
func (v *constraintValidatorImpl) getApplicableDirectives(constraints models.Constraints, scenario errors.Scenario) []models.Directive {
	rule, exists := v.scenarioRules[scenario]
	if !exists {
		// Use default rule
		rule = v.scenarioRules[errors.ScenarioInsert]
	}

	allDirectives := constraints.Directives()
	applicable := make([]models.Directive, 0, len(allDirectives))

	for _, directive := range allDirectives {
		key := directive.Key()

		// Check if this constraint should be skipped
		if v.shouldSkipConstraint(key, rule) {
			continue
		}

		// Check if this constraint is required in strict mode
		if rule.strict && v.isRequiredConstraint(key, rule) && !constraints.Has(key) {
			// In strict mode, missing required constraints is an error
			// This would be handled by the caller
			continue
		}

		// Add directive if constraint exists
		if constraints.Has(key) {
			applicable = append(applicable, directive)
		}
	}

	return applicable
}

// shouldSkipConstraint checks if a constraint should be skipped for this scenario
func (v *constraintValidatorImpl) shouldSkipConstraint(key models.Key, rule scenarioRule) bool {
	// Check skipped constraints
	for _, skipped := range rule.skipped {
		if key == skipped {
			return true
		}
	}

	return false
}

// isRequiredConstraint checks if a constraint is required for this scenario
func (v *constraintValidatorImpl) isRequiredConstraint(key models.Key, rule scenarioRule) bool {
	for _, required := range rule.required {
		if key == required {
			return true
		}
	}
	return false
}

// getCacheKey generates a cache key for constraints and scenario

// initScenarioRules initializes scenario-specific constraint rules
func (v *constraintValidatorImpl) initScenarioRules() {
	// Insert scenario: strict validation, all constraints apply
	v.scenarioRules[errors.ScenarioInsert] = scenarioRule{
		required: []models.Key{
			models.KeyRequired,
		},
		optional: []models.Key{
			models.KeyMin,
			models.KeyMax,
			models.KeyRange,
			models.KeyIn,
			models.KeyRegexp,
			models.KeyReadOnly,
			models.KeyWriteOnly,
		},
		skipped: []models.Key{},
		strict:  true,
	}

	// Update scenario: relaxed validation, but validate read-only fields to prevent modification
	v.scenarioRules[errors.ScenarioUpdate] = scenarioRule{
		required: []models.Key{},
		optional: []models.Key{
			models.KeyRequired,
			models.KeyMin,
			models.KeyMax,
			models.KeyRange,
			models.KeyIn,
			models.KeyRegexp,
			models.KeyWriteOnly,
			models.KeyReadOnly, // Validate read-only fields to prevent modification
		},
		skipped: []models.Key{},
		strict:  false,
	}

	// Query scenario: minimal validation, skip write-only fields
	v.scenarioRules[errors.ScenarioQuery] = scenarioRule{
		required: []models.Key{},
		optional: []models.Key{
			models.KeyMin,
			models.KeyMax,
			models.KeyRange,
			models.KeyIn,
			models.KeyRegexp,
			models.KeyReadOnly,
		},
		skipped: []models.Key{
			models.KeyWriteOnly,
		},
		strict: false,
	}

	// Delete scenario: no validation needed
	v.scenarioRules[errors.ScenarioDelete] = scenarioRule{
		required: []models.Key{},
		optional: []models.Key{},
		skipped: []models.Key{
			models.KeyRequired,
			models.KeyMin,
			models.KeyMax,
			models.KeyRange,
			models.KeyIn,
			models.KeyRegexp,
			models.KeyReadOnly,
			models.KeyWriteOnly,
		},
		strict: false,
	}
}

// Helper functions for common validation patterns

// ValidateRequired validates required constraint
func ValidateRequired(value any, args []string) error {
	if value == nil {
		return errors.NewValidationError("value is required")
	}

	// Use existing utility if available
	// This is a simplified implementation
	val := reflect.ValueOf(value)
	if val.IsZero() {
		return errors.NewValidationError("value is required")
	}

	return nil
}

// ValidateMin validates minimum value/length
func ValidateMin(value any, args []string) error {
	if len(args) == 0 {
		return errors.NewValidationError("min constraint requires a value")
	}

	min, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return errors.NewValidationError("invalid min value")
	}

	val := reflect.ValueOf(value)
	var actual float64

	switch val.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		actual = float64(val.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		actual = float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		actual = float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		actual = val.Float()
	default:
		return errors.NewValidationError("min constraint not applicable to this type")
	}

	if actual < min {
		return errors.NewValidationError(fmt.Sprintf("value must be at least %v", min))
	}

	return nil
}

// ValidateMax validates maximum value/length
func ValidateMax(value any, args []string) error {
	if len(args) == 0 {
		return errors.NewValidationError("max constraint requires a value")
	}

	max, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return errors.NewValidationError("invalid max value")
	}

	val := reflect.ValueOf(value)
	var actual float64

	switch val.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		actual = float64(val.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		actual = float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		actual = float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		actual = val.Float()
	default:
		return errors.NewValidationError("max constraint not applicable to this type")
	}

	if actual > max {
		return errors.NewValidationError(fmt.Sprintf("value must be at most %v", max))
	}

	return nil
}
