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

	err = fmt.Errorf("illegal value")
	return
}

func GetModel(v reflect.Value) (ret model.Model, err error) {
	obj, ok := v.Interface().(Object)
	if ok {
		ret = &obj
		return
	}

	err = fmt.Errorf("illegal value, not object")
	return
}

func SetModel(m model.Model, v reflect.Value) (ret model.Model, err error) {
	v = reflect.Indirect(v)
	vObj, ok := v.Interface().(ObjectValue)
	if !ok {
		err = fmt.Errorf("illegal model value")
		return
	}

	if m.GetName() != vObj.GetName() || m.GetPkgPath() != vObj.GetPkgPath() {
		err = fmt.Errorf("illegal model value")
		return
	}

	for idx := range vObj.Items {
		item := vObj.Items[idx]
		err = m.SetFieldValue(idx, reflect.ValueOf(item.Value))
		if err != nil {
			return
		}
	}

	ret = m
	return
}
