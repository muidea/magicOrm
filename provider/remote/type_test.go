package remote

import (
	"github.com/muidea/magicOrm/model"
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	var ii int
	itemType, itemErr := newType(reflect.TypeOf(ii))
	if itemErr != nil {
		t.Errorf("illegal type")
		return
	}

	if itemType.Name != "int" {
		t.Errorf("illegal type name, name:%s, expect name:%s", itemType.Name, "int")
		return
	}
	if itemType.Value != model.TypeIntegerValue {
		t.Errorf("illegal type value, value:%d, expect value:%d", itemType.Value, model.TypeIntegerValue)
		return
	}
}
