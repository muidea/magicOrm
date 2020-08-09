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
		return getTypeModel(v.Type().Elem())
	}

	return getValueModel(v)
}
