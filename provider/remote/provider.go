package remote

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func GetEntityType(entity interface{}) (ret model.Type, err error) {
	objPtr, ok := entity.(*Object)
	if ok {
		impl := &TypeImpl{Name: objPtr.GetName(), Value: util.TypeStructField, PkgPath: objPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objPtr.GetName(), Value: util.TypeStructField, PkgPath: objPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	valPtr, ok := entity.(*ObjectValue)
	if ok {
		impl := &TypeImpl{Name: valPtr.GetName(), Value: util.TypeStructField, PkgPath: valPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: valPtr.GetName(), Value: util.TypeStructField, PkgPath: valPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	sValPtr, ok := entity.(*SliceObjectValue)
	if ok {
		impl := &TypeImpl{Name: sValPtr.GetName(), Value: util.TypeSliceField, PkgPath: sValPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: sValPtr.GetName(), Value: util.TypeStructField, PkgPath: sValPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	err = fmt.Errorf("illegal entity")

	return
}

func GetEntityValue(entity interface{}) (ret model.Value, err error) {
	valPtr, ok := entity.(*ObjectValue)
	if ok {
		ret = newValue(valPtr)
		return
	}

	sliceValPtr, ok := entity.(*SliceObjectValue)
	if ok {
		ret = newValue(sliceValPtr)
		return
	}

	ret = newValue(entity)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err error) {
	objPtr, ok := entity.(*Object)
	if ok {
		ret = objPtr
		return
	}

	err = fmt.Errorf("illegal value, not object")
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err error) {
	val, ok := vVal.Get().(*ObjectValue)
	if !ok {
		err = fmt.Errorf("illegal remote model value")
		return
	}

	for _, item := range val.Items {
		if item.Value == nil {
			continue
		}

		err = vModel.SetFieldValue(item.Name, newValue(item.Value))
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err error) {
	val, ok := vVal.Get().(*SliceObjectValue)
	if !ok {
		err = fmt.Errorf("illegal remote model slice value")
		return
	}

	for _, item := range val.Values {
		ret = append(ret, newValue(item))
	}

	return
}

func AppendSliceValue(sliceVal model.Value, vVal model.Value) (ret model.Value, err error) {
	sVal, ok := sliceVal.Get().(*SliceObjectValue)
	if !ok {
		err = fmt.Errorf("illegal remote model slice value")
		return
	}

	val, ok := vVal.Get().(*ObjectValue)
	if !ok {
		err = fmt.Errorf("illegal remote model item value")
		return
	}

	if sVal.GetName() != val.GetName() || sVal.GetPkgPath() != val.GetPkgPath() {
		err = fmt.Errorf("mismatch slice value")
		return
	}

	sVal.Values = append(sVal.Values, val)
	ret = newValue(sVal)
	return
}
