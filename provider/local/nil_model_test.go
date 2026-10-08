package local

import (
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

func TestNilRelatedModelReturnsErrorWithoutPanic(t *testing.T) {
	var target *ValidationTestStruct
	if _, err := GetEntityModel(target, nil); err == nil || err.Code != cd.IllegalParam {
		t.Fatalf("typed nil entity accepted: %v", err)
	}
	for _, value := range []reflect.Value{reflect.ValueOf(target), {}} {
		if _, err := getValueModel(value, models.OriginView); err == nil || err.Code != cd.IllegalParam {
			t.Fatalf("invalid model accepted: %v", err)
		}
	}
	model, err := GetEntityModel(&ValidationTestStruct{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = SetModelValue(model, NewValue(reflect.ValueOf(target)), true); err == nil || err.Code != cd.IllegalParam {
		t.Fatalf("nil SetModelValue accepted: %v", err)
	}
}
