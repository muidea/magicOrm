package orm

import (
	"context"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type reconcileOldModel struct {
	ID        int64   `orm:"id key snowflake"`
	Namespace *string `orm:"namespace"`
}

type reconcileCurrentModel struct {
	ID        int64   `orm:"id key snowflake"`
	Namespace *string `orm:"namespace"`
	Audience  *string `orm:"audience"`
}

type reconcileDestructiveModel struct {
	ID int64 `orm:"id key snowflake"`
}

func reconcileModels(t *testing.T, current any) (provider.Provider, models.Model, models.Model) {
	t.Helper()
	previousObject, err := helper.GetObject(&reconcileOldModel{})
	if err != nil {
		t.Fatal(err)
	}
	currentObject, err := helper.GetObject(current)
	if err != nil {
		t.Fatal(err)
	}
	// The two Go test types intentionally model different revisions of one
	// persisted dynamic entity.
	currentObject.Name = previousObject.Name
	currentObject.PkgPath = previousObject.PkgPath

	remoteProvider := provider.NewRemoteProvider("reconcile-test", nil)
	previous, err := remoteProvider.RegisterModel(previousObject)
	if err != nil {
		t.Fatal(err)
	}
	currentModel, err := remoteProvider.RegisterModel(currentObject)
	if err != nil {
		t.Fatal(err)
	}
	return remoteProvider, previous, currentModel
}

func TestReconcileRunnerAddsNullableColumn(t *testing.T) {
	remoteProvider, previous, current := reconcileModels(t, &reconcileCurrentModel{})
	executor := &fakeExecutor{}
	runner := NewReconcileRunner(context.Background(), previous, current, executor, remoteProvider, codec.New(remoteProvider, "cas"))
	if err := runner.Reconcile(); err != nil {
		t.Fatal(err)
	}
	if len(executor.execCalls) != 1 || !strings.Contains(executor.execCalls[0].sql, "ADD COLUMN") || !strings.Contains(executor.execCalls[0].sql, "audience") {
		t.Fatalf("expected one additive column statement, got %#v", executor.execCalls)
	}
}

func TestReconcileRunnerRejectsDestructiveChangesBeforeDDL(t *testing.T) {
	remoteProvider, previous, current := reconcileModels(t, &reconcileDestructiveModel{})
	executor := &fakeExecutor{}
	runner := NewReconcileRunner(context.Background(), previous, current, executor, remoteProvider, codec.New(remoteProvider, "cas"))
	if err := runner.Reconcile(); err == nil || !strings.Contains(err.Error(), "schema migration required") {
		t.Fatalf("expected destructive migration error, got %v", err)
	}
	if len(executor.execCalls) != 0 {
		t.Fatalf("destructive update must not issue DDL: %#v", executor.execCalls)
	}
}

func TestReconcileDefaultChangesComparePhysicalColumns(t *testing.T) {
	for _, tt := range []struct {
		name, typeName string
		before, after  any
		migration      bool
	}{
		{name: "string runtime reference removed", typeName: "string", before: "$referenceExtData.entity.id"},
		{name: "integer runtime reference removed", typeName: "int64", before: "$referenceExtData.timeStamp"},
		{name: "stored numeric default changed", typeName: "int64", before: int64(1), after: int64(2), migration: true},
		{name: "stored numeric default removed", typeName: "int64", before: int64(1), migration: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			old := &remote.Object{Name: "record", PkgPath: "test", Fields: []*remote.Field{
				{Name: "id", Type: &remote.TypeImpl{Name: "int64"}, Spec: &remote.SpecImpl{PrimaryKey: true, ValueDeclare: "snowflake"}},
				{Name: "value", Type: &remote.TypeImpl{Name: tt.typeName}, Spec: &remote.SpecImpl{DefaultValue: tt.before}},
			}}
			next := old.Copy("test").(*remote.Object)
			next.Fields[1].Spec.DefaultValue = tt.after
			p := provider.NewRemoteProvider("default-test", nil)
			before, err := p.RegisterModel(old)
			if err != nil {
				t.Fatal(err)
			}
			after, err := p.RegisterModel(next)
			if err != nil {
				t.Fatal(err)
			}
			executor := &fakeExecutor{}
			runner := NewReconcileRunner(context.Background(), before, after, executor, p, codec.New(p, "test"))
			err = runner.Reconcile()
			if tt.migration {
				if err == nil || !strings.Contains(err.Error(), "schema migration required") {
					t.Fatalf("expected migration rejection, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if len(executor.execCalls) != 0 {
				t.Fatalf("default comparison executed DDL: %#v", executor.execCalls)
			}
		})
	}
}
