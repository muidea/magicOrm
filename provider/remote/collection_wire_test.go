package remote_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type byteWireRecord struct {
	ID       int64            `orm:"id key"`
	Bytes    []uint8          `orm:"bytes"`
	Child    *byteWireChild   `orm:"child"`
	Children []*byteWireChild `orm:"children"`
}
type byteWireChild struct {
	ID    int64   `orm:"id key"`
	Bytes []uint8 `orm:"bytes"`
}

func TestUint8CollectionJSONPreservesInternalValuesAndNestedNumericArrays(t *testing.T) {
	for _, numbers := range [][]uint8{nil, {}, {0}, {255}, {0, 255}} {
		want := &byteWireRecord{ID: 1, Bytes: numbers, Child: &byteWireChild{ID: 2, Bytes: numbers}, Children: []*byteWireChild{{ID: 3, Bytes: numbers}}}
		value, err := helper.GetObjectValue(want)
		if err != nil {
			t.Fatal(err)
		}
		before := value.GetFieldValue("bytes")
		payload, jsonErr := json.Marshal(value)
		if jsonErr != nil {
			t.Fatal(jsonErr)
		}
		if bytes.Contains(payload, []byte(`"/w=="`)) {
			t.Fatalf("Base64 wire: %s", payload)
		}
		// Serializing must leave the typed collection available to local consumers.
		if !reflect.DeepEqual(value.GetFieldValue("bytes"), before) {
			t.Fatal("JSON changed internal slice")
		}
		var decoded remote.ObjectValue
		if err := json.Unmarshal(payload, &decoded); err != nil {
			t.Fatal(err)
		}
		converted, convertErr := remote.ConvertObjectValue(&decoded)
		if convertErr != nil {
			t.Fatal(convertErr)
		}
		if numbers == nil {
			want.Bytes = []uint8{}
			want.Child.Bytes = []uint8{}
			want.Children[0].Bytes = []uint8{}
		}
		var got byteWireRecord
		if err := helper.UpdateEntity(converted, &got); err != nil || !reflect.DeepEqual(&got, want) {
			t.Fatalf("got=%#v want=%#v error=%v", got, want, err)
		}
	}
	plain, err := json.Marshal([]byte{255})
	if err != nil || string(plain) != `"/w=="` {
		t.Fatal("ordinary binary JSON changed")
	}
}

func TestFieldValueJSONPreservesNilAndEmptyCollections(t *testing.T) {
	for _, tc := range []struct {
		value []uint8
		want  string
	}{{nil, "null"}, {[]uint8{}, "[]"}, {[]uint8{255}, "[255]"}} {
		field := remote.FieldValue{Name: "numbers", Value: tc.value}
		raw, err := json.Marshal(field.JSONValue())
		if err != nil || string(raw) != tc.want || !reflect.DeepEqual(field.Value, tc.value) {
			t.Fatalf("got %s: %v", raw, err)
		}
	}
}

func TestFieldValueJSONProjectsPointerToNumericCollection(t *testing.T) {
	for _, numbers := range [][]uint8{nil, {}, {0, 255}} {
		field := &remote.FieldValue{Name: "numbers", Value: &numbers}
		want, err := json.Marshal(field.JSONValue())
		if err != nil {
			t.Fatal(err)
		}
		if numbers != nil && len(numbers) > 0 && string(want) != "[0,255]" {
			t.Fatalf("pointer encoded as binary: %s", want)
		}
		if !reflect.DeepEqual(field.Value, &numbers) {
			t.Fatal("projection changed pointer value")
		}
	}
	var missing *[]uint8
	if got := (&remote.FieldValue{Name: "numbers", Value: missing}).JSONValue(); got != nil {
		t.Fatal("nil pointer is not null")
	}
}
