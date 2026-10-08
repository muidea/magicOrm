package provider_test

import (
	"errors"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/utils"
	"github.com/muidea/magicOrm/validation"
	ve "github.com/muidea/magicOrm/validation/errors"
)

type ValidationContract struct {
	ID       int64  `orm:"id key auto" view:"detail,lite" constraint:"req,ro"`
	Required string `orm:"required" view:"detail,lite" constraint:"req"`
	Bounded  int32  `orm:"bounded" view:"detail,lite" constraint:"min=1,max=10"`
	Ranged   int32  `orm:"ranged" view:"detail,lite" constraint:"range=1:10"`
	Choice   string `orm:"choice" view:"detail,lite" constraint:"in=A:B"`
	Pattern  string `orm:"pattern" view:"detail,lite" constraint:"re=^[a-z]+$"`
}

func TestProviderAndScenarioValidationErrorContract(t *testing.T) {
	for _, remote := range []bool{false, true} {
		name := "local"
		p := provider.NewLocalProvider("validation-contract", utils.NewValueValidator())
		var definition any = &ValidationContract{}
		if remote {
			name = "remote"
			p = provider.NewRemoteProvider("validation-contract", utils.NewValueValidator())
			object, err := helper.GetObject(definition)
			if err != nil {
				t.Fatal(err)
			}
			definition = object
		}
		if _, err := p.RegisterModel(definition); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			field, key, reason string
			value              any
		}{
			{"required", "req", "required", ""},
			{"bounded", "min", "too small/short", int32(0)},
			{"bounded", "max", "too large/long", int32(11)},
			{"ranged", "range", "out of range", int32(0)},
			{"ranged", "range", "out of range", int32(11)},
			{"choice", "in", "must be one of", "C"},
			{"pattern", "re", "invalid format", "123"},
		} {
			t.Run(name+"/"+tc.key, func(t *testing.T) {
				valid := &ValidationContract{ID: 1, Required: "ok", Bounded: 5, Ranged: 5, Choice: "A", Pattern: "valid"}
				var input any = valid
				if remote {
					value, err := helper.GetObjectValue(valid)
					if err != nil {
						t.Fatal(err)
					}
					input = value
				}
				m, err := p.GetEntityModel(input, true)
				if err != nil {
					t.Fatal(err)
				}
				err = m.SetFieldValue(tc.field, tc.value)
				assertValidationError(t, err, tc.field, tc.key, tc.reason)
				// Scenario validation reads a model assembled with validators
				// disabled, as the ordinary update path does.
				field := m.GetField(tc.field)
				if err := field.GetValue().Set(tc.value); err != nil {
					t.Fatal(err)
				}
				for _, scenario := range []ve.Scenario{ve.ScenarioInsert, ve.ScenarioUpdate} {
					cfg := validation.DefaultConfig()
					cfg.EnableCaching = false
					manager := validation.NewValidationManager(cfg)
					e := manager.ValidateModel(m, validation.NewContext(scenario, validation.OperationCreate, nil, ""))
					var rich *cd.Error
					if !errors.As(e, &rich) {
						t.Fatalf("missing validation error: %v", e)
					}
					assertValidationError(t, rich, tc.field, tc.key, tc.reason)
				}
			})
		}
	}
}

func assertValidationError(t *testing.T, err *cd.Error, field, key, reason string) {
	t.Helper()
	if err == nil || err.Code != cd.IllegalParam {
		t.Fatalf("wrong invalid-input error: %v", err)
	}
	for _, part := range []string{"field '" + field + "'", "constraint '" + key + "'", reason} {
		if !strings.Contains(err.Message, part) {
			t.Fatalf("missing %q: %v", part, err)
		}
	}
}
