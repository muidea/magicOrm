package local

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func GetType(v reflect.Value) (ret model.Type, err error) {
	vType, vErr := newType(v.Type())
	if vErr != nil {
		err = vErr
		return
	}

	ret = vType
	return
}

func GetValue(v reflect.Value) (ret model.Value, err error) {
	vVal, vErr := newValue(v)
	if vErr != nil {
		err = vErr
		return
	}

	ret = vVal
	return
}

func GetModel(v reflect.Value) (ret model.Model, err error) {
	vType, vErr := GetType(v)
	if vErr != nil {
		err = vErr
		return
	}

	// if slice value, elem slice item
	if util.IsSliceType(vType.GetValue()) {
		ret, err = getTypeModel(v.Type().Elem())
		return
	}

	ret, err = getTypeModel(v.Type())
	return
}

func SetModel(modelInfo model.Model, modelVal reflect.Value) (ret model.Model, err error) {
	modelVal = reflect.Indirect(modelVal)
	modelType := modelVal.Type()
	fieldNum := modelType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldVal := modelVal.Field(idx)
		if util.IsNil(fieldVal) {
			continue
		}

		err = modelInfo.SetFieldValue(idx, fieldVal)
		if err != nil {
			return
		}
	}

	ret = modelInfo
	return
}
