package orm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type unregisteredTransient struct{ Trace string }

type ignoredDDLModel struct {
	ID    int64                 `orm:"id key"`
	Trace unregisteredTransient `orm:"-"`
	Name  string                `orm:"name"`
}

func TestSchemaDDLDoesNotResolveIgnoredRelation(t *testing.T) {
	p := provider.NewLocalProvider("ignored-ddl-test", nil)
	model, err := p.RegisterModel(&ignoredDDLModel{ID: 1, Name: "stored"})
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{}
	c := codec.New(p, "ignored")
	if err := NewCreateRunner(context.Background(), model, executor, p, c).Create(); err != nil {
		t.Fatal("ignored relation was resolved by the Create graph", err)
	}
	if len(executor.execCalls) != 1 || !strings.Contains(executor.execCalls[0].sql, "IgnoredDDLModel") {
		t.Fatal("ignored relation generated extra DDL", executor.execCalls)
	}
}

func TestSchemaDDLRejectsInvalidGraphBeforeExecution(t *testing.T) {
	for _, operation := range []string{"create", "drop", "reconcile", "create-runner", "drop-runner", "reconcile-runner"} {
		for _, mode := range []string{"cycle", "collision", "canceled", "nil-provider", "nil-model"} {
			t.Run(operation+"/"+mode, func(t *testing.T) {
				handler, model, executor := inspectionFixture(t)
				switch mode {
				case "cycle", "collision":
					mapped := model
					if mode == "collision" {
						owned, err := handler.modelProvider.GetTypeModel(model.GetField("owned").GetType().Elem())
						if err != nil {
							t.Fatal(err)
						}
						mapped = schemaNamedModel{Model: owned, name: model.GetName()}
					}
					handler.modelProvider = schemaMappedProvider{Provider: handler.modelProvider, mapped: mapped}
				case "canceled":
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					handler.context = ctx
				case "nil-provider":
					handler.modelProvider = nil
				case "nil-model":
					model = nil
				}
				var err *cd.Error
				switch operation {
				case "create":
					err = handler.Create(model)
				case "drop":
					err = handler.Drop(model)
				case "reconcile":
					err = handler.Reconcile(model, model)
				case "create-runner":
					err = NewCreateRunner(handler.context, model, executor, handler.modelProvider, handler.modelCodec).Create()
				case "drop-runner":
					err = NewDropRunner(handler.context, model, executor, handler.modelProvider, handler.modelCodec).Drop()
				case "reconcile-runner":
					err = NewReconcileRunner(handler.context, model, model, executor, handler.modelProvider, handler.modelCodec).Reconcile()
				}
				if err == nil || len(executor.execCalls) != 0 || len(executor.calls) != 0 {
					t.Fatalf("invalid declaration must fail without SQL: err=%v calls=%v", err, executor.execCalls)
				}
			})
		}
	}
}

type schemaFieldsModel struct {
	models.Model
	fields models.Fields
}

func (s schemaFieldsModel) GetFields() models.Fields { return s.fields }

type schemaNamedField struct {
	models.Field
	name string
}

func (s schemaNamedField) GetName() string { return s.name }

func TestSchemaTableLimitBeforeDDL(t *testing.T) {
	for _, operation := range []string{"validate", "create", "drop", "reconcile"} {
		for _, count := range []int{maxSchemaTables - 2, maxSchemaTables - 1} {
			t.Run(fmt.Sprintf("%s/%d", operation, count), func(t *testing.T) {
				handler, model, executor := inspectionFixture(t)
				fields := models.Fields{model.GetPrimaryField()}
				for i := 0; i < count; i++ {
					fields = append(fields, schemaNamedField{Field: model.GetField("owned"), name: fmt.Sprintf("owned%d", i)})
				}
				model = schemaFieldsModel{Model: model, fields: fields}
				var err *cd.Error
				switch operation {
				case "validate":
					err = ValidateSchema(handler.context, model, handler.modelProvider, "tenant_a")
				case "create":
					err = handler.Create(model)
				case "drop":
					err = handler.Drop(model)
				case "reconcile":
					err = handler.Reconcile(model, model)
				}
				if count == maxSchemaTables-1 {
					if err == nil || !strings.Contains(err.Error(), "table limit") || len(executor.execCalls) != 0 {
						t.Fatalf("oversized graph executed: %v calls=%d", err, len(executor.execCalls))
					}
				} else if err != nil {
					t.Fatal(err)
				} else if operation == "create" || operation == "drop" {
					if len(executor.execCalls) != maxSchemaTables {
						t.Fatalf("shared owned host must execute once: %d", len(executor.execCalls))
					}
				}
			})
		}
	}
}

