package provider

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestModel(t *testing.T) {
	type Base struct {
		//ID 唯一标示单元
		ID  int    `orm:"id key auto"`
		I8  int8   `orm:"i8"`
		I16 int16  `orm:"i16"`
		I32 int32  `orm:"i32"`
		I64 uint64 `orm:"i64"`
		// name 名称
		Name      string    `orm:"name"`
		Value     float32   `orm:"value"`
		F64       float64   `orm:"f64"`
		TimeStamp time.Time `orm:"ts"`
		Flag      bool      `orm:"flag"`
	}

	base := &Base{}
	lModel, lErr := local.GetEntityModel(base)
	if lErr != nil {
		t.Errorf("local.GetEntityModel failed. err:%s", lErr.Error())
		return
	}

	baseObject, baseErr := remote.GetObject(base)
	if baseErr != nil {
		t.Errorf("remote.GetObject failed, err:%s", baseErr.Error())
		return
	}
	rModel, rErr := remote.GetEntityModel(baseObject)
	if rErr != nil {
		t.Errorf("remote.GetEntityModel failed. err:%s", rErr.Error())
		return
	}

	baseObjectVal, baseValErr := remote.GetObjectValue(base)
	if baseValErr != nil {
		t.Errorf("remote.GetObjectValue failed, err:%s", baseErr.Error())
		return
	}
	rVal, rErr := remote.GetEntityValue(baseObjectVal)
	if rErr != nil {
		t.Errorf("remote.GetEntityValue failed. err:%s", rErr.Error())
		return
	}

	rModel, rErr = remote.SetModelValue(rModel, rVal)
	if rErr != nil {
		t.Errorf("remote.SetModelValue failed. err:%s", rErr.Error())
		return
	}

	if !model.CompareModel(lModel, rModel) {
		t.Errorf("CompareModel failed")
		return
	}
}
