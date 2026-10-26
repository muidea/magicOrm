package postgres

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type postgresUniqueModel struct {
	ID        int64  `orm:"id key snowflake"`
	Namespace string `orm:"namespace"`
	AccountID int64  `orm:"accountId"`
}

func TestBuildCreateTableIncludesDeclaredUniqueConstraint(t *testing.T) {
	object, err := helper.GetObject(&postgresUniqueModel{})
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
	if !strings.Contains(result.SQL(), `CONSTRAINT "cas_PostgresUniqueModel_uq_namespace_account_2c105558cc1e0e33" UNIQUE ("namespace", "accountId")`) {
		t.Fatalf("unique constraint missing from SQL: %s", result.SQL())
	}
}

func TestBuildCreateUniqueConstraintIsRetrySafe(t *testing.T) {
	object, err := helper.GetObject(&postgresUniqueModel{})
	if err != nil {
		t.Fatal(err)
	}
	remoteProvider := provider.NewRemoteProvider("test", nil)
	model, err := remoteProvider.RegisterModel(object)
	if err != nil {
		t.Fatal(err)
	}
	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "cas"))
	result, buildErr := builder.BuildCreateUniqueConstraint(model, models.UniqueConstraint{
		Name:   "uq_namespace_account",
		Fields: []string{"namespace", "accountId"},
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	if !strings.Contains(result.SQL(), "DO $$") ||
		!strings.Contains(result.SQL(), `conrelid = '"cas_PostgresUniqueModel"'::regclass`) ||
		!strings.Contains(result.SQL(), `ADD CONSTRAINT "cas_PostgresUniqueModel_uq_namespace_account_2c105558cc1e0e33" UNIQUE ("namespace", "accountId")`) {
		t.Fatalf("retry-safe unique constraint SQL missing: %s", result.SQL())
	}
}

func TestPhysicalSchemaObjectNamesAreBoundedAndInstanceScoped(t *testing.T) {
	first := schemaObjectName("instance_one_Record", "uq_target_stamp")
	second := schemaObjectName("instance_two_Record", "uq_target_stamp")
	if first == second {
		t.Fatal("instance namespaces collided")
	}
	if schemaObjectName("A_B", "C") == schemaObjectName("A", "B_C") {
		t.Fatal("table/declaration tuple boundary collided")
	}
	long := strings.Repeat("a", 80)
	one, two := schemaObjectName(long, "first"), schemaObjectName(long, "second")
	if len(one) != 63 || len(two) != 63 || one == two || one != schemaObjectName(long, "first") {
		t.Fatal("long catalog name is unstable, truncated or colliding")
	}
}
