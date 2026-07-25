package remote

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestEnablePrivateReadFieldsOnlyAllowsWriteOnlyBasicFields(t *testing.T) {
	objectPtr := &Object{
		Name: "account",
		Fields: []*Field{
			{Name: "passwordHash", Type: &TypeImpl{Value: models.TypeStringValue}, Spec: &SpecImpl{Constraint: "wo"}},
			{Name: "account", Type: &TypeImpl{Value: models.TypeStringValue}, Spec: &SpecImpl{}},
		},
	}
	if err := objectPtr.EnablePrivateReadFields([]string{"passwordHash"}); err != nil {
		t.Fatalf("EnablePrivateReadFields write-only field: %v", err)
	}
	if !objectPtr.AllowsPrivateReadField("passwordHash") {
		t.Fatal("write-only field was not marked for private read")
	}
	if err := objectPtr.EnablePrivateReadFields([]string{"account"}); err == nil {
		t.Fatal("normal field must not be enabled as a private read field")
	}
}
