package local

import (
	"fmt"
	"github.com/muidea/magicOrm/util"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

func GetEntityType(entity interface{}) (ret model.Type, err error) {
	rType := reflect.TypeOf(entity)
	vType, vErr := newType(rType)
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
	rVal := vVal.Get().(reflect.Value)
	rVal = reflect.Indirect(rVal)
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
	rVal := vVal.Get().(reflect.Value)
	rVal = reflect.Indirect(rVal)
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
	rSliceVal := sliceVal.Get().(reflect.Value)
	isPtr := rSliceVal.Kind() == reflect.Ptr
	rSliceVal = reflect.Indirect(rSliceVal)
	rSliceType := rSliceVal.Type()
	if rSliceType.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice value, slice type:%s", rSliceType.String())
		return
	}

	rVal := val.Get().(reflect.Value)
	rType := rVal.Type()
	if rSliceType.Elem().String() != rType.String() {
		err = fmt.Errorf("illegal slice item value, slice type:%s, item type:%s", rSliceType.String(), rType.String())
		return
	}

	rSliceVal = reflect.Append(rSliceVal, rVal)
	if isPtr {
		rSliceVal = rSliceVal.Addr()
	}

	ret.Set(rSliceVal)
	return
}
