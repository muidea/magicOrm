package provider

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

func checkType(val interface{}, t *testing.T) bool {
	lType, lErr := local.GetType(reflect.ValueOf(val))
	if lErr != nil {
		t.Errorf("local.GetType failed, err:%s", lErr.Error())
		return false
	}

	rType, rErr := remote.GetType(reflect.ValueOf(val))
	if rErr != nil {
		t.Errorf("local.GetType failed, err:%s", lErr.Error())
		return false
	}

	if !model.CompareType(lType, rType) {
		t.Errorf("compareType failed")
		return false
	}

	return true
}

func TestType(t *testing.T) {
	var iVal int
	if !checkType(iVal, t) {
		return
	}
	var fVal float64
	if !checkType(fVal, t) {
		return
	}
	var bVal bool
	if !checkType(bVal, t) {
		return
	}
	var strVal string
	if !checkType(strVal, t) {
		return
	}
}

func TestModel(t *testing.T) {
	type Base struct {
		//ID 唯一标示单元
		ID  int    `json:"id" orm:"id key auto"`
		I8  int8   `orm:"i8"`
		I16 int16  `orm:"i16"`
		I32 int32  `orm:"i32"`
		I64 uint64 `orm:"i64"`
		// Name 名称
		Name      string    `json:"name" orm:"name"`
		Value     float32   `json:"value" orm:"value"`
		F64       float64   `orm:"f64"`
		TimeStamp time.Time `json:"timeStamp" orm:"ts"`
		Flag      bool      `orm:"flag"`
	}

	base := Base{}
	lModel, lErr := local.GetModel(reflect.ValueOf(base))
	if lErr != nil {
		t.Errorf("local.GetModel failed. err:%s", lErr.Error())
		return
	}

	baseObject, baseErr := remote.GetObject(base)
	if baseErr != nil {
		t.Errorf("remote.GetObject failed, err:%s", baseErr.Error())
		return
	}
	rModel, rErr := remote.GetModel(reflect.ValueOf(baseObject))
	if rErr != nil {
		t.Errorf("remote.GetModel failed. err:%s", rErr.Error())
		return
	}
	if !model.CompareModel(lModel, rModel) {
		t.Errorf("CompareModel failed")
		return
	}
}