type failingSchemaBuilder struct{ database.Builder }

func (s failingSchemaBuilder) BuildCreateRelationTable(models.Model, models.Field) (database.Result, *cd.Error) {
	return nil, cd.NewError(cd.IllegalParam, "late relation builder failure")
}
func (s failingSchemaBuilder) BuildDropRelationTable(models.Model, models.Field) (database.Result, *cd.Error) {
	return nil, cd.NewError(cd.IllegalParam, "late relation builder failure")
}

func TestSchemaDDLBuildsAllStatementsBeforeExecution(t *testing.T) {
	for _, operation := range []string{"create", "drop"} {
		t.Run(operation, func(t *testing.T) {
			handler, model, executor := inspectionFixture(t)
			var err *cd.Error
			if operation == "create" {
				runner := NewCreateRunner(handler.context, model, executor, handler.modelProvider, handler.modelCodec)
				runner.sqlBuilder = failingSchemaBuilder{runner.sqlBuilder}
				err = runner.Create()
			} else {
				runner := NewDropRunner(handler.context, model, executor, handler.modelProvider, handler.modelCodec)
				runner.sqlBuilder = failingSchemaBuilder{runner.sqlBuilder}
				err = runner.Drop()
			}
			if err == nil || len(executor.execCalls) != 0 {
				t.Fatal(err, executor.execCalls)
			}
		})
	}
}

type reconcileLateRequired struct {
	ID        int64   `orm:"id key snowflake"`
	Namespace *string `orm:"namespace"`
	Optional  *string `orm:"optional"`
	Required  string  `orm:"required"`
}

func TestReconcileRejectsLateRequiredColumnBeforeAnyDDL(t *testing.T) {
	p, previous, current := reconcileModels(t, &reconcileLateRequired{})
	executor := &fakeExecutor{}
	runner := NewReconcileRunner(context.Background(), previous, current, executor, p, codec.New(p, "tenant_a"))
	if err := runner.Reconcile(); err == nil || len(executor.execCalls) != 0 {
		t.Fatal("partial migration on invalid declaration", err, executor.execCalls)
	}
}

type reconcileNewOwned struct {
	ID        int64        `orm:"id key snowflake"`
	Namespace *string      `orm:"namespace"`
	Owned     inspectOwned `orm:"owned"`
}

func TestReconcileCreatesNewOwnedHostBeforeRelation(t *testing.T) {
	p, previous, current := reconcileModels(t, &reconcileNewOwned{})
	object, err := helper.GetObject(&inspectOwned{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.RegisterModel(object); err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{}
	runner := NewReconcileRunner(context.Background(), previous, current, executor, p, codec.New(p, "tenant_a"))
	if err := runner.Reconcile(); err != nil {
		t.Fatal(err)
	}
	relation, err := codec.New(p, "tenant_a").ConstructRelationTableName(current, current.GetField("owned"))
	if err != nil {
		t.Fatal(err)
	}
	if len(executor.execCalls) != 2 || !strings.Contains(executor.execCalls[0].sql, "tenant_a_InspectOwned") ||
		!strings.Contains(executor.execCalls[1].sql, relation) {
		t.Fatal(executor.execCalls)
	}
}

type cancelSchemaExecutor struct {
	fakeExecutor
	cancel context.CancelFunc
}

func (s *cancelSchemaExecutor) Execute(sql string, args ...any) (int64, *cd.Error) {
	rows, err := s.fakeExecutor.Execute(sql, args...)
	s.cancel()
	return rows, err
}

func TestSchemaDDLStopsAtCancellationBetweenStatements(t *testing.T) {
	for _, operation := range []string{"create", "drop", "reconcile"} {
		t.Run(operation, func(t *testing.T) {
			handler, model, _ := inspectionFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			executor := &cancelSchemaExecutor{cancel: cancel}
			handler.context, handler.executor = ctx, executor
			var err *cd.Error
			switch operation {
			case "create":
				err = handler.Create(model)
			case "drop":
				err = handler.Drop(model)
			case "reconcile":
				p, previous, desired := reconcileModels(t, &reconcileCurrentModel{})
				err = NewReconcileRunner(ctx, previous, desired, executor, p, codec.New(p, "tenant_a")).Reconcile()
			}
			if err == nil || len(executor.execCalls) != 1 {
				t.Fatal("canceled DDL continued or reported completion", err, executor.execCalls)
			}
		})
	}
}
