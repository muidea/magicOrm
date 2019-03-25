package remote

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/util"
)

func TestType(t *testing.T) {
	var ii int
	itemType, itemErr := GetType(reflect.TypeOf(ii))
	if itemErr != nil {
		t.Errorf("illegal type")
		return
	}

	if itemType.Name != "int" {
		t.Errorf("illegal type name, name:%s, expect name:%s", itemType.Name, "int")
		return
	}
	if itemType.Value != util.TypeIntegerField {
		t.Errorf("illegal type value, value:%d, expect value:%d", itemType.Value, util.TypeIntegerField)
		return
	}

	if itemType.GetType().Kind() != reflect.Int {
		t.Errorf("illegal type kind, kind:%v, expect kind:%v", itemType.GetType().Kind(), reflect.Int)
		return
	}
}
