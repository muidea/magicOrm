package mysql

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

func TestBuildCreateTableIncludesDeclaredIndex(t *testing.T) {
	object, err := helper.GetObject(&mysqlUniqueModel{})
	if err != nil {
		t.Fatal(err)
	}
	object.Indexes = []models.Index{{Name: "idx_namespace_account", Fields: []string{"namespace", "accountId"}}}
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
	if !strings.Contains(result.SQL(), "INDEX `idx_namespace_account` (`namespace`, `accountId`)") {
		t.Fatalf("index missing from SQL: %s", result.SQL())
	}
}
