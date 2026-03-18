package remote

import "testing"

func TestSpecGetConstraintsCachesParsedResult(t *testing.T) {
	spec := &SpecImpl{Constraint: "req,min=1,max=8"}

	first := spec.GetConstraints()
	second := spec.GetConstraints()
	if first == nil || second == nil {
		t.Fatal("expected parsed constraints")
	}
	if first != second {
		t.Fatal("expected GetConstraints to reuse parsed constraints")
	}
}

func TestSpecCopyPreservesParsedConstraints(t *testing.T) {
	spec := &SpecImpl{Constraint: "req,re=^[a-z]{2,8}$"}
	parsed := spec.GetConstraints()
	copySpec := spec.Copy()
	if copySpec == nil {
		t.Fatal("expected copied spec")
	}

	if got := copySpec.GetConstraints(); got != parsed {
		t.Fatal("expected copied spec to preserve parsed constraints cache")
	}
}
