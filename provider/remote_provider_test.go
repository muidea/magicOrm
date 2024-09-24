package provider

import (
	"testing"
	"time"

	fu "github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/provider/util"
)

func TestNewRemoteProvider(t *testing.T) {
	s001 := &Simple{}
	provider := NewRemoteProvider("t001")

	s001ObjectPtr, s001ObjectErr := remote.GetObject(s001)
	if s001ObjectErr != nil {
		t.Errorf("%s", s001ObjectErr.Error())
		return
	}

	s001ObjectValuePtr, s001ObjectValueErr := remote.GetObjectValue(s001)
	if s001ObjectValueErr != nil {
		t.Errorf("%s", s001ObjectValueErr.Error())
		return
	}

	sModelVal, sModelErr := provider.RegisterModel(s001ObjectPtr)
	if sModelErr != nil {
		t.Errorf("%s", sModelErr.Error())
		return
	}

	sModelVal, sModelErr = provider.GetEntityModel(s001ObjectValuePtr)
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
	s001ObjectPtr, s001ObjectErr = remote.GetObject(s001)
	if s001ObjectErr != nil {
		t.Errorf("%s", s001ObjectErr.Error())
		return
	}

	s001ObjectValuePtr, s001ObjectValueErr = remote.GetObjectValue(s001)
	if s001ObjectValueErr != nil {
		t.Errorf("%s", s001ObjectValueErr.Error())
		return
	}

	sValueVal, sValueErr := provider.GetEntityValue(s001ObjectValuePtr)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr.Error())
		return
	}

	sTypeVal, sTypeErr := provider.GetEntityType(s001ObjectPtr)
	if sTypeErr != nil {
		t.Errorf("%s", sTypeErr.Error())
		return
	}

	_, sFilterErr := provider.GetEntityFilter(s001ObjectPtr, model.LiteView)
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
	s001NewObjectValuePtr := sModelVal.Interface(true, model.OriginView).(*remote.ObjectValue)
	if !util.IsSameValue(s001ObjectValuePtr, s001NewObjectValuePtr) {
		t.Errorf("get value model failed")
		return
	}

	s001NewObjectValuePtr.SetFieldValue("id", 987654)
	s001NewObjectValuePtr.ID = "987654"
	sValueVal, sValueErr = provider.DecodeValue(s001NewObjectValuePtr, sTypeVal)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr)
		return
	}
	s001NewObjectValue2Ptr := sValueVal.Interface().(*remote.ObjectValue)
	if !util.IsSameValue(s001NewObjectValuePtr, s001NewObjectValue2Ptr) {
		t.Errorf("decode value model failed")
		return
	}

	sValueVal, sValueErr = provider.DecodeValue(23456, sTypeVal)
	if sValueErr != nil {
		t.Errorf("%s", sValueErr)
		return
	}
	s001NewObjectValue3Ptr := sValueVal.Interface().(*remote.ObjectValue)
	if s001NewObjectValue3Ptr.GetFieldValue("id").(int) != 23456 {
		t.Errorf("decode value model failed")
		return
	}

	dt001 := time.Now().UTC()
	dt001Str := dt001.Format(fu.CSTLayout)

	dtValueVal, dtValueErr := provider.GetEntityValue(dt001Str)
	if dtValueErr != nil {
		t.Errorf("%s", dtValueErr.Error())
		return
	}

	dtTypeVal, dtTypeErr := provider.GetEntityType(dt001Str)
	if dtTypeErr != nil {
		t.Errorf("%s", dtTypeErr.Error())
		return
	}
	if dtTypeVal.IsPtrType() {
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

	dtValueVal, dtValueErr = provider.DecodeValue(eVal, dtTypeVal)
	if dtValueErr != nil {
		t.Errorf("%s", dtValueErr.Error())
		return
	}

	dVal := dtValueVal.Interface()
	switch dVal.(type) {
	case string:
	default:
		t.Errorf("decodeValue failed")
	}

	dtDecodeVal := dtValueVal.Interface().(string)
	if dtDecodeVal != dt001Str {
		t.Errorf("illegal decode value")
	}

	dtSlice := []string{}
	dtSliceValueVal, dtSliceValueErr := provider.GetEntityValue(dtSlice)
	if dtSliceValueErr != nil {
		t.Errorf("%s", dtSliceValueErr.Error())
		return
	}
	dtSliceTypeVal, dtSliceTypeErr := provider.GetEntityType(dtSlice)
	if dtSliceTypeErr != nil {
		t.Errorf("%s", dtSliceTypeErr.Error())
		return
	}

	dtSliceValueVal, dtSliceValueErr = provider.AppendSliceValue(dtSliceValueVal, dtValueVal)
	if dtSliceValueErr != nil {
		t.Errorf("%s", dtSliceValueErr.Error())
		return
	}
	dtSliceValueVal, dtSliceValueErr = provider.AppendSliceValue(dtSliceValueVal, dtValueVal)
	if dtSliceValueErr != nil {
		t.Errorf("%s", dtSliceValueErr.Error())
		return
	}

	eVal, eErr = provider.EncodeValue(dtSliceValueVal, dtSliceTypeVal)
	if eErr != nil {
		t.Errorf("%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case []interface{}:
		t.Logf("%v", eVal.([]interface{}))
	default:
		t.Errorf("illegal value")
		return
	}

	eVal = []string{"2024-09-21 03:12:31", "2024-09-21 03:12:31"}
	dtSliceValueVal, dtSliceValueErr = provider.DecodeValue(eVal, dtSliceTypeVal)
	if dtSliceValueErr != nil {
		t.Errorf("%s", dtSliceValueErr.Error())
		return
	}
	dtSlice = dtSliceValueVal.Interface().([]string)
	if len(dtSlice) != 2 {
		t.Errorf("decodeValue failed")
	}

	provider.UnregisterModel(s001)
}
