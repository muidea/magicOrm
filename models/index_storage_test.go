package models_test

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestIndexValidatesPhysicalStorageWithoutChangingRelations(t *testing.T) {
	for _, kind := range []string{"scalar", "inline", "reference", "object", "collection", "mixed"} {
		t.Run(kind, func(t *testing.T) {
			ref := &remote.TypeImpl{Name: "Target", PkgPath: "/example", IsPtr: kind != "object"}
			if kind == "collection" {
				ref = &remote.TypeImpl{Name: "slice", ElemType: ref}
			}
			object := &remote.Object{Name: "Record", PkgPath: "/example", Fields: []*remote.Field{
				{Name: "id", Type: &remote.TypeImpl{Name: "int64"}, Spec: &remote.SpecImpl{PrimaryKey: true}},
				{Name: "name", Type: &remote.TypeImpl{Name: "string"}, Spec: &remote.SpecImpl{}},
				{Name: "target", Type: ref, Spec: &remote.SpecImpl{}},
			}}
			fields := []string{"target"}
			if kind == "scalar" {
				fields = []string{"name", "id"}
			}
			if kind == "inline" {
				object.UniqueConstraints = []models.UniqueConstraint{{Name: "uq_target", Fields: []string{"target", "name"}}}
			}
			if kind == "mixed" {
				fields = []string{"name", "target"}
			}
			object.Indexes = []models.Index{{Name: "idx_lookup", Fields: fields}}
			if err := object.Indexes[0].Verify(object.GetFields()); err != nil {
				t.Fatal("legacy declaration inspection must remain available", err)
			}
			err := models.ValidateIndexes(object)
			allowed := kind == "scalar" || kind == "inline"
			if (err == nil) != allowed {
				t.Fatalf("storage=%s error=%v", kind, err)
			}
			if !allowed && (!strings.Contains(err.Message, "idx_lookup") || !strings.Contains(err.Message, "target")) {
				t.Fatal("missing safe index/field location", err)
			}
			if models.IsInlineReference(object, object.GetField("target")) != (kind == "inline") {
				t.Fatal("index changed relation storage")
			}
		})
	}
}
