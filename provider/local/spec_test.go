package local

import "testing"

func TestSpec(t *testing.T) {
	spec1 := "spec"
	specPtr, specErr := getSpec(spec1)
	if specErr != nil {
		t.Errorf("newSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("newSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if specPtr.IsAutoIncrement() || specPtr.IsPrimaryKey() {
		t.Errorf("newSpec failed")
		return
	}

	spec2 := "spec auto"
	specPtr, specErr = getSpec(spec2)
	if specErr != nil {
		t.Errorf("newSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("newSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if !specPtr.IsAutoIncrement() || specPtr.IsPrimaryKey() {
		t.Errorf("newSpec failed")
		return
	}

	spec3 := "spec auto key"
	specPtr, specErr = getSpec(spec3)
	if specErr != nil {
		t.Errorf("newSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("newSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if !specPtr.IsAutoIncrement() || !specPtr.IsPrimaryKey() {
		t.Errorf("newSpec failed")
		return
	}

	spec4 := "spec key auto"
	specPtr, specErr = getSpec(spec4)
	if specErr != nil {
		t.Errorf("newSpec failed, err:%s", specErr.Error())
		return
	}
	if specPtr.GetFieldName() != "spec" {
		t.Errorf("newSpec failed,current:%s, expect:%s", specPtr.GetFieldName(), "spec")
	}
	if !specPtr.IsAutoIncrement() || !specPtr.IsPrimaryKey() {
		t.Errorf("newSpec failed")
		return
	}
}
