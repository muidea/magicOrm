package remote

import "testing"

func TestTag(t *testing.T) {
	tag := ""
	_, err := GetItemTag(tag)
	if err == nil {
		t.Errorf("illegal tag value")
		return
	}

	tag = "test"
	itemTag, err := GetItemTag(tag)
	if err == nil {
		t.Errorf("illegal tag value")
		return
	}
	if itemTag.GetName() != "test" {
		t.Errorf("illegal tag name")
		return
	}
	if itemTag.IsPrimaryKey() {
		t.Errorf("illegal tag define")
		return
	}
	if itemTag.IsAutoIncrement() {
		t.Errorf("illegal tag define")
		return
	}

	tag = "test auto key"
	itemTag, err = GetItemTag(tag)
	if err == nil {
		t.Errorf("illegal tag value")
		return
	}
	if itemTag.GetName() != "test" {
		t.Errorf("illegal tag name")
		return
	}
	if !itemTag.IsPrimaryKey() {
		t.Errorf("illegal tag define")
		return
	}
	if !itemTag.IsAutoIncrement() {
		t.Errorf("illegal tag define")
		return
	}
}
