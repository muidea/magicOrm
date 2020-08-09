package local

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
)

func GetType(v reflect.Type) (ret model.Type, err error) {
	vType, vErr := newType(v)
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

func GetValueModel(v reflect.Value) (ret model.Model, err error) {
	return getValueModel(v)
}

func GetTypeModel(t reflect.Type) (ret model.Model, err error) {
	return getTypeModel(t)
}
