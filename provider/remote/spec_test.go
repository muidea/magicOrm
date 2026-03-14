package remote

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestSpecImplGetters(t *testing.T) {
	spec := &SpecImpl{
		FieldName:    "name",
		PrimaryKey:   true,
		ValueDeclare: models.UUID,
		ViewDeclare:  []models.ViewDeclare{models.DetailView},
		Constraint:   "req",
		DefaultValue: "guest",
	}

	if spec.GetFieldName() != "name" {
		t.Fatalf("GetFieldName mismatch, got %q", spec.GetFieldName())
	}
	if !spec.IsPrimaryKey() {
		t.Fatal("IsPrimaryKey mismatch")
	}
	if spec.GetValueDeclare() != models.UUID {
		t.Fatalf("GetValueDeclare mismatch, got %v", spec.GetValueDeclare())
	}
	if spec.GetDefaultValue() != "guest" {
		t.Fatalf("GetDefaultValue mismatch, got %#v", spec.GetDefaultValue())
	}
}

func TestCompareSpec(t *testing.T) {
	left := &SpecImpl{FieldName: "name", PrimaryKey: true, ValueDeclare: models.UUID}
	right := &SpecImpl{FieldName: "name", PrimaryKey: true, ValueDeclare: models.UUID}
	if !compareSpec(left, right) {
		t.Fatal("compareSpec identical specs should be true")
	}
	if compareSpec(left, &SpecImpl{FieldName: "other", PrimaryKey: true, ValueDeclare: models.UUID}) {
		t.Fatal("compareSpec different field names should be false")
	}
	if compareSpec(left, nil) || compareSpec(nil, left) {
		t.Fatal("compareSpec nil mismatch should be false")
	}
	if !compareSpec(nil, nil) {
		t.Fatal("compareSpec nil specs should be true")
	}
}
