package remote

import (
	fu "github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/model"
	"testing"
	"time"
)

func TestGetEntityType(t *testing.T) {
	var entityVal interface{}
	iVal := 123

	entityVal = iVal
	eTypeVal, eTypeErr := GetEntityType(entityVal)
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

	typePtr, typeErr := getEntityType(entityVal)
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

	eVal, eErr := eTypeVal.Interface(entityVal)
	if eErr != nil {
		t.Errorf("eTypeVal.Interface failed, error:%s", eErr.Error())
		return
	}
	iVal2, iVal2OK := eVal.Interface().Value().(int)
	if !iVal2OK || iVal2 != iVal {
		t.Errorf("eVal.Interface().(int) failed")
		return
	}

	dt := time.Now()
	entityVal = &dt
	eTypeVal, eTypeErr = GetEntityType(entityVal)
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
	eVal, eErr = eTypeVal.Interface(entityVal)
	if eErr != nil {
		t.Errorf("eTypeVal.Interface failed, error:%s", eErr.Error())
		return
	}
	strVal, strValOK := eVal.Interface().Value().(string)
	if !strValOK || strVal != dt.Format(fu.CSTLayout) {
		t.Errorf("eVal.Interface().(string) failed")
		return
	}
}

func TestGetEntityValue(t *testing.T) {
	var entityVal interface{}
	iVal := 123

	entityVal = iVal
	eTypeVal, eTypeErr := GetEntityType(entityVal)
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

	eValueVal, eValueErr := GetEntityValue(entityVal)
	if eValueErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eValueErr.Error())
		return
	}

	encodeVal, encodeErr := _codec.Encode(eValueVal, eTypeVal)
	if encodeErr != nil {
		t.Errorf("_codec.Encode failed, error:%s", encodeErr.Error())
		return
	}
	switch encodeVal.Value().(type) {
	case int:
		t.Logf("%d", encodeVal.Value().(int))
	default:
		t.Errorf("_codec.Encode failed")
		return
	}

	dt := time.Now()
	entityVal = dt
	eTypeVal, eTypeErr = GetEntityType(entityVal)
	if eTypeErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eTypeErr.Error())
		return
	}
	if !eTypeVal.IsBasic() || eTypeVal.IsPtrType() {
		t.Errorf("GetEntityType failed")
		return
	}
	if eTypeVal.GetValue() != model.TypeDateTimeValue {
		t.Errorf("GetEntityType failed")
		return
	}
	eValueVal, eValueErr = GetEntityValue(entityVal)
	if eValueErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eValueErr.Error())
		return
	}

	encodeVal, encodeErr = _codec.Encode(eValueVal, eTypeVal)
	if encodeErr != nil {
		t.Errorf("_codec.Encode failed, error:%s", encodeErr.Error())
		return
	}
	switch encodeVal.Value().(type) {
	case string:
		t.Logf("%s", encodeVal.Value().(string))
	default:
		t.Errorf("_codec.Encode failed")
		return
	}
}

func TestRemoteModel(t *testing.T) {
	desc := "abc"
	sPtr := &Simple{
		ID:   100,
		Desc: &desc,
		Add:  []int{},
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
	switch eValueVal.Interface().Value().(type) {
	case *ObjectValue:
		valuePtr = eValueVal.Interface().Value().(*ObjectValue)
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
	switch eVal.Value().(type) {
	case int64:
		t.Logf("%d", eVal.Value().(int64))
	default:
		t.Errorf("EncodeValue failed")
	}
}

func TestAppendSliceValue(t *testing.T) {
	iSliceVal := []int{}

	iSliceValueVal, iSliceValueErr := GetEntityValue(iSliceVal)
	if iSliceValueErr != nil {
		t.Errorf("GetEntityValue(iSliceVal) failed, error:%s", iSliceValueErr.Error())
		return
	}

	iVal := 123
	iValueVal, iValueErr := GetEntityValue(iVal)
	if iValueErr != nil {
		t.Errorf("GetEntityValue(iVal) failed, error:%s", iValueErr.Error())
		return
	}

	iSliceValueVal, iSliceValueErr = AppendSliceValue(iSliceValueVal, iValueVal)
	if iSliceValueErr != nil {
		t.Errorf("AppendSliceValue(iSliceValueVal, iValueVal) failed, error:%s", iSliceValueErr.Error())
		return
	}
	switch iSliceValueVal.Interface().Value().(type) {
	case []int:
		t.Logf("%v", iSliceValueVal.Interface())
	default:
		t.Errorf("AppendSliceValue failed")
		return
	}

	iSliceValueVal, iSliceValueErr = GetEntityValue(iSliceVal)
	if iSliceValueErr != nil {
		t.Errorf("GetEntityValue(iSliceVal) failed, error:%s", iSliceValueErr.Error())
		return
	}

	iSliceValueVal, iSliceValueErr = AppendSliceValue(iSliceValueVal, iValueVal)
	if iSliceValueErr != nil {
		t.Errorf("AppendSliceValue(iSliceValueVal, iValueVal) failed, error:%s", iSliceValueErr.Error())
		return
	}
	switch iSliceValueVal.Interface().Value().(type) {
	case []int:
		t.Logf("%+v", iSliceValueVal.Interface())
	default:
		t.Errorf("AppendSliceValue failed")
		return
	}

}
