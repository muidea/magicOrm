package test

import (
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type reconcileUniqueConstraintEntity struct {
	ID int64 `orm:"id key snowflake"`
	// Use indexable identities for both backends; ordinary strings map to TEXT
	// in MySQL and cannot participate in a full-column unique constraint.
	Namespace int64 `orm:"namespace"`
	AccountID int64 `orm:"accountId"`
}

// TestRemoteReconcileUniqueConstraintRetry covers the window where the
// database DDL succeeded but a caller retries before persisting its model
// declaration. PostgreSQL guards repeated constraint DDL; basic MySQL
// Reconcile fails closed and requires controlled physical schema recovery.
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
	err = o.Reconcile(previousModel, currentModel)
	if _, mysqlBackend := config.(*mysql.Config); mysqlBackend {
		if err == nil || err.Code != cd.Unexpected || !strings.Contains(err.Error(), "1061") {
			t.Fatalf("MySQL retry must report duplicate constraint DDL, got %v", err)
		}
	} else if err != nil {
		t.Fatalf("PostgreSQL retry reconcile failed: %s", err.Error())
	}

	// Verify the physical constraint survives both retry outcomes. Different
	// primary keys must not allow duplicate compound identities.
	for i := int64(1); i <= 2; i++ {
		value := currentModel.Copy(models.OriginView)
		for field, fieldValue := range map[string]int64{"id": i, "namespace": 7, "accountId": 11} {
			if err = value.SetFieldValue(field, fieldValue); err != nil {
				t.Fatal(err)
			}
		}
		_, err = o.Insert(value)
		if i == 1 && err != nil {
			t.Fatalf("initial insert failed: %s", err.Error())
		}
		if i == 2 && (err == nil || err.Code != cd.Duplicated) {
			t.Fatalf("compound uniqueness must reject duplicate identity, got %v", err)
		}
	}
}
