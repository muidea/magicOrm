package mysql

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type mysqlUniqueModel struct {
	ID        int64  `orm:"id key snowflake"`
	Namespace string `orm:"namespace"`
	AccountID int64  `orm:"accountId"`
}

func TestBuildCreateTableIncludesDeclaredUniqueConstraint(t *testing.T) {
	object, err := helper.GetObject(&mysqlUniqueModel{})
	if err != nil {
		t.Fatal(err)
	}
	object.UniqueConstraints = []models.UniqueConstraint{{Name: "uq_namespace_account", Fields: []string{"namespace", "accountId"}}}
	remoteProvider := provider.NewRemoteProvider("test", nil)
	model, err := remoteProvider.RegisterModel(object)
	if err != nil {
		t.Fatal(err)
	}
	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "cas"))
	result, buildErr := builder.BuildCreateTable(model)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	if !strings.Contains(result.SQL(), "CONSTRAINT `uq_namespace_account` UNIQUE (`namespace`, `accountId`)") {
		t.Fatalf("unique constraint missing from SQL: %s", result.SQL())
	}
}
