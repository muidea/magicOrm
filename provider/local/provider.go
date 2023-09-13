package local

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local/codec"
	"github.com/muidea/magicOrm/provider/util"
)

var _codec codec.Codec

func init() {
	_codec = codec.New(ElemDependValue)
}

func GetCodec() codec.Codec {
	return _codec
}

func GetType(vType reflect.Type) (ret model.Type, err error) {
	ret, err = NewType(vType)
	return
}

func GetEntityType(entity interface{}) (ret model.Type, err error) {
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

func GetEntityValue(entity interface{}) (ret model.Value, err error) {
	rVal := reflect.ValueOf(entity)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}

	ret = NewValue(rVal)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err error) {
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
		err = fmt.Errorf("illegal entity, must be a struct entity")
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

func GetModelFilter(vModel model.Model) (ret model.Filter, err error) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true)))
	ret = NewFilter(valuePtr)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err error) {
	if vVal.IsZero() {
		ret = vModel
		return
	}

	rVal := reflect.Indirect(vVal.Get().(reflect.Value))
	vType, vErr := NewType(rVal.Type())
	if vErr != nil {
		err = vErr
		return
	}

	if !model.IsStructType(vType.GetValue()) || vType.GetPkgKey() != vModel.GetPkgKey() {
		err = fmt.Errorf("illegal model value, mode PkgKey:%s, value PkgKey:%s", vModel.GetPkgKey(), vType.GetPkgKey())
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
		err = vModel.SetFieldValue(fieldName, fieldVal)
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err error) {
	rVal := reflect.Indirect(vVal.Get().(reflect.Value))
	if rVal.Kind() == reflect.Struct {
		ret = append(ret, vVal)
		return
	}

	if rVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice value")
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		val := NewValue(rVal.Index(idx))
		ret = append(ret, val)
	}

	return
}

func AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err error) {
	// *[]xx , []xx
	rSliceVal := sliceVal.Get().(reflect.Value)
	riSliceVal := reflect.Indirect(rSliceVal)
	riSliceType := riSliceVal.Type()
	if riSliceType.Kind() != reflect.Slice {
		err = fmt.Errorf("append slice value failed, illegal slice value, slice type:%s", riSliceType.String())
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
		err = fmt.Errorf("append slice value failed, illegal slice item value, slice type:%s, item type:%s", riSliceType.String(), rType.String())
		return
	}

	rNewVal := reflect.Append(riSliceVal, rVal)
	riSliceVal.Set(rNewVal)

	ret = NewValue(rSliceVal)
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache, codec codec.Codec) (ret interface{}, err error) {
	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = fmt.Errorf("illegal value type,type:%s", vType.GetName())
		return
	}

	if vVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		ret, err = codec.Encode(vVal, pkField.GetType())
		return
	}

	vModel, vErr := SetModelValue(tModel.Copy(), vVal)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	ret, err = codec.Encode(pkField.GetValue(), pkField.GetType())
	return
}

func encodeSliceModel(tVal model.Value, tType model.Type, mCache model.Cache, codec codec.Codec) (ret string, err error) {
	vVals, vErr := ElemDependValue(tVal)
	if vErr != nil {
		err = vErr
		return
	}
	if len(vVals) == 0 {
		return
	}

	items := []string{}
	for _, v := range vVals {
		strVal, strErr := encodeModel(v, tType.Elem(), mCache, codec)
		if strErr != nil {
			err = strErr
			return
		}

		items = append(items, fmt.Sprintf("%v", strVal))
	}

	ret = strings.Join(items, ",")
	return
}

func EncodeValue(tVal model.Value, tType model.Type, mCache model.Cache) (ret interface{}, err error) {
	if tType.IsBasic() {
		ret, err = _codec.Encode(tVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = encodeModel(tVal, tType, mCache, _codec)
		return
	}

	ret, err = encodeSliceModel(tVal, tType, mCache, _codec)
	return
}

func DecodeValue(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err error) {
	if tType.IsBasic() {
		ret, err = _codec.Decode(tVal, tType)
		return
	}

	err = fmt.Errorf("unexpected type, type name:%s", tType.GetName())
	return
}

func GetValue(valueDeclare model.ValueDeclare) (ret model.Value) {
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
