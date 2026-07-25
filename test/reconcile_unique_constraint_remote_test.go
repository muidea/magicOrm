package test

import (
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type reconcileUniqueConstraintEntity struct {
	ID        int64  `orm:"id key snowflake"`
	Namespace string `orm:"namespace"`
	AccountID int64  `orm:"accountId"`
}

// TestRemoteReconcileUniqueConstraintRetry covers the window where the
// database DDL succeeded but a caller retries before persisting its model
// declaration. The second reconcile must be harmless.
func TestRemoteReconcileUniqueConstraintRetry(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	previousObject, err := helper.GetObject(&reconcileUniqueConstraintEntity{})
	if err != nil {
		t.Fatal(err)
	}
	previousProvider := provider.NewRemoteProvider("reconcile-previous", nil)
	previousModel, err := previousProvider.RegisterModel(previousObject)
	if err != nil {
		t.Fatal(err)
	}

	currentObject, err := helper.GetObject(&reconcileUniqueConstraintEntity{})
	if err != nil {
		t.Fatal(err)
	}
	currentObject.UniqueConstraints = []models.UniqueConstraint{{
		Name:   "uq_reconcile_namespace_account",
		Fields: []string{"namespace", "accountId"},
	}}
	currentProvider := provider.NewRemoteProvider("reconcile-current", nil)
	currentModel, err := currentProvider.RegisterModel(currentObject)
	if err != nil {
		t.Fatal(err)
	}

	o, err := orm.NewOrm(currentProvider, config, "reconcile_unique_constraint")
	if err != nil {
		t.Fatal(err)
	}
	defer o.Release()
	defer func() {
		if dropErr := o.Drop(currentModel); dropErr != nil {
			t.Errorf("drop model failed: %s", dropErr.Error())
		}
	}()

	if err = o.Drop(currentModel); err != nil {
		t.Fatal(err)
	}
	if err = o.Create(previousModel); err != nil {
		t.Fatal(err)
	}
	if err = o.Reconcile(previousModel, currentModel); err != nil {
		t.Fatalf("initial reconcile failed: %s", err.Error())
	}
	if err = o.Reconcile(previousModel, currentModel); err != nil {
		t.Fatalf("retry reconcile failed: %s", err.Error())
	}
}
