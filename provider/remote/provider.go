package remote

import (
	"fmt"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote/codec"
	"github.com/muidea/magicOrm/provider/util"
)

var _codec codec.Codec

func init() {
	_codec = codec.New(ElemDependValue)
}

func GetEntityType(entity interface{}) (ret model.Type, err *cd.Result) {
	objInfo, objOK := entity.(Object)
	if objOK {
		impl := &TypeImpl{Name: objInfo.GetName(), Value: model.TypeStructValue, PkgPath: objInfo.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objInfo.GetName(), Value: model.TypeStructValue, PkgPath: objInfo.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	valInfo, valOK := entity.(ObjectValue)
	if valOK {
		impl := &TypeImpl{Name: valInfo.GetName(), Value: model.TypeStructValue, PkgPath: valInfo.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: valInfo.GetName(), Value: model.TypeStructValue, PkgPath: valInfo.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	sValInfo, sValOK := entity.(SliceObjectValue)
	if sValOK {
		impl := &TypeImpl{Name: sValInfo.GetName(), Value: model.TypeSliceValue, PkgPath: sValInfo.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: sValInfo.GetName(), Value: model.TypeStructValue, PkgPath: sValInfo.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	typeInfo, typeOK := entity.(TypeImpl)
	if typeOK {
		ret = &typeInfo
		return
	}

	objPtr, objOK := entity.(*Object)
	if objOK {
		impl := &TypeImpl{Name: objPtr.GetName(), Value: model.TypeStructValue, PkgPath: objPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objPtr.GetName(), Value: model.TypeStructValue, PkgPath: objPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	valPtr, valOK := entity.(*ObjectValue)
	if valOK {
		impl := &TypeImpl{Name: valPtr.GetName(), Value: model.TypeStructValue, PkgPath: valPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: valPtr.GetName(), Value: model.TypeStructValue, PkgPath: valPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	sValPtr, sValOK := entity.(*SliceObjectValue)
	if sValOK {
		impl := &TypeImpl{Name: sValPtr.GetName(), Value: model.TypeSliceValue, PkgPath: sValPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: sValPtr.GetName(), Value: model.TypeStructValue, PkgPath: sValPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	typePtr, typeOK := entity.(*TypeImpl)
	if typeOK {
		ret = typePtr
		return
	}

	err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal entity, entity:%v", entity))
	return
}

func GetEntityValue(entity interface{}) (ret model.Value, err *cd.Result) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("%v", errInfo))
		}
	}()

	ret = NewValue(entity)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err *cd.Result) {
	var objPtr *Object
	ptrVal, ptrOK := entity.(*Object)
	if ptrOK {
		objPtr = ptrVal
	}

	infoVal, infoOK := entity.(Object)
	if infoOK {
		objPtr = &infoVal
	}
	if objPtr == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal entity value, not object entity"))
		return
	}

	err = objPtr.Verify()
	if err != nil {
		return
	}

	ret = objPtr
	return
}

func GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Result) {
	objectImpl, objectOK := vModel.(*Object)
	if !objectOK {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("invalid model value"))
		return
	}

	ret = NewFilter(objectImpl)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err *cd.Result) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("SetModelValue failed, illegal value, err:%v", errInfo))
			log.Errorf("SetModelValue failed, err:%s", err.Error())
			return
		}
	}()

	if vVal.IsNil() {
		ret = vModel
		return
	}
	if vVal.IsBasic() {
		vModel.SetPrimaryFieldValue(vVal)

		ret = vModel
		return
	}

	val := vVal.Interface()
	rValPtr, rValOK := val.(*ObjectValue)
	if !rValOK {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model value, val:%v", val))
		log.Errorf("SetModelValue failed, err:%s", err.Error())
		return
	}

	if rValPtr.GetPkgKey() != vModel.GetPkgKey() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model value, mode PkgKey:%s, value PkgKey:%s", vModel.GetPkgKey(), rValPtr.GetPkgKey()))
		log.Errorf("SetModelValue failed, err:%s", err.Error())
		return
	}
	for idx := 0; idx < len(rValPtr.Fields); idx++ {
		fieldVal := rValPtr.Fields[idx]
		vModel.SetFieldValue(fieldVal.GetName(), fieldVal.GetValue())
	}

	ret = vModel
	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err *cd.Result) {
	if vVal.IsNil() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("vVal is nil"))
		log.Errorf("ElemDependValue failed, er:%s", err.Error())
		return
	}

	sliceObjectPtrValue, slicePtrOK := vVal.Get().(*SliceObjectValue)
	if slicePtrOK {
		for idx := 0; idx < len(sliceObjectPtrValue.Values); idx++ {
			ret = append(ret, NewValue(sliceObjectPtrValue.Values[idx]))
		}
		return
	}

	listObjectValuePtr, listOK := vVal.Get().([]*ObjectValue)
	if listOK {
		for idx := 0; idx < len(listObjectValuePtr); idx++ {
			ret = append(ret, NewValue(listObjectValuePtr[idx]))
		}
		return
	}

	objectPtrValue, objectPtrOK := vVal.Get().(*ObjectValue)
	if objectPtrOK {
		ret = append(ret, NewValue(objectPtrValue))
		return
	}

	rVal := reflect.ValueOf(vVal.Get())
	if rVal.Kind() != reflect.Slice {
		ret = append(ret, NewValue(vVal.Get()))
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		ret = append(ret, NewValue(rVal.Index(idx).Interface()))
	}
	return
}

