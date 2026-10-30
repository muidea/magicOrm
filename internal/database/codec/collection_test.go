package codec

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type collectionRecord struct {
	ID       int64    `orm:"id key"`
	Bytes    []uint8  `orm:"bytes"`
	Signed   []int64  `orm:"signed"`
	Unsigned []uint64 `orm:"unsigned"`
	Small    []int8   `orm:"small"`
	Strings  []string `orm:"strings"`
}

func collectionModels(t *testing.T, input *collectionRecord) map[string]struct {
	provider provider.Provider
	model    models.Model
} {
	t.Helper()
	local := provider.NewLocalProvider("collections", nil)
	if _, err := local.RegisterModel(input); err != nil {
		t.Fatal(err)
	}
	lm, err := local.GetEntityModel(input, true)
	if err != nil {
		t.Fatal(err)
	}
	declaration, err := helper.GetObject(input)
	if err != nil {
		t.Fatal(err)
	}
	value, err := helper.GetObjectValue(input)
	if err != nil {
		t.Fatal(err)
	}
	remote := provider.NewRemoteProvider("collections", nil)
	if _, err := remote.RegisterModel(declaration); err != nil {
		t.Fatal(err)
	}
	rm, err := remote.GetEntityModel(value, true)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]struct {
		provider provider.Provider
		model    models.Model
	}{"local": {local, lm}, "remote": {remote, rm}}
}

func TestStoredCollectionRoundTripUsesNumericArraysAndExactIntegers(t *testing.T) {
	for _, bytes := range [][]uint8{nil, {}, {0}, {255}, {0, 128, 255}} {
		input := &collectionRecord{ID: 1, Bytes: bytes, Signed: []int64{math.MinInt64, -9007199254740993, 0, 9007199254740993, math.MaxInt64}, Unsigned: []uint64{0, 9007199254740993, math.MaxUint64}}
		for name, fixture := range collectionModels(t, input) {
			t.Run(name, func(t *testing.T) {
				c := New(fixture.provider, "test")
				for fieldName, want := range map[string]any{"bytes": input.Bytes, "signed": input.Signed, "unsigned": input.Unsigned} {
					field := fixture.model.GetField(fieldName)
					if fieldName == "bytes" && bytes == nil {
						want = []uint8{}
					}
					packed, err := c.PackedBasicFieldValue(field, field.GetValue())
					if err != nil {
						t.Fatal(err)
					}
					if fieldName == "bytes" {
						var numbers []uint16
						if err := json.Unmarshal([]byte(packed.(string)), &numbers); err != nil {
							t.Fatalf("bytes not numeric array: %s: %v", packed, err)
						}
					}
					got, err := c.ExtractBasicFieldValue(field, packed)
					if err != nil || !reflect.DeepEqual(got, want) {
						t.Fatalf("%s: packed=%s got=%#v want=%#v error=%v", fieldName, packed, got, want, err)
					}
				}
			})
		}
	}
}

func TestStoredCollectionCompatibilityIsRestrictedToDeclaredUint8(t *testing.T) {
	for name, fixture := range collectionModels(t, &collectionRecord{ID: 1}) {
		t.Run(name, func(t *testing.T) {
			c := New(fixture.provider, "test")
			for _, raw := range []any{`"/w=="`, []byte(`"/w=="`)} {
				got, err := c.ExtractBasicFieldValue(fixture.model.GetField("bytes"), raw)
				if err != nil || !reflect.DeepEqual(got, []uint8{255}) {
					t.Fatalf("legacy uint8: %#v %v", got, err)
				}
			}
			for _, raw := range []string{`null`, `[]`} {
				if got, err := c.ExtractBasicFieldValue(fixture.model.GetField("bytes"), raw); err != nil || !reflect.DeepEqual(got, []uint8{}) {
					t.Fatalf("empty collection: %#v %v", got, err)
				}
			}
			for field, inputs := range map[string][]string{
				"bytes":    {`"bad base64"`, `[-1]`, `[256]`, `[1.5]`, `[null]`, `["255"]`, `[255] [0]`, `{}`},
				"signed":   {`"/w=="`, `[9223372036854775808]`, `[-9223372036854775809]`, `[0.5]`},
				"unsigned": {`[-1]`, `[18446744073709551616]`},
				"small":    {`[128]`, `[-129]`},
				"strings":  {`"/w=="`},
			} {
				for _, raw := range inputs {
					if got, err := c.ExtractBasicFieldValue(fixture.model.GetField(field), raw); err == nil {
						t.Fatalf("%s accepted %s as %#v", field, raw, got)
					}
				}
			}
		})
	}
}
