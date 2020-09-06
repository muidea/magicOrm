package local

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func GetType(val reflect.Value) (ret model.Type, err error) {
	vType, vErr := newType(val.Type())
	if vErr != nil {
		err = vErr
		return
	}

	ret = vType
	return
}

func GetValue(val reflect.Value) (ret model.Value, err error) {
	vVal, vErr := newValue(val)
	if vErr != nil {
		err = vErr
		return
	}

	ret = vVal
	return
}

func GetModel(vVal reflect.Value) (ret model.Model, err error) {
	vType, vErr := GetType(vVal)
	if vErr != nil {
		err = vErr
		return
	}

	// if slice value, elem slice item
	if util.IsSliceType(vType.GetValue()) {
		ret, err = getTypeModel(vVal.Type().Elem())
		return
	}

	ret, err = getTypeModel(vVal.Type())
	return
}

func SetModel(vModel model.Model, vVal reflect.Value) (ret model.Model, err error) {
	vVal = reflect.Indirect(vVal)
	vType := vVal.Type()
	fieldNum := vType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldVal := vVal.Field(idx)
		if util.IsNil(fieldVal) {
			continue
		}

		err = vModel.SetFieldValue(idx, fieldVal)
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}

func ElemDependValue(vType model.Type, val reflect.Value) (ret []model.Value, err error) {
	if vType.GetValue() == util.TypeSliceField {
		for idx := 0; idx < val.Len(); idx++ {
			vVal, vErr := newValue(val.Index(idx))
			if vErr != nil {
				err = vErr
				return
			}

			ret = append(ret, vVal)
		}

		return
	}

	vVal, vErr := newValue(val)
	if vErr != nil {
		err = vErr
		return
	}
	ret = append(ret, vVal)

	return
}
