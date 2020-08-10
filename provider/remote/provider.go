package remote

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
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
	return
}

func SetModel(m model.Model, v reflect.Value) (ret model.Model, err error) {
	return
}

var _referenceVal ObjectValue
var _referenceType = reflect.TypeOf(_referenceVal)

func getValueModel(val reflect.Value) (ret *Object, err error) {
	return
}
