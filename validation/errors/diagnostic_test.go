package errors

import (
	"strings"
	"testing"
)

func TestMultipleErrorsRetainFieldConstraintAndReason(t *testing.T) {
	collector := NewErrorCollector()
	collector.AddError(NewValidationError("constraint 'min': too small/short").WithField("price").WithConstraint("min"))
	collector.AddError(NewValidationError("constraint 'in': must be one of [A B]").WithField("choice").WithConstraint("in"))
	err := collector.ToRichError()
	if err == nil {
		t.Fatal("missing diagnostics")
	}
	for _, part := range []string{"field 'price'", "constraint 'min'", "too small/short", "field 'choice'", "constraint 'in'", "must be one of"} {
		if !strings.Contains(err.Message, part) {
			t.Fatalf("missing %q: %v", part, err)
		}
	}
	if err.Message != collector.ToRichError().Message {
		t.Fatal("unstable diagnostic order")
	}
}
