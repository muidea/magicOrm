package provider

import (
	fu "github.com/muidea/magicCommon/foundation/util"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type Simple struct {
	ID        int       `orm:"id key auto" view:"view,lite"`
	I8        int8      `orm:"i8" view:"view,lite"`
	I16       int16     `orm:"i16" view:"view,lite"`
	I32       int32     `orm:"i32" view:"view,lite"`
	I64       uint64    `orm:"i64" view:"view,lite"`
	Name      string    `orm:"name" view:"view,lite"`
	Value     float32   `orm:"value" view:"view,lite"`
	F64       float64   `orm:"f64" view:"view,lite"`
	TimeStamp time.Time `orm:"ts dateTime" view:"view,lite"`
	Flag      bool      `orm:"flag" view:"view,lite"`
	Namespace string    `orm:"namespace"`
}

func TestNewLocalProvider(t *testing.T) {
	s001 := &Simple{}
	provider := NewLocalProvider("t001")
	sModelVal, sModelErr := provider.RegisterModel(s001)
	if sModelErr != nil {
		t.Errorf("%s", sModelErr.Error())
		return
	}

	sModelVal, sModelErr = provider.GetEntityModel(s001)
	if sModelErr != nil {
		t.Errorf("%s", sModelErr.Error())
		return
	}

	s001.ID = 100
	s001.I8 = 8
	s001.I16 = 16
	s001.I32 = 32
	s001.I64 = 64
	s001.Name = "s001"
	s001.Value = 123.456
	s001.F64 = 456789.123
	s001.TimeStamp = time.Now()
	s001.Flag = true
	s001.Namespace = "t001"
	sValueVal, sValueErr := provider.GetEntityValue(s001)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr.Error())
		return
	}

	sTypeVal, sTypeErr := provider.GetEntityType(s001)
	if sTypeErr != nil {
		t.Errorf("%s", sTypeErr.Error())
		return
	}

	_, sFilterErr := provider.GetEntityFilter(s001, model.LiteView)
	if sFilterErr != nil {
		t.Errorf("%s", sFilterErr.Error())
		return
	}

	eVal, eErr := provider.EncodeValue(sValueVal, sTypeVal)
	if eErr != nil {
		t.Errorf("%s", eErr.Error())
		return
	}
	iVal, iOK := eVal.(int)
	if !iOK || iVal != 100 {
		t.Errorf("provider.EncodeValue failed, illegal value")
		return
	}

	sModelVal, sModelErr = provider.GetValueModel(sValueVal, sTypeVal)
	if sModelErr != nil {
		t.Errorf("%s", sModelErr.Error())
		return
	}
	s001NewVal := sModelVal.Interface(true, model.OriginView).(*Simple)
	if !util.IsSameValue(s001, s001NewVal) {
		t.Errorf("get value model failed")
		return
	}

	s001NewVal.ID = 987654
	sValueVal, sValueErr = provider.DecodeValue(s001NewVal, sTypeVal)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr)
		return
	}
	s001NewVal2 := sValueVal.Interface().(*Simple)
	if !util.IsSameValue(s001NewVal, s001NewVal2) {
		t.Errorf("decode value model failed")
		return
	}

	sValueVal, sValueErr = provider.DecodeValue(23456, sTypeVal)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr)
		return
	}
	s001NewVal3 := sValueVal.Interface().(*Simple)
	if s001NewVal3.ID != 23456 {
		t.Errorf("decode value model failed")
		return
	}

	dt001 := time.Now().UTC()
	dt001Str := dt001.Format(fu.CSTLayout)
	dtValueVal, dtValueErr := provider.GetEntityValue(&dt001)
	if dtValueErr != nil {
		t.Errorf("%s", dtValueErr.Error())
		return
	}

	dtTypeVal, dtTypeErr := provider.GetEntityType(&dt001)
	if dtTypeErr != nil {
		t.Errorf("%s", dtTypeErr.Error())
		return
	}
	if !dtTypeVal.IsPtrType() {
		t.Errorf("illetal type value")
		return
	}
	eVal, eErr = provider.EncodeValue(dtValueVal, dtTypeVal)
	if eErr != nil {
		t.Errorf("%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		t.Logf("%s", eVal.(string))
	default:
		t.Errorf("illegal value")
		return
	}

	dtValueVal, dtValueErr = provider.DecodeValue(dt001Str, dtTypeVal)
	if dtValueErr != nil {
		t.Errorf("%s", dtValueErr.Error())
		return
	}

	dVal := dtValueVal.Interface()
	switch dVal.(type) {
	case *time.Time:
	default:
		t.Errorf("decodeValue failed")
	}

	gap := dtValueVal.Interface().(*time.Time).Sub(dt001)
	if gap >= time.Second || gap < -1*time.Second {
		t.Errorf("illegal decode value")
	}

	provider.UnregisterModel(s001)
}
