package test

import (
	"math"
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type NumericCollection struct {
	ID       int64    `orm:"id key" view:"detail,lite"`
	Bytes    []uint8  `orm:"bytes" view:"detail,lite"`
	Signed   []int64  `orm:"signed" view:"detail,lite"`
	Unsigned []uint64 `orm:"unsigned" view:"detail,lite"`
}

func TestNumericCollectionDatabaseRoundTrip(t *testing.T) {
	requireIntegration(t)
	orm.Initialize()
	defer orm.Uninitialized()
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			p := provider.NewLocalProvider("numeric-"+kind, nil)
			var declaration any = &NumericCollection{}
			if kind == "remote" {
				p = provider.NewRemoteProvider("numeric-remote", nil)
				object, e := helper.GetObject(declaration)
				if e != nil {
					t.Fatal(e)
				}
				declaration = object
			}
			model, e := p.RegisterModel(declaration)
			if e != nil {
				t.Fatal(e)
			}
			h, e := orm.NewOrm(p, config, "numeric_"+kind)
			if e != nil {
				t.Fatal(e)
			}
			defer h.Release()
			if e := h.Create(model); e != nil {
				t.Fatal(e)
			}
			defer func() {
				if e := h.Drop(model); e != nil {
					t.Error(e)
				}
			}()
			for n, bytes := range [][]uint8{{}, {0}, {255}, {0, 128, 255}} {
				input := &NumericCollection{ID: int64(n + 1), Bytes: bytes, Signed: []int64{math.MinInt64, -9007199254740993, 0, 9007199254740993, math.MaxInt64}, Unsigned: []uint64{0, 9007199254740993, math.MaxUint64}}
				var value any = input
				if kind == "remote" {
					value, e = helper.GetObjectValue(input)
					if e != nil {
						t.Fatal(e)
					}
				}
				row, e := p.GetEntityModel(value, true)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = h.Insert(row); e != nil {
					t.Fatal(e)
				}
				got, e := h.Query(row)
				if e != nil {
					t.Fatal(e)
				}
				for field, want := range map[string]any{"bytes": input.Bytes, "signed": input.Signed, "unsigned": input.Unsigned} {
					if actual := got.GetField(field).GetValue().Get(); !reflect.DeepEqual(actual, want) {
						t.Fatalf("%s: %#v != %#v", field, actual, want)
					}
				}
			}
		})
	}
}
