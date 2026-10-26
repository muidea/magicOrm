package constraints

import (
	"errors"
	"testing"

	"github.com/muidea/magicOrm/models"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

type constraintDirective struct {
	key  models.Key
	args []string
}

func (d constraintDirective) Key() models.Key { return d.key }
func (d constraintDirective) Args() []string  { return d.args }
func (d constraintDirective) HasArgs() bool   { return len(d.args) > 0 }

type constraintSet struct {
	directives []models.Directive
}

func (c constraintSet) Has(key models.Key) bool {
	_, ok := c.Get(key)
	return ok
}
func (c constraintSet) Get(key models.Key) (models.Directive, bool) {
	for _, directive := range c.directives {
		if directive.Key() == key {
			return directive, true
		}
	}
	return nil, false
}
func (c constraintSet) Directives() []models.Directive { return c.directives }

func TestConstraintValidatorScenarios(t *testing.T) {
	validator := NewConstraintValidator(true)

	insertConstraints := constraintSet{directives: []models.Directive{
		constraintDirective{key: models.KeyRequired},
		constraintDirective{key: models.KeyMin, args: []string{"3"}},
	}}
	if err := validator.ValidateConstraints("ab", insertConstraints, verrors.ScenarioInsert); err == nil {
		t.Fatal("expected insert validation to reject short value")
	}
	if err := validator.ValidateConstraints("abcd", insertConstraints, verrors.ScenarioInsert); err != nil {
		t.Fatalf("expected insert validation success, got %v", err)
	}

	queryConstraints := constraintSet{directives: []models.Directive{
		constraintDirective{key: models.KeyWriteOnly},
	}}
	if err := validator.ValidateConstraints("secret", queryConstraints, verrors.ScenarioQuery); err != nil {
		t.Fatalf("expected query scenario to skip write-only validation, got %v", err)
	}

	deleteConstraints := constraintSet{directives: []models.Directive{
		constraintDirective{key: models.KeyRequired},
	}}
	if err := validator.ValidateConstraints(nil, deleteConstraints, verrors.ScenarioDelete); err != nil {
		t.Fatalf("expected delete scenario to skip constraints, got %v", err)
	}
}

func TestConstraintValidatorCustomAndCache(t *testing.T) {
	validator := NewConstraintValidator(true)
	impl := validator.(*constraintValidatorImpl)

	customKey := models.Key("even")
	if err := validator.RegisterCustomConstraint(customKey, func(val any, _ []string) error {
		v, ok := val.(int)
		if !ok || v%2 != 0 {
			return errors.New("must be even")
		}
		return nil
	}); err != nil {
		t.Fatalf("register custom constraint failed: %v", err)
	}

	customConstraints := constraintSet{directives: []models.Directive{
		constraintDirective{key: customKey},
	}}
	if err := validator.ValidateConstraints(3, customConstraints, verrors.ScenarioInsert); err == nil {
		t.Fatal("expected custom constraint failure")
	}
	if err := validator.ValidateConstraints(4, customConstraints, verrors.ScenarioInsert); err != nil {
		t.Fatalf("expected custom constraint success, got %v", err)
	}

	if impl.validationCache == nil {
		t.Fatal("expected validation cache when caching is enabled")
	}
	if _, ok := impl.validationCache.GetConstraintResult(4, customConstraints, verrors.ScenarioInsert); !ok {
		t.Fatal("expected successful validation result to be cached")
	}

	validator.ClearCache()
	if _, ok := impl.validationCache.GetConstraintResult(4, customConstraints, verrors.ScenarioInsert); ok {
		t.Fatal("expected cache to be empty after ClearCache")
	}
}

func TestGetApplicableConstraints(t *testing.T) {
	validator := NewConstraintValidator(false)

	insertKeys := validator.GetApplicableConstraints(verrors.ScenarioInsert)
	if len(insertKeys) == 0 {
		t.Fatal("expected insert scenario to have applicable constraints")
	}

	unknownKeys := validator.GetApplicableConstraints(verrors.Scenario("unknown"))
	if len(unknownKeys) != 0 {
		t.Fatalf("expected unknown scenario to return no constraints, got %v", unknownKeys)
	}
}

func TestStandaloneConstraintHelpers(t *testing.T) {
	if err := ValidateRequired(nil, nil); err == nil {
		t.Fatal("expected required helper to reject nil")
	}
	if err := ValidateRequired("value", nil); err != nil {
		t.Fatalf("expected required helper success, got %v", err)
	}

	if err := ValidateMin("ab", []string{"3"}); err == nil {
		t.Fatal("expected min helper to reject short string")
	}
	if err := ValidateMin(5, []string{"3"}); err != nil {
		t.Fatalf("expected min helper success, got %v", err)
	}
	if err := ValidateMin("value", nil); err == nil {
		t.Fatal("expected min helper to require arg")
	}

	if err := ValidateMax("abcd", []string{"3"}); err == nil {
		t.Fatal("expected max helper to reject long string")
	}
	if err := ValidateMax(2, []string{"3"}); err != nil {
		t.Fatalf("expected max helper success, got %v", err)
	}
	if err := ValidateMax(struct{}{}, []string{"3"}); err == nil {
		t.Fatal("expected max helper to reject unsupported type")
	}
}
