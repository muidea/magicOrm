package local

import (
	"fmt"
	"github.com/muidea/magicCommon/foundation/log"
	"reflect"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local/codec"
	"github.com/muidea/magicOrm/provider/util"
)

var _codec codec.Codec

func init() {
	_codec = codec.New(ElemDependValue)
}

func GetType(vType reflect.Type) (ret model.Type, err *cd.Result) {
	ret, err = NewType(vType)
	return
}

func GetEntityType(entity interface{}) (ret model.Type, err *cd.Result) {
	rVal := reflect.ValueOf(entity)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	vType, vErr := getValueType(rVal)
	if vErr != nil {
		err = vErr
		return
	}

	ret = vType
	return
}

func GetEntityValue(entity interface{}) (ret model.Value, err *cd.Result) {
	rVal := reflect.ValueOf(entity)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}

	ret = NewValue(rVal)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err *cd.Result) {
	rVal := reflect.ValueOf(entity)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}

	vType, vErr := NewType(rVal.Type())
	if vErr != nil {
		err = vErr
		return
	}
	if !model.IsStructType(vType.GetValue()) {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal entity, must be a struct entity"))
		return
	}

	implPtr, implErr := getValueModel(rVal)
	if implErr != nil {
		err = implErr
		return
	}

	ret = implPtr
	return
}

func GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Result) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true, model.OriginView)))
	ret = newFilter(valuePtr)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err *cd.Result) {
	if vVal.IsZero() {
		ret = vModel
		return
	}
	if vVal.IsBasic() {
		pkField := vModel.GetPrimaryField()
		pkField.SetValue(vVal)
		return
	}

	rVal := reflect.Indirect(vVal.Get().(reflect.Value))
	vType, vErr := NewType(rVal.Type())
	if vErr != nil {
		err = vErr
		return
	}

	if !model.IsStructType(vType.GetValue()) || vType.GetPkgKey() != vModel.GetPkgKey() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model value, mode PkgKey:%s, value PkgKey:%s", vModel.GetPkgKey(), vType.GetPkgKey()))
		return
	}

	rType := vType.getRawType()
	fieldNum := rType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := rType.Field(idx)
		fieldName, fieldErr := getFieldName(fieldType)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		fieldVal := NewValue(rVal.Field(idx))
		/*
			if fieldVal.IsNil() {
				continue
			}
		*/
		vModel.SetFieldValue(fieldName, fieldVal)
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err *cd.Result) {
	rVal := reflect.Indirect(vVal.Get().(reflect.Value))
	if rVal.Kind() == reflect.Struct {
		ret = append(ret, vVal)
		return
	}

	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal slice value"))
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		val := NewValue(rVal.Index(idx))
		ret = append(ret, val)
	}
	return
}

func AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err *cd.Result) {
	// *[]xx , []xx
	rSliceVal := sliceVal.Get().(reflect.Value)
	riSliceVal := reflect.Indirect(rSliceVal)
	riSliceType := riSliceVal.Type()
	if riSliceType.Kind() != reflect.Slice {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("append slice value failed, illegal slice value, slice type:%s", riSliceType.String()))
		return
	}

	isElemPtr := false
	rElemType := riSliceType.Elem()
	if rElemType.Kind() == reflect.Ptr {
		isElemPtr = true
	}

	rVal := val.Get().(reflect.Value)
	if !isElemPtr {
		rVal = reflect.Indirect(rVal)
	}

	rType := rVal.Type()
	if riSliceType.Elem().String() != rType.String() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("append slice value failed, illegal slice item value, slice type:%s, item type:%s", riSliceType.String(), rType.String()))
		return
	}

	rNewVal := reflect.Append(riSliceVal, rVal)
	riSliceVal.Set(rNewVal)

	ret = NewValue(rSliceVal)
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache) (ret interface{}, err *cd.Result) {
	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model type,type:%s", vType.GetName()))
		log.Errorf("encodeModel failed, err:%s", err.Error())
		return
	}

	if vVal.IsBasic() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model value"))
		log.Errorf("encodeModel failed, err:%s", err.Error())
		return
	}

	vModel, vErr := SetModelValue(tModel.Copy(true), vVal)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	tType := pkField.GetType()
	tVal := pkField.GetValue()
	if tVal.IsNil() {
		tVal, _ = tType.Interface(nil)
	}

	ret, err = _codec.Encode(tVal, tType)
	return
}

