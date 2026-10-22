package validation_test

import (
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/validation"
	ve "github.com/muidea/magicOrm/validation/errors"
	"testing"
)

type nullableTextModel struct {
	ID       int64   `orm:"id key"`
	Date     *string `orm:"date" constraint:"re=^[0-9]{4}-[0-9]{2}-[0-9]{2}$"`
	State    *string `orm:"state" constraint:"in=open:closed"`
	Required *string `orm:"required" constraint:"req,re=^[A-Z]+$"`
}

func TestNullableTextProviderValidationPreservesNullAndRequired(t *testing.T) {
	good, invalid, day, state := "OK", "bad", "2026-10-14", "open"
	for _, name := range []string{"local", "remote"} {
		for _, sample := range []string{"null", "values", "bad-date", "bad-enum", "missing-required"} {
			t.Run(name+"/"+sample, func(t *testing.T) {
				entity := &nullableTextModel{ID: 1, Required: &good}
				if sample == "values" {
					entity.Date, entity.State = &day, &state
				}
				if sample == "bad-date" {
					entity.Date = &invalid
				}
				if sample == "bad-enum" {
					entity.State = &invalid
				}
				if sample == "missing-required" {
					entity.Required = nil
				}
				model, err := local.GetEntityModel(entity, nil)
				if err != nil {
					t.Fatal(err)
				}
				if name == "remote" {
					object, err := helper.GetObject(entity)
					if err != nil {
						t.Fatal(err)
					}
					model, err = remote.GetEntityModel(object, nil)
					if err != nil {
						t.Fatal(err)
					}
					value, valueErr := helper.GetObjectValue(entity)
					if valueErr != nil {
						t.Fatal(valueErr)
					}
					model, err = remote.SetModelValue(model, remote.NewValue(value), true)
					if err != nil {
						t.Fatal(err)
					}
				}
				manager := validation.NewValidationManager(validation.DefaultConfig())
				ctx := validation.NewContext(ve.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
				if err := manager.ValidateModel(model, ctx); (err == nil) != (sample == "null" || sample == "values") {
					t.Fatal(sample, err)
				}
			})
		}
	}
}
