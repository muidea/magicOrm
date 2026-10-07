package remote_test

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type integerWireRecord struct {
	ID       int64             `orm:"id key"`
	Signed   int64             `orm:"signed"`
	Unsigned uint64            `orm:"unsigned"`
	Numbers  []int64           `orm:"numbers"`
	Optional *int64            `orm:"optional"`
	Child    *integerWireChild `orm:"child"`
}

type integerWireChild struct {
	ID    int64 `orm:"id key"`
	Value int64 `orm:"value"`
}

func TestIntegerWireRoundTripPreservesDeclaredTypes(t *testing.T) {
	for _, value := range []int64{math.MinInt64, math.MinInt64 + 1, -(1 << 53) - 1, 0, (1 << 53) + 1, math.MaxInt64 - 1, math.MaxInt64} {
		want := integerWireRecord{ID: value, Signed: value, Unsigned: math.MaxUint64, Numbers: []int64{0, (1 << 53) + 1, value}, Optional: &value, Child: &integerWireChild{ID: 1, Value: value}}
		wire, valueErr := helper.GetObjectValue(&want)
		if valueErr != nil {
			t.Fatal(valueErr)
		}
		data, err := json.Marshal(wire)
		if err != nil {
			t.Fatal(err)
		}
		// Ordinary DTO parsing must be just as exact as the ORM decoder.
		var dto struct {
			Value *remote.ObjectValue `json:"value"`
		}
		if err := json.Unmarshal(append(append([]byte(`{"value":`), data...), '}'), &dto); err != nil {
			t.Fatal(err)
		}
		converted, convertErr := remote.ConvertObjectValue(dto.Value)
		if convertErr != nil {
			t.Fatal(convertErr)
		}
		for _, object := range []*remote.ObjectValue{converted, converted.Copy()} {
			var got integerWireRecord
			if err := helper.UpdateEntity(object, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("integer %d changed: got=%+v want=%+v", value, got, want)
			}
		}
		sliceData, err := json.Marshal(&remote.SliceObjectValue{Name: wire.Name, PkgPath: wire.PkgPath, Values: []*remote.ObjectValue{wire}})
		if err != nil {
			t.Fatal(err)
		}
		slice, decodeErr := remote.DecodeSliceObjectValue(sliceData)
		if decodeErr != nil {
			t.Fatal(decodeErr)
		}
		var got integerWireRecord
		if err := helper.UpdateEntity(slice.Copy().Values[0], &got); err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("slice integer %d changed: got=%+v err=%v", value, got, err)
		}
	}
}

func TestIntegerWireFilterAndAssignedZero(t *testing.T) {
	var filter remote.ObjectFilter
	if err := json.Unmarshal([]byte(`{"equal":[{"name":"id","value":9223372036854775807}],"in":[{"name":"id","value":[0,9007199254740993,9223372036854775807]}]}`), &filter); err != nil {
		t.Fatal(err)
	}
	if filter.EqualFilter[0].Value != json.Number("9223372036854775807") {
		t.Fatal("filter lost numeric precision", filter.EqualFilter)
	}
	var field remote.FieldValue
	if err := json.Unmarshal([]byte(`{"name":"value","value":0,"assigned":true}`), &field); err != nil {
		t.Fatal(err)
	}
	if !field.Assigned || !field.IsZero() || !field.GetValue().(*remote.ValueImpl).IsAssigned() {
		t.Fatal("explicit numeric zero lost assignment", field)
	}
	if err := json.Unmarshal([]byte(`{"name":"value","value":0}`), &field); err != nil {
		t.Fatal(err)
	}
	if field.GetValue().(*remote.ValueImpl).IsAssigned() {
		t.Fatal("implicit zero became assigned", field)
	}
}
