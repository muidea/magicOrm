package local

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
	"github.com/muidea/magicOrm/utils"
)

var _codec util.Codec

func init() {
	_codec = util.New(ElemDependValue)
}

func GetType(vType reflect.Type) (ret model.Type, err *cd.Result) {
	ret, err = NewType(vType)
	return
}

func GetEntityType(entity interface{}) (ret model.Type, err *cd.Result) {
	if entity == nil {
		err = cd.NewResult(cd.IllegalParam, "nil entity value")
		return
	}

	rVal := reflect.ValueOf(entity)
	if !rVal.IsValid() {
		err = cd.NewResult(cd.IllegalParam, "invalid entity value")
		return
	}

	if rVal.Kind() == reflect.Ptr && rVal.IsNil() {
		err = cd.NewResult(cd.IllegalParam, "nil pointer entity value")
		return
	}

	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}

	if !rVal.IsValid() {
		err = cd.NewResult(cd.IllegalParam, "invalid entity value after dereference")
		return
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
	if !utils.IsReallyValid(entity) {
		err = cd.NewResult(cd.IllegalParam, "entity is invalid")
		return
	}
	rVal := reflect.ValueOf(entity)
	ret = NewValue(rVal)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err *cd.Result) {
	if !utils.IsReallyValid(entity) {
		err = cd.NewResult(cd.IllegalParam, "entity is invalid")
		return
	}

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
		err = cd.NewResult(cd.UnExpected, "illegal entity, must be a struct entity")
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

func GetModelFilter(vModel model.Model, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Result) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true, viewSpec)))
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

	log.Infof("SetModelValue: %v", vVal.Get().(reflect.Value).Type().String())
	rVal := reflect.Indirect(vVal.Get().(reflect.Value))
	vType, vErr := NewType(rVal.Type())
	if vErr != nil {
		err = vErr
		return
	}

	if !model.IsStructType(vType.GetValue()) || vType.GetPkgKey() != vModel.GetPkgKey() {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal model value, mode PkgKey:%s, value PkgKey:%s", vModel.GetPkgKey(), vType.GetPkgKey()))
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
			if fieldVal.IsValid() {
				continue
			}
		*/
		vModel.SetFieldValue(fieldName, fieldVal)
	}

	ret = vModel
	return
}

func ElemDependValue(eVal model.RawVal) (ret []model.Value, err *cd.Result) {
	vVal := reflect.ValueOf(eVal.Value())
	if vVal.Kind() == reflect.Interface {
		vVal = vVal.Elem()
	}
	rVal := reflect.Indirect(vVal)
	if rVal.Kind() == reflect.Struct {
		ret = append(ret, NewValue(vVal))
		return
	}

	if rVal.Kind() != reflect.Slice {
		err = cd.NewResult(cd.UnExpected, "illegal slice value")
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		val := NewValue(rVal.Index(idx))
		ret = append(ret, val)
	}
	return
}

func AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err *cd.Result) {
	rSliceVal := sliceVal.Get().(reflect.Value)
	rSliceIndirectVal := reflect.Indirect(rSliceVal)
	riSliceIndirectType := rSliceIndirectVal.Type()
	if riSliceIndirectType.Kind() != reflect.Slice {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("append slice value failed, illegal slice value, slice type:%s", riSliceIndirectType.String()))
		return
	}

	rVal := val.Get().(reflect.Value)
	rType := rVal.Type()
	if riSliceIndirectType.Elem().String() != rType.String() {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("append slice value failed, illegal slice item value, slice type:%s, item type:%s", riSliceIndirectType.String(), rType.String()))
		return
	}
	rSliceNewVal := reflect.New(riSliceIndirectType).Elem()
	rSliceIndirectVal = reflect.Append(rSliceIndirectVal, rVal)
	rSliceNewVal.Set(rSliceIndirectVal)
	if rSliceVal.Kind() == reflect.Pointer {
		rSliceNewVal = rSliceNewVal.Addr()
	}
	ret = NewValue(rSliceNewVal)
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache) (ret model.RawVal, err *cd.Result) {
	if mCache == nil {
		err = cd.NewResult(cd.IllegalParam, "nil model cache parameter")
		log.Errorf("encodeModel failed, err:%s", err.Error())
		return
	}

	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal model type,type:%s", vType.GetName()))
		log.Errorf("encodeModel failed, err:%s", err.Error())
		return
	}

	if vVal.IsBasic() {
		err = cd.NewResult(cd.UnExpected, "illegal model value")
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
	if !tVal.IsValid() {
		tVal, _ = tType.Interface(nil)
	}

	ret, err = _codec.Encode(tVal, tType)
	return
}

func encodeSliceModel(tVal model.Value, tType model.Type, mCache model.Cache) (ret model.RawVal, err *cd.Result) {
	vVals, vErr := ElemDependValue(tVal.Interface())
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

		items = append(items, mVal.Value())
	}

	ret = model.NewRawVal(items)
	return
}

func EncodeValue(tVal model.Value, tType model.Type, mCache model.Cache) (ret model.RawVal, err *cd.Result) {
	if mCache == nil {
		err = cd.NewResult(cd.IllegalParam, "nil model cache parameter")
		log.Errorf("EncodeValue failed, err:%s", err.Error())
		return
	}

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

func decodeModel(eVal model.RawVal, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	tModel := mCache.Fetch(tType.GetPkgKey())
	if tModel == nil {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal value type,type:%s", tType.GetName()))
		return
	}

	mVal, mErr := GetEntityValue(eVal.Value())
	if mErr != nil {
		err = mErr
		return
	}

	var vErr *cd.Result
	vModel := tModel.Copy(true)
	if mVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		pkVal, pkErr := _codec.Decode(eVal, pkField.GetType())
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

	ret, err = GetEntityValue(vModel.Interface(tType.IsPtrType(), model.OriginView))
	return
}

func decodeSliceModel(eVal model.RawVal, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	tModel := mCache.Fetch(tType.GetPkgKey())
	if tModel == nil {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal value type,type:%s", tType.GetName()))
		return
	}

	log.Infof("decodeSliceModel: %v", eVal.Value().(reflect.Value).Type().String())
	mVals, mErr := ElemDependValue(eVal)
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
			log.Infof("SetModelValue: %v", val.Get().(reflect.Value).Type().String())
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

func DecodeValue(eVal model.RawVal, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	if tType.IsBasic() {
		ret, err = _codec.Decode(eVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = decodeModel(eVal, tType, mCache)
		return
	}

	ret, err = decodeSliceModel(eVal, tType, mCache)
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
