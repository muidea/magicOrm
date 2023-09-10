package local

import (
	"testing"

	"github.com/muidea/magicOrm/model"
)

func TestSpec(t *testing.T) {
	spec1 := "spec"
	specPtr, specErr := getSpec(spec1)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() == model.AutoIncrement || specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}

	spec2 := "spec auto"
	specPtr, specErr = getSpec(spec2)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() != model.AutoIncrement || specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}

	spec3 := "spec auto key"
	specPtr, specErr = getSpec(spec3)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() != model.AutoIncrement || !specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}

	spec4 := "spec key auto"
	specPtr, specErr = getSpec(spec4)
	if specErr != nil {
		t.Errorf("NewSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("NewSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.GetValueDeclare() != model.AutoIncrement || !specPtr.IsPrimaryKey() {
		t.Errorf("NewSpec failed")
		return
	}
}
