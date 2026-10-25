package validation

import (
	"fmt"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
	verrors "github.com/muidea/magicOrm/validation/errors"
	"testing"
	"time"
)

func reviewConfigField() *testField {
	return &testField{name: "value", typ: &testType{name: "int", value: models.TypeIntegerValue}, spec: &testSpec{constraints: utils.ParseConstraints("min=1")}, value: &testValue{value: 1, valid: true, assigned: true}}
}
func TestEffectiveConfiguredCacheTTL(t *testing.T) {
	for _, factory := range []bool{false, true} {
		t.Run(fmt.Sprint(factory), func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.EnableTypeValidation = false
			cfg.EnableDatabaseValidation = false
			cfg.EnableScenarioAdaptation = false
			cfg.CacheTTL = 10 * time.Millisecond
			m := NewValidationManager(cfg)
			if factory {
				m = NewValidationFactory().CreateValidationManager(cfg)
			}
			calls := 0
			m.RegisterCustomConstraint(models.KeyMin, func(any, []string) error { calls++; return nil })
			ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "")
			f := reviewConfigField()
			if err := m.ValidateField(f, 1, ctx); err != nil {
				t.Fatal(err)
			}
			<-time.NewTimer(2 * cfg.CacheTTL).C
			if err := m.ValidateField(f, 1, ctx); err != nil {
				t.Fatal(err)
			}
			if calls != 2 {
				t.Fatalf("configured TTL ignored: validator calls=%d, want 2 after expiry", calls)
			}
		})
	}
}
func TestEffectiveConfiguredCacheBudget(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableTypeValidation = false
	cfg.EnableDatabaseValidation = false
	cfg.EnableScenarioAdaptation = false
	cfg.MaxCacheSize = 1
	m := NewValidationManager(cfg)
	calls := 0
	m.RegisterCustomConstraint(models.KeyMin, func(any, []string) error { calls++; return nil })
	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "")
	f := reviewConfigField()
	for _, v := range []int{1, 2, 1} {
		if err := m.ValidateField(f, v, ctx); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 3 {
		t.Fatalf("configured one-entry cache ignored: validator calls=%d, want 3", calls)
	}
}
func TestEffectiveConfiguredStopOnFirstError(t *testing.T) {
	cfg := SimpleConfig()
	cfg.DefaultOptions.StopOnFirstError = true
	m := NewValidationManager(cfg)
	a := reviewConfigField()
	a.name = "first"
	a.value = &testValue{value: 0, valid: true, assigned: true}
	b := reviewConfigField()
	b.name = "second"
	b.value = &testValue{value: 0, valid: true, assigned: true}
	model := &testModel{name: "sample", fields: models.Fields{a, b}}
	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "")
	if err := m.ValidateModel(model, ctx); err == nil {
		t.Fatal("expected invalid input")
	}
	if got := m.GetValidationStats().TotalValidations; got != 1 {
		t.Fatalf("configured StopOnFirstError ignored: validated %d fields, want 1", got)
	}
}

func TestEffectiveContextBoundaries(t *testing.T) {
	cfg := SimpleConfig()
	cfg.DefaultOptions.StopOnFirstError = true
	m := NewValidationManager(cfg)
	f := reviewConfigField()
	f.value = &testValue{value: 0, valid: true, assigned: true}
	second := reviewConfigField()
	second.name = "second"
	second.value = &testValue{value: 0, valid: true, assigned: true}
	model := &testModel{name: "invalid", fields: models.Fields{f, second}}
	ctx := NewContext("", OperationUpdate, nil, "")
	m.SetScenario(verrors.ScenarioUpdate)
	// Update must skip unassigned fields before any validation.
	f.value = &testValue{value: 0, valid: true, assigned: false}
	second.value = &testValue{value: 0, valid: true, assigned: false}
	if err := m.ValidateModel(model, ctx); err != nil {
		t.Fatal(err)
	}
	m.SetScenario(verrors.ScenarioInsert)
	if m.ValidateModel(model, ctx) == nil {
		t.Fatal("context retained previous default scenario")
	}
	m.ResetStats()
	opts := cfg.DefaultOptions
	opts.StopOnFirstError = false
	if m.ValidateModel(model, ctx.WithOptions(opts)) == nil || m.GetValidationStats().TotalValidations != 2 {
		t.Fatal("explicit false did not reach outer traversal")
	}
	for _, entry := range []func(ValidationContext) error{func(c ValidationContext) error { return m.Validate(nil, c) }, func(c ValidationContext) error { return m.ValidateField(nil, nil, c) }, func(c ValidationContext) error { return m.ValidateModel(nil, c) }} {
		if entry(NewContext("unknown", OperationCreate, nil, "")) == nil {
			t.Fatal("unknown scenario accepted by short circuit")
		}
	}
}

func TestAdaptedModelAndRawModelResolveSameOptions(t *testing.T) {
	for _, adapted := range []bool{false, true} {
		cfg := SimpleConfig()
		cfg.DefaultOptions.StopOnFirstError = true
		m := NewValidationManager(cfg)
		field := reviewConfigField()
		field.value = &testValue{value: 0, valid: true, assigned: false}
		model := &testModel{name: "sample", fields: models.Fields{field}}
		ctx := NewContext(verrors.ScenarioUpdate, OperationUpdate, nil, "")
		if adapted {
			ctx.Model = AdaptModel(model)
		}
		if err := m.ValidateModel(model, ctx); err != nil {
			t.Fatal("unassigned update validated", adapted, err)
		}
		ctx.Scenario = verrors.ScenarioInsert
		if err := m.ValidateModel(model, ctx); err == nil {
			t.Fatal("insert skipped invalid value", adapted)
		}
	}
}

func TestManagerResetDiscardsUseSpecificState(t *testing.T) {
	cfg := SimpleConfig()
	m := NewValidationManager(cfg).(*validationManagerImpl)
	m.SetScenario(verrors.ScenarioQuery)
	if err := m.DisableLayer(LayerConstraint); err != nil {
		t.Fatal(err)
	}
	if err := m.RegisterCustomConstraint(models.KeyMin, func(any, []string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := m.ValidateField(reviewConfigField(), 0, NewContext(verrors.ScenarioInsert, OperationCreate, nil, "")); err != nil {
		t.Fatal(err)
	}
	m.Reset()
	if m.defaultScenario != "" || m.GetValidationStats().TotalValidations != 0 {
		t.Fatal("manager retained default scenario or statistics")
	}
	if err := m.ValidateField(reviewConfigField(), 0, NewContext(verrors.ScenarioInsert, OperationCreate, nil, "")); err == nil {
		t.Fatal("manager retained custom rules or disabled layer")
	}
}
