package remote

import "testing"

func TestSpec(t *testing.T) {
	spec := ""
	_, err := getSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}

	spec = "test"
	itemSpec, err := getSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}
	if itemSpec.GetFieldName() != "test" {
		t.Errorf("illegal spec name")
		return
	}
	if itemSpec.IsPrimaryKey() {
		t.Errorf("illegal spec define")
		return
	}
	if itemSpec.IsAutoIncrement() {
		t.Errorf("illegal spec define")
		return
	}

	spec = "test auto key"
	itemSpec, err = getSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}
	if itemSpec.GetFieldName() != "test" {
		t.Errorf("illegal spec name")
		return
	}
	if !itemSpec.IsPrimaryKey() {
		t.Errorf("illegal spec define")
		return
	}
	if !itemSpec.IsAutoIncrement() {
		t.Errorf("illegal spec define")
		return
	}
}