func AppendSliceValue(sliceVal model.Value, vVal model.Value) (ret model.Value, err *cd.Result) {
	sliceObjectValuePtr, sliceObjectValueOK := sliceVal.Get().(*SliceObjectValue)
	if sliceObjectValuePtr == nil || !sliceObjectValueOK {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal slice item value"))
		log.Errorf("AppendSliceValue failed, err:%s", err.Error())
		return
	}

	objectValuePtr, objectValueOK := vVal.Get().(*ObjectValue)
	if objectValuePtr == nil || !objectValueOK {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal item value"))
		log.Errorf("AppendSliceValue failed, err:%s", err.Error())
		return
	}

	if sliceObjectValuePtr.GetPkgKey() != objectValuePtr.GetPkgKey() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("mismatch slice value, slice pkgKey:%v, item pkgkey:%v", sliceObjectValuePtr.GetPkgKey(), objectValuePtr.GetPkgKey()))
		log.Errorf("AppendSliceValue failed, err:%s", err.Error())
		return
	}

	sliceObjectValuePtr.Values = append(sliceObjectValuePtr.Values, objectValuePtr)
	ret = NewValue(sliceObjectValuePtr)
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache) (ret interface{}, err *cd.Result) {
	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model type,type:%s", vType.GetName()))
		log.Errorf("encodeModel failed, err:%s", err.Error())
		return
	}

	if vVal.IsBasic() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model value"))
		log.Errorf("encodeModel failed, err:%s", err.Error())
		return
	}

	vModel, vErr := SetModelValue(tModel.Copy(true), vVal)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	tType := pkField.GetType()
	tVal := pkField.GetValue()
	if tVal.IsNil() {
		tVal, _ = tType.Interface(nil)
	}

	ret, err = _codec.Encode(tVal, tType)
	return
}

func encodeSliceModel(tVal model.Value, tType model.Type, mCache model.Cache) (ret interface{}, err *cd.Result) {
	vVals, vErr := ElemDependValue(tVal)
	if vErr != nil {
		err = vErr
		return
	}
	if len(vVals) == 0 {
		return
	}

	items := []interface{}{}
	for _, v := range vVals {
		mVal, mErr := encodeModel(v, tType.Elem(), mCache)
		if mErr != nil {
			err = mErr
			return
		}

		items = append(items, mVal)
	}

	ret = items
	return
}

func EncodeValue(tVal model.Value, tType model.Type, mCache model.Cache) (ret interface{}, err *cd.Result) {
	if tType.IsBasic() {
		ret, err = _codec.Encode(tVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = encodeModel(tVal, tType, mCache)
		return
	}

	ret, err = encodeSliceModel(tVal, tType, mCache)
	return
}

func decodeModel(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	tModel := mCache.Fetch(tType.GetPkgKey())
	if tModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal value type,type:%s", tType.GetName()))
		return
	}

	mVal, mErr := GetEntityValue(tVal)
	if mErr != nil {
		err = mErr
		return
	}

	var vErr *cd.Result
	vModel := tModel.Copy(true)
	if mVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		pkVal, pkErr := _codec.Decode(tVal, pkField.GetType())
		if pkErr != nil {
			err = pkErr
			return
		}

		vModel.SetPrimaryFieldValue(pkVal)
	} else {
		vModel, vErr = SetModelValue(vModel, mVal)
	}

	if vErr != nil {
		err = vErr
		return
	}

	tVal = vModel.Interface(tType.IsPtrType(), model.OriginView)
	ret, err = GetEntityValue(tVal)
	return
}

func decodeSliceModel(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	tModel := mCache.Fetch(tType.GetPkgKey())
	if tModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal value type,type:%s", tType.GetName()))
		return
	}
	mVal, mErr := GetEntityValue(tVal)
	if mErr != nil {
		err = mErr
		return
	}

	var mVals []model.Value
	mVals, mErr = ElemDependValue(mVal)
	if mErr != nil {
		err = mErr
		return
	}

	var vErr *cd.Result
	vVals, _ := tType.Interface(nil)
	for _, val := range mVals {
		vModel := tModel.Copy(true)
		if val.IsBasic() {
			pkField := tModel.GetPrimaryField()
			pkVal, pkErr := _codec.Decode(val.Interface(), pkField.GetType())
			if pkErr != nil {
				err = pkErr
				return
			}
			vModel.SetPrimaryFieldValue(pkVal)
		} else {
			vModel, vErr = SetModelValue(vModel, val)
			if vErr != nil {
				err = vErr
				return
			}
		}

		iVal, _ := GetEntityValue(vModel.Interface(tType.Elem().IsPtrType(), model.OriginView))
		vVals, vErr = AppendSliceValue(vVals, iVal)
		if vErr != nil {
			err = vErr
			return
		}
	}

	ret = vVals
	return
}

func DecodeValue(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err *cd.Result) {
	if tType.IsBasic() {
		ret, err = _codec.Decode(tVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = decodeModel(tVal, tType, mCache)
		return
	}

	ret, err = decodeSliceModel(tVal, tType, mCache)
	return
}

func GetNewValue(valueDeclare model.ValueDeclare) (ret model.Value) {
	var rVal interface{}
	switch valueDeclare {
	case model.SnowFlake:
		rVal = util.GetNewSnowFlakeID()
	case model.UUID:
		rVal = util.GetNewUUID()
	case model.DateTime:
		rVal = util.GetCurrentDateTimeStr()
	}
	if rVal != nil {
		ret = NewValue(rVal)
		return
	}

	ret = &NilValue
	return
}
