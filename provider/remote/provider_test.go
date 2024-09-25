package remote

import (
	fu "github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/model"
	"testing"
	"time"
)

func TestGetEntityType(t *testing.T) {
	var entity interface{}
	iVal := 123

	entity = iVal
	eTypeVal, eTypeErr := GetEntityType(entity)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}
	if !eTypeVal.IsBasic() {
		t.Errorf("GetEntityType failed")
		return
	}
	if eTypeVal.GetValue() != model.TypeIntegerValue {
		t.Errorf("GetEntityType failed")
		return
	}

	typePtr, typeErr := getEntityType(entity)
	if typeErr != nil {
		t.Errorf("getEntityType failed, error:%s", typeErr.Error())
		return
	}

	eTypeVal, eTypeErr = GetEntityType(typePtr)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}
	if !eTypeVal.IsBasic() {
		t.Errorf("GetEntityType failed")
		return
	}
	if eTypeVal.GetValue() != model.TypeIntegerValue {
		t.Errorf("GetEntityType failed")
		return
	}

	eVal, eErr := eTypeVal.Interface(entity)
	if eErr != nil {
		t.Errorf("eTypeVal.Interface failed, error:%s", eErr.Error())
		return
	}
	iVal2, iVal2OK := eVal.Interface().(int)
	if !iVal2OK || iVal2 != iVal {
		t.Errorf("eVal.Interface().(int) failed")
		return
	}

	dt := time.Now()
	entity = &dt
	eTypeVal, eTypeErr = GetEntityType(entity)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}
	if !eTypeVal.IsBasic() || !eTypeVal.IsPtrType() {
		t.Errorf("GetEntityType failed")
		return
	}
	if eTypeVal.GetValue() != model.TypeDateTimeValue {
		t.Errorf("GetEntityType failed")
		return
	}
	eVal, eErr = eTypeVal.Interface(entity)
	if eErr != nil {
		t.Errorf("eTypeVal.Interface failed, error:%s", eErr.Error())
		return
	}
	strVal, strValOK := eVal.Interface().(string)
	if !strValOK || strVal != dt.Format(fu.CSTLayout) {
		t.Errorf("eVal.Interface().(string) failed")
		return
	}
}

func TestGetEntityValue(t *testing.T) {
	var entity interface{}
	iVal := 123

	entity = iVal
	eTypeVal, eTypeErr := GetEntityType(entity)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}
	if !eTypeVal.IsBasic() {
		t.Errorf("GetEntityType failed")
		return
	}
	if eTypeVal.GetValue() != model.TypeIntegerValue {
		t.Errorf("GetEntityType failed")
		return
	}

	eValueVal, eValueErr := GetEntityValue(entity)
	if eValueErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eValueErr.Error())
		return
	}

	encodeVal, encodeErr := _codec.Encode(eValueVal, eTypeVal)
	if encodeErr != nil {
		t.Errorf("_codec.Encode failed, error:%s", encodeErr.Error())
		return
	}
	switch encodeVal.(type) {
	case int:
		t.Logf("%d", encodeVal.(int))
	default:
		t.Errorf("_codec.Encode failed")
		return
	}

	dt := time.Now()
	entity = &dt
	eTypeVal, eTypeErr = GetEntityType(entity)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}
	if !eTypeVal.IsBasic() || !eTypeVal.IsPtrType() {
		t.Errorf("GetEntityType failed")
		return
	}
	if eTypeVal.GetValue() != model.TypeDateTimeValue {
		t.Errorf("GetEntityType failed")
		return
	}
	eValueVal, eValueErr = GetEntityValue(entity)
	if eValueErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eValueErr.Error())
		return
	}

	encodeVal, encodeErr = _codec.Encode(eValueVal, eTypeVal)
	if encodeErr != nil {
		t.Errorf("_codec.Encode failed, error:%s", encodeErr.Error())
		return
	}
	switch encodeVal.(type) {
	case string:
		t.Logf("%s", encodeVal.(string))
	default:
		t.Errorf("_codec.Encode failed")
		return
	}
}

func TestRemoteModel(t *testing.T) {
	sPtr := &Simple{
		ID: 100,
	}

	eTypeVal, eTypeErr := GetEntityType(sPtr)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}

	eValueVal, eValueErr := GetEntityValue(sPtr)
	if eValueErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eValueErr.Error())
		return
	}
	var valuePtr *ObjectValue
	switch eValueVal.Interface().(type) {
	case *ObjectValue:
		valuePtr = eValueVal.Interface().(*ObjectValue)
	default:
		t.Errorf("GetEntityValue failed")
		return
	}
	if valuePtr.GetPkgKey() != eTypeVal.GetPkgKey() {
		t.Errorf("mismatch pkgKey")
		return
	}

	mCache := model.NewCache()
	sModel, sErr := GetEntityModel(sPtr)
	if sErr != nil {
		t.Errorf("GetEntityModel failed, error:%s", sErr.Error())
		return
	}
	mCache.Put(sModel.GetPkgKey(), sModel)
	eVal, eErr := EncodeValue(eValueVal, eTypeVal, mCache)
	if eErr != nil {
		t.Errorf("EncodeValue failed,error:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int64:
		t.Logf("%d", eVal.(int64))
	default:
		t.Errorf("EncodeValue failed")
	}
}
