package test

import "testing"

func isZeroGroupView(val *Group) bool {
	return val != nil && val.ID == 0 && val.Name == "" && val.Users == nil && val.Parent == nil
}

func assertStatusLiteView(t *testing.T, got, want *Status) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Fatalf("unexpected status: %#v", got)
		}
		return
	}
	if got == nil {
		t.Fatal("status is nil")
	}
	if got.ID != want.ID || got.Value != want.Value {
		t.Fatalf("status mismatch, got=%#v want=%#v", got, want)
	}
}

func assertGroupLiteView(t *testing.T, got, want *Group) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Fatalf("unexpected group: %#v", got)
		}
		return
	}
	if got == nil {
		t.Fatal("group is nil")
	}
	if got.ID != want.ID || got.Name != want.Name {
		t.Fatalf("group mismatch, got=%#v want=%#v", got, want)
	}

	if want.Parent == nil {
		if got.Parent != nil && !isZeroGroupView(got.Parent) {
			t.Fatalf("unexpected group parent: %#v", got.Parent)
		}
		return
	}
	if got.Parent == nil {
		t.Fatal("group parent is nil")
	}
	if got.Parent.ID != want.Parent.ID || got.Parent.Name != want.Parent.Name {
		t.Fatalf("group parent mismatch, got=%#v want=%#v", got.Parent, want.Parent)
	}
}

func assertUserDetailWithLiteRelations(t *testing.T, got, want *User) {
	t.Helper()

	if got == nil || want == nil {
		t.Fatalf("user is nil, got=%#v want=%#v", got, want)
	}
	if got.ID != want.ID || got.Name != want.Name || got.EMail != want.EMail {
		t.Fatalf("user basic fields mismatch, got=%#v want=%#v", got, want)
	}

	assertStatusLiteView(t, got.Status, want.Status)

	if len(got.Group) != len(want.Group) {
		t.Fatalf("user group length mismatch, got=%d want=%d", len(got.Group), len(want.Group))
	}
	for idx := range want.Group {
		assertGroupLiteView(t, got.Group[idx], want.Group[idx])
	}
}

func assertGroupDetailWithLiteParent(t *testing.T, got, want *Group) {
	t.Helper()

	if got == nil || want == nil {
		t.Fatalf("group is nil, got=%#v want=%#v", got, want)
	}
	if got.ID != want.ID || got.Name != want.Name {
		t.Fatalf("group basic fields mismatch, got=%#v want=%#v", got, want)
	}

	if want.Parent == nil {
		if got.Parent != nil && !isZeroGroupView(got.Parent) {
			t.Fatalf("unexpected parent: %#v", got.Parent)
		}
		return
	}

	if got.Parent == nil {
		t.Fatal("parent is nil")
	}
	if got.Parent.ID != want.Parent.ID || got.Parent.Name != want.Parent.Name {
		t.Fatalf("group parent mismatch, got=%#v want=%#v", got.Parent, want.Parent)
	}
}
