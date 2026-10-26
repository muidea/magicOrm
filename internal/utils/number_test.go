package utils

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func TestNumericConversionsKeepExactBoundsAndRejectOverflow(t *testing.T) {
	for _, tt := range []struct {
		kind  reflect.Kind
		input any
		want  any
	}{
		{reflect.Int64, json.Number("9223372036854775807"), int64(math.MaxInt64)},
		{reflect.Int64, json.Number("-9223372036854775808"), int64(math.MinInt64)},
		{reflect.Int64, json.Number("9007199254740993.0"), int64(9007199254740993)},
		{reflect.Int64, json.Number("9.007199254740993e15"), int64(9007199254740993)},
		{reflect.Uint64, json.Number("18446744073709551615"), uint64(math.MaxUint64)},
		{reflect.Int8, json.Number("127.9"), int8(127)},
		{reflect.Int64, json.Number("1e-9999999999"), int64(0)},
		{reflect.Float64, json.Number("1.25e2"), float64(125)},
	} {
		got, err := convertNumberVal(tt.kind, reflect.ValueOf(tt.input))
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("kind=%v input=%v got=%v want=%v err=%v", tt.kind, tt.input, got, tt.want, err)
		}
	}
	for _, tt := range []struct {
		kind  reflect.Kind
		input any
	}{
		{reflect.Int64, json.Number("9223372036854775808")},
		{reflect.Uint64, json.Number("18446744073709551616")},
		{reflect.Int64, json.Number("1e9999999999")},
		{reflect.Int8, json.Number("128")},
		{reflect.Int8, int64(128)},
		{reflect.Uint8, int64(-1)},
		{reflect.Int64, uint64(math.MaxUint64)},
		{reflect.Int64, math.Ldexp(1, 63)},
		{reflect.Uint64, math.Ldexp(1, 64)},
		{reflect.Uint64, -1.0},
		{reflect.Int64, math.NaN()},
		{reflect.Int64, math.Inf(1)},
		{reflect.Float32, math.MaxFloat64},
	} {
		if value, err := convertNumberVal(tt.kind, reflect.ValueOf(tt.input)); err == nil {
			t.Fatalf("kind=%v input=%v silently became %v", tt.kind, tt.input, value)
		}
	}
}

func TestIntegerComparisonDoesNotCollapseAdjacentLargeValues(t *testing.T) {
	for _, pair := range [][2]any{
		{int64(math.MaxInt64), json.Number("9223372036854775806")},
		{uint64(math.MaxUint64), uint64(math.MaxUint64 - 1)},
		{json.Number("9007199254740993"), float64(9007199254740992)},
	} {
		if same, _ := CompareWithNumericConversion(pair[0], pair[1]); same {
			t.Fatal("different integers compared equal", pair)
		}
	}
	if same, diff := CompareWithNumericConversion(int64(math.MaxInt64), json.Number("9223372036854775807")); !same {
		t.Fatal(diff)
	}
}
