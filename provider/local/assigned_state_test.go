package local

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestMetaModelSetFieldValuePreservesAssignedNilAndZero(t *testing.T) {
	entity := assignedPatch{
		ID:    1,
		Count: 9,
		Child: &assignedChild{ID: 2, Name: "child"},
		Children: []*assignedChild{
			{ID: 3, Name: "child-1"},
		},
	}

	model, err := GetEntityModel(&entity, nil)
	if err != nil {
		t.Fatalf("GetEntityModel failed: %v", err)
	}
	metaModel := model.Copy(models.MetaView)

	if models.IsAssignedField(metaModel.GetField("count")) {
		t.Fatal("meta count should start unassigned")
	}
	if models.IsAssignedField(metaModel.GetField("child")) {
		t.Fatal("meta child should start unassigned")
	}
	if models.IsAssignedField(metaModel.GetField("children")) {
		t.Fatal("meta children should start unassigned")
	}

	if err := metaModel.SetFieldValue("count", 0); err != nil {
		t.Fatalf("SetFieldValue(count) failed: %v", err)
	}
	var clearedChild *assignedChild
	if err := metaModel.SetFieldValue("child", clearedChild); err != nil {
		t.Fatalf("SetFieldValue(child) failed: %v", err)
	}
	var clearedChildren []*assignedChild
	if err := metaModel.SetFieldValue("children", clearedChildren); err != nil {
		t.Fatalf("SetFieldValue(children) failed: %v", err)
	}

	if !models.IsAssignedField(metaModel.GetField("count")) {
		t.Fatal("explicit zero count should be assigned")
	}
	if !models.IsAssignedField(metaModel.GetField("child")) {
		t.Fatal("explicit nil child should be assigned")
	}
	if !models.IsAssignedField(metaModel.GetField("children")) {
		t.Fatal("explicit nil children should be assigned")
	}
	if models.IsValidField(metaModel.GetField("child")) {
		t.Fatal("explicit nil child should remain invalid")
	}
	if models.IsValidField(metaModel.GetField("children")) {
		t.Fatal("explicit nil children should remain invalid")
	}
	if childValue := metaModel.GetField("child").GetValue().Get(); reflect.ValueOf(childValue).Kind() != reflect.Ptr || !reflect.ValueOf(childValue).IsNil() {
		t.Fatal("explicit nil child should export a typed nil pointer")
	}
	if childrenValue := metaModel.GetField("children").GetValue().Get(); reflect.ValueOf(childrenValue).Kind() != reflect.Slice || !reflect.ValueOf(childrenValue).IsNil() {
		t.Fatal("explicit nil children should export a typed nil slice")
	}
}
