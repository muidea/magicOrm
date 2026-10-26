package provider_test

import (
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/utils"
)

// Concrete return types remain assignable at existing application call sites.
var (
	_ func(reflect.Type) (*local.TypeImpl, *cd.Error)       = local.NewType
	_ func(reflect.StructTag) (*local.SpecImpl, *cd.Error)  = local.NewSpec
	_ func(reflect.Value) *local.ValueImpl                  = local.NewValue
	_ func(any) *remote.ValueImpl                           = remote.NewValue
	_ func(*remote.Object) *remote.ObjectFilter             = remote.NewFilter
	_ func(string, models.ValueValidator) provider.Provider = provider.NewLocalProvider
	_ func(string, models.ValueValidator) provider.Provider = provider.NewRemoteProvider
)

func TestPublicConcreteTypesKeepPackageIdentity(t *testing.T) {
	// Aliasing these types to an internal package would preserve many source
	// calls but change reflection identity and registration/serialization keys.
	for _, group := range []struct {
		path  string
		types []reflect.Type
	}{
		{"provider/local", []reflect.Type{reflect.TypeFor[local.TypeImpl](), reflect.TypeFor[local.SpecImpl](), reflect.TypeFor[local.ValueImpl]()}},
		{"utils", []reflect.Type{reflect.TypeFor[utils.Pagination](), reflect.TypeFor[utils.SortFilter]()}},
		{"provider/remote", []reflect.Type{reflect.TypeFor[remote.Object](), reflect.TypeFor[remote.Field](), reflect.TypeFor[remote.TypeImpl](), reflect.TypeFor[remote.SpecImpl](), reflect.TypeFor[remote.ObjectValue](), reflect.TypeFor[remote.SliceObjectValue](), reflect.TypeFor[remote.FieldValue](), reflect.TypeFor[remote.ValueImpl](), reflect.TypeFor[remote.ObjectFilter]()}},
	} {
		for _, typ := range group.types {
			if got, want := typ.PkgPath(), "github.com/muidea/magicOrm/"+group.path; got != want {
				t.Errorf("%s package = %s, want %s", typ.Name(), got, want)
			}
		}
	}
}
