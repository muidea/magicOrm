package local

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

var _helper helper.Helper

func init() {
	_helper = helper.New(ElemDependValue)
}

func GetHelper() helper.Helper {
	return _helper
}

func GetType(vType reflect.Type) (ret model.Type, err error) {
	ret, err = newType(vType)
	return
}

func GetEntityType(entity interface{}) (ret model.Type, err error) {
	rVal := reflect.ValueOf(entity)
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
	ret = newValue(rVal)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err error) {
	rVal := reflect.ValueOf(entity)
	if rVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("must be a pointer entity")
		return
	}
	rVal = rVal.Elem()

	vType, vErr := newType(rVal.Type())
	if vErr != nil {
		err = vErr
		return
	}
	if !util.IsStructType(vType.GetValue()) {
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

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err error) {
	rVal := vVal.Get()
	vType := rVal.Type()
	if vType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal model value, mode name:%s", vModel.GetName())
		return
	}

	fieldNum := vType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := vType.Field(idx)
		fieldVal := newValue(rVal.Field(idx))
		if fieldVal.IsNil() {
			continue
		}

		err = vModel.SetFieldValue(fieldType.Name, fieldVal)
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err error) {
	rVal := vVal.Get()
	if rVal.Kind() == reflect.Struct {
		ret = append(ret, vVal)
		return
	}

	if rVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice value")
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		val := newValue(rVal.Index(idx))
		ret = append(ret, val)
	}

	return
}

func AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err error) {
	rSliceVal := sliceVal.Get()
	rSliceType := rSliceVal.Type()
	if rSliceType.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice value, slice type:%s", rSliceType.String())
		return
	}

	isElemPtr := false
	rElemType := rSliceType.Elem()
	if rElemType.Kind() == reflect.Ptr {
		isElemPtr = true
	}

	rVal := val.Get()
	if isElemPtr {
		rVal = rVal.Addr()
	}

	rType := rVal.Type()
	if rSliceType.Elem().String() != rType.String() {
		err = fmt.Errorf("illegal slice item value, slice type:%s, item type:%s", rSliceType.String(), rType.String())
		return
	}

	rNewVal := reflect.Append(rSliceVal, rVal)
	rSliceVal.Set(rNewVal)

	ret = newValue(rSliceVal)
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache, helper helper.Helper) (ret interface{}, err error) {
	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = fmt.Errorf("illegal value type,type:%s", vType.GetName())
		return
	}

	if vVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		vType := pkField.GetType()
		ret, err = helper.Encode(vVal, vType)
		return
	}

	vModel, vErr := SetModelValue(tModel.Copy(), vVal)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	ret, err = helper.Encode(pkField.GetValue(), pkField.GetType())
	return
}

func encodeSliceModel(tVal model.Value, tType model.Type, mCache model.Cache, helper helper.Helper) (ret string, err error) {
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
		strVal, strErr := encodeModel(v, tType.Elem(), mCache, helper)
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
		ret, err = _helper.Encode(tVal, tType)
		return
	}
	if util.IsStructType(tType.GetValue()) {
		ret, err = encodeModel(tVal, tType, mCache, _helper)
		return
	}

	ret, err = encodeSliceModel(tVal, tType, mCache, _helper)
	return
}

func DecodeValue(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err error) {
	if tType.IsBasic() {
		ret, err = _helper.Decode(tVal, tType)
		return
	}

	err = fmt.Errorf("unexecption type, type name:%s", tType.GetName())
	return
}