func encodeSliceModel(tVal model.Value, tType model.Type, mCache model.Cache) (ret interface{}, err *cd.Result) {
	vVals, vErr := ElemDependValue(tVal)
	if vErr != nil {
		err = vErr
		return
	}
	if len(vVals) == 0 {
		return
	}

	items := []interface{}{}
	for _, v := range vVals {
		mVal, mErr := encodeModel(v, tType.Elem(), mCache)
		if mErr != nil {
			err = mErr
			return
		}

		items = append(items, mVal)
	}

	ret = items
	return
}

func EncodeValue(tVal model.Value, tType model.Type, mCache model.Cache) (ret interface{}, err *cd.Result) {
	if tType.IsBasic() {
		ret, err = _codec.Encode(tVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = encodeModel(tVal, tType, mCache)
		return
	}

	ret, err = encodeSliceModel(tVal, tType, mCache)
	return
}

func decodeModel(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	tModel := mCache.Fetch(tType.GetPkgKey())
	if tModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal value type,type:%s", tType.GetName()))
		return
	}

	mVal, mErr := GetEntityValue(tVal)
	if mErr != nil {
		err = mErr
		return
	}

	var vErr *cd.Result
	vModel := tModel.Copy(true)
	if mVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		pkVal, pkErr := _codec.Decode(tVal, pkField.GetType())
		if pkErr != nil {
			err = pkErr
			return
		}

		vModel.SetPrimaryFieldValue(pkVal)
	} else {
		vModel, vErr = SetModelValue(vModel, mVal)
	}

	if vErr != nil {
		err = vErr
		return
	}

	tVal = vModel.Interface(tType.IsPtrType(), model.OriginView)
	ret, err = GetEntityValue(tVal)
	return
}

func decodeSliceModel(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	tModel := mCache.Fetch(tType.GetPkgKey())
	if tModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal value type,type:%s", tType.GetName()))
		return
	}
	mVal, mErr := GetEntityValue(tVal)
	if mErr != nil {
		err = mErr
		return
	}

	var mVals []model.Value
	mVals, mErr = ElemDependValue(mVal)
	if mErr != nil {
		err = mErr
		return
	}

	var vErr *cd.Result
	vVals, _ := tType.Interface(nil)
	for _, val := range mVals {
		vModel := tModel.Copy(true)
		if val.IsBasic() {
			pkField := tModel.GetPrimaryField()
			pkVal, pkErr := _codec.Decode(val.Interface(), pkField.GetType())
			if pkErr != nil {
				err = pkErr
				return
			}
			vModel.SetPrimaryFieldValue(pkVal)
		} else {
			vModel, vErr = SetModelValue(vModel, val)
			if vErr != nil {
				err = vErr
				return
			}
		}

		iVal, _ := GetEntityValue(vModel.Interface(tType.Elem().IsPtrType(), model.OriginView))
		vVals, vErr = AppendSliceValue(vVals, iVal)
		if vErr != nil {
			err = vErr
			return
		}
	}

	ret = vVals
	return
}

func DecodeValue(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	if tType.IsBasic() {
		ret, err = _codec.Decode(tVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = decodeModel(tVal, tType, mCache)
		return
	}

	ret, err = decodeSliceModel(tVal, tType, mCache)
	return
}

func GetNewValue(valueDeclare model.ValueDeclare) (ret model.Value) {
	var rVal interface{}
	switch valueDeclare {
	case model.SnowFlake:
		rVal = util.GetNewSnowFlakeID()
	case model.UUID:
		rVal = util.GetNewUUID()
	case model.DateTime:
		rVal = util.GetCurrentDateTime()
	}
	if rVal != nil {
		ret = NewValue(reflect.ValueOf(rVal))
		return
	}

	ret = &NilValue
	return
}
