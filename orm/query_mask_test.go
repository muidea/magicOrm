package orm

import (
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type queryMaskReference struct {
	ID          int       `orm:"id key auto"`
	Name        string    `orm:"name"`
	FValue      float32   `orm:"value"`
	Flag        bool      `orm:"flag"`
	StrArray    []string  `orm:"strArray"`
	PtrArray    *[]string `orm:"ptrArray"`
	PtrStrArray *[]string `orm:"ptrStrArray"`
}

func TestBuildFullQueryMaskModelExpandsRemoteBasicFields(t *testing.T) {
	remoteProvider := provider.NewRemoteProvider("tenant", nil)

	definition, err := helper.GetObject(&queryMaskReference{})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if _, err := remoteProvider.RegisterModel(definition); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	queryValue := &remote.ObjectValue{
		Name:    "queryMaskReference",
		PkgPath: "github.com/muidea/magicOrm/orm",
		Fields: []*remote.FieldValue{
			{Name: "id", Value: 1, Assigned: true},
		},
	}

	queryModel, err := remoteProvider.GetEntityModel(queryValue, true)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}

	maskModel, err := buildFullQueryMaskModel(queryModel)
	if err != nil {
		t.Fatalf("buildFullQueryMaskModel failed: %v", err)
	}

	for _, fieldName := range []string{"name", "value", "flag", "strArray"} {
		field := maskModel.GetField(fieldName)
		if field == nil {
			t.Fatalf("mask field %q should exist", fieldName)
		}
		if !models.IsValidField(field) {
			t.Fatalf("mask field %q should be valid for full-row query", fieldName)
		}
	}

	for _, fieldName := range []string{"ptrArray", "ptrStrArray"} {
		field := maskModel.GetField(fieldName)
		if field == nil {
			t.Fatalf("mask field %q should exist", fieldName)
		}
		if models.IsValidField(field) {
			t.Fatalf("pointer field %q should keep query mask semantics", fieldName)
		}
	}
}
