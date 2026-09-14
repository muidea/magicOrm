package orm

import (
	"context"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
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
