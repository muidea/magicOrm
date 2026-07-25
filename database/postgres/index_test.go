package postgres

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

func TestBuildCreateTableIncludesDeclaredIndex(t *testing.T) {
	object, err := helper.GetObject(&postgresUniqueModel{})
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
	if !strings.Contains(result.SQL(), `CREATE INDEX IF NOT EXISTS "idx_namespace_account" ON "cas_PostgresUniqueModel" ("namespace", "accountId")`) {
		t.Fatalf("index missing from SQL: %s", result.SQL())
	}
}
