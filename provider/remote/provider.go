package remote

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

var _helper helper.Helper

func init() {
	_helper = helper.New(ElemDependValue)
}

func isRemoteType(vType model.Type) bool {
	switch vType.GetValue() {
	case util.TypeDoubleField, util.TypeBooleanField, util.TypeStringField:
		return true
	}

	return false
}

func GetHelper() helper.Helper {
	return _helper
}

func GetEntityType(entity interface{}) (ret model.Type, err error) {
	objPtr, ok := entity.(*Object)
	if ok {
		impl := &TypeImpl{Name: objPtr.GetName(), Value: util.TypeStructField, PkgPath: objPtr.GetPkgPath(), IsPtr: objPtr.IsPtr}
		impl.ElemType = &TypeImpl{Name: objPtr.GetName(), Value: util.TypeStructField, PkgPath: objPtr.GetPkgPath(), IsPtr: objPtr.IsPtr}

		ret = impl
		return
	}

	valPtr, ok := entity.(*ObjectValue)
	if ok {
		impl := &TypeImpl{Name: valPtr.GetName(), Value: util.TypeStructField, PkgPath: valPtr.GetPkgPath(), IsPtr: valPtr.IsPtr}
		impl.ElemType = &TypeImpl{Name: valPtr.GetName(), Value: util.TypeStructField, PkgPath: valPtr.GetPkgPath(), IsPtr: valPtr.IsPtr}

		ret = impl
		return
	}

	sValPtr, ok := entity.(*SliceObjectValue)
	if ok {
		impl := &TypeImpl{Name: sValPtr.GetName(), Value: util.TypeSliceField, PkgPath: sValPtr.GetPkgPath(), IsPtr: sValPtr.IsPtr}
		impl.ElemType = &TypeImpl{Name: sValPtr.GetName(), Value: util.TypeStructField, PkgPath: sValPtr.GetPkgPath(), IsPtr: sValPtr.IsElemPtr}

		ret = impl
		return
	}

	typeImpl, typeErr := newType(reflect.TypeOf(entity))
	if typeErr != nil {
		err = typeErr
		return
	}
	if !isRemoteType(typeImpl.Elem()) {
		err = fmt.Errorf("illegal entity type, name:%s", typeImpl.GetName())
		return
	}

	ret = typeImpl
	return
}

func GetEntityValue(entity interface{}) (ret model.Value, err error) {
	rVal := reflect.ValueOf(entity)
	ret = newValue(rVal)
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
	rVal := vVal.Get()
	nameVal := rVal.FieldByName("Name")
	pkgVal := rVal.FieldByName("PkgPath")
	itemsVal := rVal.FieldByName("Items")
	if util.IsNil(nameVal) || util.IsNil(pkgVal) || util.IsNil(itemsVal) {
		err = fmt.Errorf("illegal model value")
		return
	}
	if nameVal.String() != vModel.GetName() || pkgVal.String() != vModel.GetPkgPath() {
		err = fmt.Errorf("illegal model value, mismatch model value")
		return
	}

	for idx := 0; idx < itemsVal.Len(); idx++ {
		iVal := reflect.Indirect(itemsVal.Index(idx))

		iName := iVal.FieldByName("Name").String()
		iValue := iVal.FieldByName("Value")

		vField := vModel.GetField(iName)
		if vField == nil || util.IsNil(iValue) {
			continue
		}

		vType := vField.GetType()
		if vType.IsBasic() {
			vValue, vErr := _helper.Decode(iValue.Interface(), vField.GetType())
			if vErr != nil {
				err = vErr
				return
			}

			err = vField.SetValue(vValue)
			if err != nil {
				return
			}

			continue
		}

		vValue := newValue(iValue)
		err = vField.SetValue(vValue)
		if err != nil {
			return
		}
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err error) {
	rVal := vVal.Get()
	if rVal.Kind() == reflect.Slice {
		for idx := 0; idx < rVal.Len(); idx++ {
			ret = append(ret, newValue(rVal.Index(idx)))
		}

		return
	}

	if rVal.Type().String() == reflect.TypeOf(_declareObjectSliceValue).String() {
		objectsVal := rVal.FieldByName("Values")
		if !util.IsNil(objectsVal) {
			for idx := 0; idx < objectsVal.Len(); idx++ {
				ret = append(ret, newValue(objectsVal.Index(idx)))
			}

			return
		}
	}

	if rVal.Type().String() == reflect.TypeOf(_declareObjectValue).String() {
		itemsVal := rVal.FieldByName("Items")
		if !util.IsNil(itemsVal) {
			ret = append(ret, vVal)
			return
		}

	}

	err = fmt.Errorf("illegal remote slice value, type:%s", rVal.Type().String())
	return
}

func AppendSliceValue(sliceVal model.Value, vVal model.Value) (ret model.Value, err error) {
	sliceName := sliceVal.Get().FieldByName("Name").String()
	slicePkg := sliceVal.Get().FieldByName("PkgPath").String()
	objectName := vVal.Get().FieldByName("Name").String()
	objectPkg := vVal.Get().FieldByName("PkgPath").String()

	if sliceName != objectName || slicePkg != objectPkg {
		err = fmt.Errorf("mismatch slice value")
		return
	}

	sliceObjects := sliceVal.Get().FieldByName("Values")
	if util.IsNil(sliceObjects) {
		err = fmt.Errorf("illegal remote model slice value")
		return
	}

	sliceObjects = reflect.Append(sliceObjects, vVal.Get().Addr())
	sliceVal.Get().FieldByName("Values").Set(sliceObjects)

	ret = sliceVal
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache, helper helper.Helper) (ret interface{}, err error) {
	tModel := mCache.Fetch(vType.GetName())
	if tModel == nil {
		err = fmt.Errorf("illegal value type,type:%s", vType.GetName())
		return
	}

	vModel, vErr := SetModelValue(tModel.Copy(), vVal)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	tType := pkField.GetType()
	tVal := pkField.GetValue()
	if tVal.IsNil() {
		tVal, _ = tType.Interface()
	}

	ret, err = helper.Encode(tVal, tType)
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

	dataVal, dataErr := json.Marshal(items)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret = string(dataVal)
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
