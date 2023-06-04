package remote

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

var _helper helper.Helper
var nilValue model.Value

func init() {
	_helper = helper.New(ElemDependValue)

	nilValue = &valueImpl{}
}

func isRemoteType(vType model.Type) bool {
	switch vType.GetValue() {
	case util.TypeDoubleValue, util.TypeBooleanValue, util.TypeStringValue:
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
		impl := &TypeImpl{Name: objPtr.GetName(), Value: util.TypeStructValue, PkgPath: objPtr.GetPkgPath(), IsPtr: objPtr.IsPtr}
		impl.ElemType = &TypeImpl{Name: objPtr.GetName(), Value: util.TypeStructValue, PkgPath: objPtr.GetPkgPath(), IsPtr: objPtr.IsPtr}

		ret = impl
		return
	}

	valPtr, ok := entity.(*ObjectValue)
	if ok {
		impl := &TypeImpl{Name: valPtr.GetName(), Value: util.TypeStructValue, PkgPath: valPtr.GetPkgPath(), IsPtr: valPtr.IsPtr}
		impl.ElemType = &TypeImpl{Name: valPtr.GetName(), Value: util.TypeStructValue, PkgPath: valPtr.GetPkgPath(), IsPtr: valPtr.IsPtr}

		ret = impl
		return
	}

	sValPtr, ok := entity.(*SliceObjectValue)
	if ok {
		impl := &TypeImpl{Name: sValPtr.GetName(), Value: util.TypeSliceValue, PkgPath: sValPtr.GetPkgPath(), IsPtr: sValPtr.IsPtr}
		impl.ElemType = &TypeImpl{Name: sValPtr.GetName(), Value: util.TypeStructValue, PkgPath: sValPtr.GetPkgPath(), IsPtr: sValPtr.IsElemPtr}

		ret = impl
		return
	}

	filterPtr, ok := entity.(*ObjectFilter)
	if ok {
		impl := &TypeImpl{Name: filterPtr.GetName(), Value: util.TypeStructValue, PkgPath: filterPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: filterPtr.GetName(), Value: util.TypeStructValue, PkgPath: filterPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	typeImpl, typeErr := newType(reflect.TypeOf(entity))
	if typeErr != nil {
		err = typeErr
		return
	}
	if !isRemoteType(typeImpl.Elem()) {
		err = fmt.Errorf("illegal entity type, name:%s", typeImpl.GetPkgKey())
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
	if !ok {
		err = fmt.Errorf("illegal entity value, not object")
		return
	}

	err = objPtr.verify()
	if err != nil {
		return
	}

	ret = objPtr
	return
}

func GetModelFilter(vModel model.Model) (ret model.Filter, err error) {
	objectImpl, objectOK := vModel.(*Object)
	if !objectOK {
		err = fmt.Errorf("illegal model type")
		return
	}

	ret = NewFilter(objectImpl)
	return
}

func setFieldValue(iVal reflect.Value, vModel model.Model) (err error) {
	iName := iVal.FieldByName("Name").String()
	iValue := iVal.FieldByName("Value")
	if iValue.Kind() == reflect.Interface {
		iValue = iValue.Elem()
	}

	vField := vModel.GetField(iName)
	if util.IsNil(iValue) {
		vField.SetValue(nilValue)
		return
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
		return
	}

	vValue := newValue(iValue)
	err = vField.SetValue(vValue)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err error) {
	rVal := reflect.Indirect(vVal.Get())
	nameVal := rVal.FieldByName("Name")
	pkgVal := rVal.FieldByName("PkgPath")
	itemsVal := rVal.FieldByName("Fields")
	if util.IsNil(nameVal) || util.IsNil(pkgVal) {
		err = fmt.Errorf("illegal model value")
		return
	}
	if nameVal.String() != vModel.GetName() || pkgVal.String() != vModel.GetPkgPath() {
		err = fmt.Errorf("illegal model value, mismatch model value")
		return
	}

	if !util.IsNil(itemsVal) {
		for idx := 0; idx < itemsVal.Len(); idx++ {
			iVal := reflect.Indirect(itemsVal.Index(idx))
			err = setFieldValue(iVal, vModel)
			if err != nil {
				return
			}
		}
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err error) {
	rVal := reflect.Indirect(vVal.Get())
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
		itemsVal := rVal.FieldByName("Fields")
		if !util.IsNil(itemsVal) {
			ret = append(ret, vVal)
			return
		}
	}

	tVal, tErr := util.GetTypeEnum(rVal.Type())
	if tErr != nil {
		err = tErr
		return
	}

	if util.IsSliceType(tVal) {
		for idx := 0; idx < rVal.Len(); idx++ {
			ret = append(ret, newValue(rVal.Index(idx)))
		}
		return
	}

	if util.IsBasicType(tVal) {
		ret = append(ret, vVal)
		return
	}

	err = fmt.Errorf("illegal remote slice value, type:%s", rVal.Type().String())
	return
}

func AppendSliceValue(sliceVal model.Value, vVal model.Value) (ret model.Value, err error) {
	rvVal := reflect.Indirect(vVal.Get())
	sliceName := sliceVal.Get().FieldByName("Name").String()
	slicePkg := sliceVal.Get().FieldByName("PkgPath").String()
	objectName := rvVal.FieldByName("Name").String()
	objectPkg := rvVal.FieldByName("PkgPath").String()

	if sliceName != objectName || slicePkg != objectPkg {
		err = fmt.Errorf("mismatch slice value")
		return
	}

	sliceObjects := sliceVal.Get().FieldByName("Values")
	if util.IsNil(sliceObjects) {
		err = fmt.Errorf("illegal remote model slice value")
		return
	}

	sliceObjects = reflect.Append(sliceObjects, vVal.Get())
	sliceVal.Get().FieldByName("Values").Set(sliceObjects)

	ret = sliceVal
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache, helper helper.Helper) (ret interface{}, err error) {
	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = fmt.Errorf("illegal value type,type:%s", vType.GetName())
		return
	}

	if vVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		vType := pkField.GetType()
		ret, err = helper.Encode(vVal, vType)
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
		tVal = tType.Interface()
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

	ret = strings.Join(items, ",")
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
