package helper

import (
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/remote"
)

type maskRole struct {
	ID        int64  `json:"id" orm:"id key snowflake" view:"detail,lite"`
	Name      string `json:"name" orm:"name" view:"detail,lite"`
	Status    int    `json:"status" orm:"status" view:"detail,lite"`
	Privilege string `json:"privilege" orm:"privilege" view:"detail"`
}

type maskAccount struct {
	ID        int64     `json:"id" orm:"id key snowflake" view:"detail,lite"`
	Account   string    `json:"account" orm:"account" view:"detail,lite"`
	Status    int       `json:"status" orm:"status" view:"detail,lite"`
	Namespace string    `json:"namespace" orm:"namespace" view:"detail,lite"`
	Role      *maskRole `json:"role" orm:"role" view:"detail"`
}

type maskAccountWithRelationView struct {
	ID   int64     `json:"id" orm:"id key snowflake" view:"detail,lite"`
	Role *maskRole `json:"role" orm:"role" view:"detail" relationView:"detail"`
}

func TestBuildFieldMask(t *testing.T) {
	mask, err := buildFieldMask(&maskAccount{Role: &maskRole{}},
		"id", "account", "status", "namespace",
		"role.id", "role.name", "role.status",
	)
	if err != nil {
		t.Fatalf("buildFieldMask failed: %v", err)
	}
	if mask == nil {
		t.Fatal("buildFieldMask returned nil")
	}

	if len(mask.Fields) != 5 {
		t.Fatalf("unexpected top-level field count: %d", len(mask.Fields))
	}

	var roleValFound bool
	for _, field := range mask.Fields {
		if field.Name != "role" {
			continue
		}
		roleValFound = true
		roleVal, ok := field.Value.(*remote.ObjectValue)
		if !ok {
			rawRole, ok := field.Value.(remote.ObjectValue)
			if !ok {
				t.Fatalf("unexpected role field type: %T", field.Value)
			}
			roleVal = &rawRole
		}
		if len(roleVal.Fields) != 3 {
			t.Fatalf("unexpected role field count: %d", len(roleVal.Fields))
		}
		for _, roleField := range roleVal.Fields {
			switch roleField.Name {
			case "id", "name", "status":
			default:
				t.Fatalf("unexpected role field: %s", roleField.Name)
			}
		}
	}
	if !roleValFound {
		t.Fatal("role field not found in mask")
	}
}

func TestBuildViewMaskUsesLiteRelationViewByDefault(t *testing.T) {
	mask, err := BuildViewMask(&maskAccount{}, models.DetailView)
	if err != nil {
		t.Fatalf("BuildViewMask failed: %v", err)
	}
	if mask == nil {
		t.Fatal("BuildViewMask returned nil")
	}

	var roleValFound bool
	for _, field := range mask.Fields {
		if field.Name != "role" {
			continue
		}
		roleValFound = true
		roleVal, ok := field.Value.(*remote.ObjectValue)
		if !ok {
			rawRole, ok := field.Value.(remote.ObjectValue)
			if !ok {
				t.Fatalf("unexpected role field type: %T", field.Value)
			}
			roleVal = &rawRole
		}
		for _, roleField := range roleVal.Fields {
			switch roleField.Name {
			case "id", "name", "status":
			default:
				t.Fatalf("unexpected role field in lite relation view: %s", roleField.Name)
			}
		}
	}
	if !roleValFound {
		t.Fatal("role field not found in view mask")
	}
}

func TestBuildViewMaskWithTopLevelUntaggedFields(t *testing.T) {
	type internalAccount struct {
		ID           int64     `json:"id" orm:"id key snowflake" view:"detail,lite"`
		Account      string    `json:"account" orm:"account" view:"detail,lite"`
		PasswordHash string    `json:"passwordHash" orm:"passwordHash"`
		Namespace    string    `json:"namespace" orm:"namespace"`
		Role         *maskRole `json:"role" orm:"role" view:"detail"`
	}

	mask, err := buildViewMaskWithOptions(&internalAccount{}, models.DetailView, &viewMaskOptions{
		IncludeUntaggedTopLevel: true,
	})
	if err != nil {
		t.Fatalf("buildViewMaskWithOptions failed: %v", err)
	}

	if mask.GetFieldValue("passwordHash") == nil {
		t.Fatal("passwordHash should be kept as untagged top-level field")
	}
	if mask.GetFieldValue("namespace") == nil {
		t.Fatal("namespace should be kept as untagged top-level field")
	}

	roleField := mask.GetFieldValue("role")
	roleValue, ok := roleField.(*remote.ObjectValue)
	if !ok {
		rawRole, ok := roleField.(remote.ObjectValue)
		if !ok {
			t.Fatalf("unexpected role field type: %T", roleField)
		}
		roleValue = &rawRole
	}
	for _, field := range roleValue.Fields {
		switch field.Name {
		case "id", "name", "status":
		default:
			t.Fatalf("unexpected relation field in lite relation view: %s", field.Name)
		}
	}
}

func TestBuildViewMaskIgnoresRelationViewTag(t *testing.T) {
	mask, err := BuildViewMask(&maskAccountWithRelationView{}, models.DetailView)
	if err != nil {
		t.Fatalf("BuildViewMask failed: %v", err)
	}

	roleField := mask.GetFieldValue("role")
	roleValue, ok := roleField.(*remote.ObjectValue)
	if !ok {
		rawRole, ok := roleField.(remote.ObjectValue)
		if !ok {
			t.Fatalf("unexpected role field type: %T", roleField)
		}
		roleValue = &rawRole
	}

	for _, field := range roleValue.Fields {
		switch field.Name {
		case "id", "name", "status":
		default:
			t.Fatalf("relationView tag should not expand child view, got field %s", field.Name)
		}
	}
}
