package remote

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestObjectVerifyUniqueConstraintRequiresDeclaredBasicFields(t *testing.T) {
	object := &Object{
		Name:    "LoginState",
		PkgPath: "/cas",
		Fields: []*Field{
			{Name: "id", Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, Spec: &SpecImpl{FieldName: "id", PrimaryKey: true, ValueDeclare: models.Snowflake}},
			{Name: "namespace", Type: &TypeImpl{Name: "string", Value: models.TypeStringValue}, Spec: &SpecImpl{FieldName: "namespace"}},
			{Name: "accountId", Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, Spec: &SpecImpl{FieldName: "accountId"}},
		},
		UniqueConstraints: []models.UniqueConstraint{{Name: "uq_login_state_namespace_account", Fields: []string{"namespace", "accountId"}}},
	}
	if err := object.Verify(); err != nil {
		t.Fatalf("valid unique constraint rejected: %v", err)
	}
	object.UniqueConstraints[0].Fields = []string{"unknown"}
	if err := object.Verify(); err == nil {
		t.Fatal("constraint referencing an unknown field must be rejected")
	}
}
