package orm

import (
	"testing"
)

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int64", int64(42), "42"},
		{"int", 42, "42"},
		{"string", "uuid-string-123", "uuid-string-123"},
		{"zero int64", int64(0), "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeID(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeID(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// assertSameSet 比较两个 slice 是否作为集合相等（顺序无关）
func assertSameSet(t *testing.T, label string, want, got []any) {
	t.Helper()
	wantSet := make(map[string]int)
	for _, v := range want {
		key := normalizeID(v)
		wantSet[key]++
	}
	gotSet := make(map[string]int)
	for _, v := range got {
		key := normalizeID(v)
		gotSet[key]++
	}
	if len(wantSet) != len(gotSet) {
		t.Errorf("%s: set size mismatch: want %d elements %v, got %d elements %v", label, len(want), want, len(got), got)
		return
	}
	for k, c := range wantSet {
		if gotSet[k] != c {
			t.Errorf("%s: count for %q: want %d, got %d (want=%v got=%v)", label, k, c, gotSet[k], want, got)
		}
	}
}

func TestDiffRelationIDs(t *testing.T) {
	tests := []struct {
		name       string
		existing   []any
		new        []any
		wantDelete []any
		wantInsert []any
	}{
		{"both empty", nil, nil, nil, nil},
		{"both empty slices", []any{}, []any{}, nil, nil},
		{"existing empty, new has items", nil, []any{1, 2}, nil, []any{1, 2}},
		{"new empty, existing has items", []any{1, 2}, nil, []any{1, 2}, nil},
		{"identical sets", []any{1, 2}, []any{1, 2}, nil, nil},
		{"partial overlap", []any{1, 2}, []any{2, 3}, []any{1}, []any{3}},
		{"complete replacement", []any{1, 2}, []any{3, 4}, []any{1, 2}, []any{3, 4}},
		{"new has duplicates", []any{1}, []any{2, 2}, []any{1}, []any{2}},
		{"string IDs", []any{"a", "b"}, []any{"b", "c"}, []any{"a"}, []any{"c"}},
		{"mixed int types", []any{int64(1), int64(2)}, []any{int64(2), int64(3)}, []any{int64(1)}, []any{int64(3)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDelete, gotInsert := diffRelationIDs(tt.existing, tt.new)
			assertSameSet(t, "toDelete", tt.wantDelete, gotDelete)
			assertSameSet(t, "toInsert", tt.wantInsert, gotInsert)
		})
	}
}
