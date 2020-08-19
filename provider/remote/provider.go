package remote

import (
	"fmt"
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

	if util.IsBasicType(vType.GetValue()) {
		ret = vType
		return
	}

	if util.IsStructType(vType.GetValue()) {
		v := reflect.Indirect(v)
		obj, ok := v.Interface().(Object)
		if ok {
			vType.Name = obj.GetName()
			vType.PkgPath = obj.GetPkgPath()

			ret = vType
			return
		}

		vObj, ok := v.Interface().(ObjectValue)
		if ok {
			vType.Name = vObj.GetName()
			vType.PkgPath = vObj.GetPkgPath()

			ret = vType
			return
		}
	}

	err = fmt.Errorf("illegal value type, type:%s", v.Type().String())
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

func GetModel(v reflect.Value) (ret model.Model, err error) {
	v = reflect.Indirect(v)
	obj, ok := v.Interface().(Object)
	if ok {
		ret = &obj
		return
	}

	err = fmt.Errorf("illegal value, not object")
	return
}

func SetModel(vModel model.Model, vVal reflect.Value) (ret model.Model, err error) {
	vVal = reflect.Indirect(vVal)
	vName := vVal.FieldByName("Name")
	vPkgPath := vVal.FieldByName("PkgPath")
	vFields := vVal.FieldByName("Items")
	if vModel.GetName() != vName.String() || vModel.GetPkgPath() != vPkgPath.String() {
		err = fmt.Errorf("illegal model value")
		return
	}

	for idx := 0; idx < vFields.Len(); {
		item := vFields.Index(idx)
		fValue := item.FieldByName("Value")
		if util.IsNil(fValue) {
			continue
		}

		err = vModel.SetFieldValue(idx, fValue)
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}
