package local

import "testing"

func TestTag(t *testing.T) {
	tag1 := "tag"
	tagPtr, tagErr := newTag(tag1)
	if tagErr != nil {
		t.Errorf("newTag failed, err:%s", tagErr.Error())
		return
	}
	if tagPtr.GetName() != "tag" {
		t.Errorf("newTag failed,current:%s, expect:%s", tagPtr.GetName(), "tag")
	}
	if tagPtr.IsAutoIncrement() || tagPtr.IsPrimaryKey() {
		t.Errorf("newTag failed")
		return
	}

	tag2 := "tag auto"
	tagPtr, tagErr = newTag(tag2)
	if tagErr != nil {
		t.Errorf("newTag failed, err:%s", tagErr.Error())
		return
	}
	if tagPtr.GetName() != "tag" {
		t.Errorf("newTag failed,current:%s, expect:%s", tagPtr.GetName(), "tag")
	}
	if !tagPtr.IsAutoIncrement() || tagPtr.IsPrimaryKey() {
		t.Errorf("newTag failed")
		return
	}

	tag3 := "tag auto key"
	tagPtr, tagErr = newTag(tag3)
	if tagErr != nil {
		t.Errorf("newTag failed, err:%s", tagErr.Error())
		return
	}
	if tagPtr.GetName() != "tag" {
		t.Errorf("newTag failed,current:%s, expect:%s", tagPtr.GetName(), "tag")
	}
	if !tagPtr.IsAutoIncrement() || !tagPtr.IsPrimaryKey() {
		t.Errorf("newTag failed")
		return
	}

	tag4 := "tag key auto"
	tagPtr, tagErr = newTag(tag4)
	if tagErr != nil {
		t.Errorf("newTag failed, err:%s", tagErr.Error())
		return
	}
	if tagPtr.GetName() != "tag" {
		t.Errorf("newTag failed,current:%s, expect:%s", tagPtr.GetName(), "tag")
	}
	if !tagPtr.IsAutoIncrement() || !tagPtr.IsPrimaryKey() {
		t.Errorf("newTag failed")
		return
	}
}
