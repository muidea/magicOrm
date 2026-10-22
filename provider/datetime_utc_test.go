package provider

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestDateTimeStorageUsesUTC(t *testing.T) {
	instant := time.Date(2026, 10, 21, 0, 30, 45, 123000000, time.FixedZone("business", 8*3600))
	localType, err := local.NewType(reflect.TypeOf(instant))
	if err != nil {
		t.Fatal(err)
	}
	localPointer, err := local.NewType(reflect.TypeOf(&instant))
	if err != nil {
		t.Fatal(err)
	}
	localSlice, err := local.NewType(reflect.TypeOf([]time.Time{}))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []struct {
		name                   string
		provider               Provider
		scalar, pointer, slice models.Type
	}{
		{"local", NewLocalProvider("utc", nil), localType, localPointer, localSlice},
		{"remote", NewRemoteProvider("utc", nil), &remote.TypeImpl{Name: "datetime"}, &remote.TypeImpl{Name: "datetime", IsPtr: true}, &remote.TypeImpl{Name: "datetime", ElemType: &remote.TypeImpl{Name: "datetime"}}},
	} {
		t.Run(p.name, func(t *testing.T) {
			want := "2026-10-20 16:30:45.123"
			for _, value := range []any{instant, instant.UTC(), "2026-10-21T00:30:45.123+08:00", "2026-10-20T11:30:45.123-05:00", "2026-10-20T16:30:45.123Z", want} {
				encoded, err := p.provider.EncodeValue(value, p.scalar)
				if err != nil || encoded != want {
					t.Fatalf("same instant %v encoded as %v: %v", value, encoded, err)
				}
			}
			encoded, err := p.provider.EncodeValue(&instant, p.pointer)
			if err != nil {
				t.Fatal(err)
			}
			text, ok := encoded.(*string)
			if !ok || text == nil || *text != want {
				t.Fatalf("nullable datetime encoded as %v", encoded)
			}
			encoded, err = p.provider.EncodeValue([]time.Time{instant, instant.UTC()}, p.slice)
			if err != nil || !reflect.DeepEqual(encoded, []string{want, want}) {
				t.Fatalf("datetime slice encoded as %v: %v", encoded, err)
			}
			for _, value := range []any{"2026-02-30T00:00:00Z", "2026-10-21T00:30:45", "local time", 123} {
				if _, err := p.provider.EncodeValue(value, p.scalar); err == nil {
					t.Fatalf("ambiguous or invalid datetime admitted: %v", value)
				}
			}
			// Ordinary strings remain ordinary strings; only datetime fields convert.
			strType, err := local.NewType(reflect.TypeOf(""))
			var stringType models.Type = strType
			if p.name == "remote" {
				stringType, err = &remote.TypeImpl{Name: "string"}, nil
			}
			if err != nil {
				t.Fatal(err)
			}
			if text, err := p.provider.EncodeValue("2026-10-21T00:30:45.123+08:00", stringType); err != nil || text != "2026-10-21T00:30:45.123+08:00" {
				t.Fatal("ordinary string changed", text, err)
			}
		})
	}
	decoded, err := local.DecodeValue("2026-10-21T00:30:45.123+08:00", localType)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := decoded.(time.Time); !ok || !value.Equal(instant) || value.Location() != time.UTC {
		t.Fatal("decoded datetime is not the same UTC instant", decoded)
	}
}
