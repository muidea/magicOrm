package validation_test

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

type generatedPrimaryModel struct {
	ID   uint64 `orm:"id key auto" constraint:"req,ro"`
	Name string `orm:"name" constraint:"req"`
}

func TestInsertValidationDefersMissingAutoIncrementValue(t *testing.T) {
	for _, provider := range []string{"local", "remote"} {
		t.Run(provider, func(t *testing.T) {
			entity := &generatedPrimaryModel{Name: "abc"}
			model, err := local.GetEntityModel(entity, nil)
			if err != nil {
				t.Fatal(err)
			}
			if provider == "remote" {
				object, err := helper.GetObject(entity)
				if err != nil {
					t.Fatal(err)
				}
				model, err = remote.GetEntityModel(object, nil)
				if err != nil {
					t.Fatal(err)
				}
				if err := model.SetFieldValue("name", "abc"); err != nil {
					t.Fatal(err)
				}
			}
			manager := validation.NewValidationManager(validation.DefaultConfig())
			ctx := validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
			if err := manager.ValidateModel(model, ctx); err != nil {
				t.Fatalf("generated primary rejected: %v", err)
			}
			if err := model.SetFieldValue("name", ""); err != nil {
				t.Fatal(err)
			}
			ctx = validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
			if err := manager.ValidateModel(model, ctx); err == nil {
				t.Fatal("ordinary required field bypassed")
			}
			ctx = validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
			if err := manager.ValidateField(model.GetPrimaryField(), "invalid", ctx); err == nil {
				t.Fatal("explicit invalid generated identity accepted")
			}
			manual := &remote.Field{Name: "id", Type: &remote.TypeImpl{Name: "uint64"}, Spec: &remote.SpecImpl{PrimaryKey: true, Constraint: "req"}}
			ctx = validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
			if err := manager.ValidateField(manual, uint64(0), ctx); err == nil {
				t.Fatal("zero manual identity accepted")
			}
		})
	}
}
