package remote

import (
	"fmt"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
	"reflect"
	"time"
)

var _codec util.Codec

func init() {
	_codec = util.New(ElemDependValue)
}

func GetEntityType(entity interface{}) (ret model.Type, err *cd.Result) {
	if entity == nil {
		err = cd.NewError(cd.UnExpected, "entity is nil")
		return
	}
	// 为了提升执行性能，以下判断条件的顺序不可以随便调整
	objectInfo, objectInfoOK := entity.(Object)
	if objectInfoOK {
		impl := &TypeImpl{Name: objectInfo.GetName(), Value: model.TypeStructValue, PkgPath: objectInfo.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objectInfo.GetName(), Value: model.TypeStructValue, PkgPath: objectInfo.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	objectPtr, objectPtrOK := entity.(*Object)
	if objectPtrOK {
		impl := &TypeImpl{Name: objectPtr.GetName(), Value: model.TypeStructValue, PkgPath: objectPtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objectPtr.GetName(), Value: model.TypeStructValue, PkgPath: objectPtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	typeImplInfo, typeImplOK := entity.(TypeImpl)
	if typeImplOK {
		ret = &typeImplInfo
		return
	}

	typeImplPtr, typeImplPtrOK := entity.(*TypeImpl)
	if typeImplPtrOK {
		ret = typeImplPtr
		return
	}

	for {
		valueImplInfo, valueImplOK := entity.(ValueImpl)
		if valueImplOK {
			entity = valueImplInfo.Interface()
			break
		}

		valueImplPtr, valueImplPtrOK := entity.(*ValueImpl)
		if valueImplPtrOK {
			entity = valueImplPtr.Interface()
		}
		break
	}

	objectValue, objectValueOK := entity.(ObjectValue)
	if objectValueOK {
		impl := &TypeImpl{Name: objectValue.GetName(), Value: model.TypeStructValue, PkgPath: objectValue.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objectValue.GetName(), Value: model.TypeStructValue, PkgPath: objectValue.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	sObjectValue, sObjectValueOK := entity.(SliceObjectValue)
	if sObjectValueOK {
		impl := &TypeImpl{Name: sObjectValue.GetName(), Value: model.TypeSliceValue, PkgPath: sObjectValue.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: sObjectValue.GetName(), Value: model.TypeStructValue, PkgPath: sObjectValue.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	objectValuePtr, objectValuePtrOK := entity.(*ObjectValue)
	if objectValuePtrOK {
		impl := &TypeImpl{Name: objectValuePtr.GetName(), Value: model.TypeStructValue, PkgPath: objectValuePtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: objectValuePtr.GetName(), Value: model.TypeStructValue, PkgPath: objectValuePtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	sObjectValuePtr, sObjectValuePtrOK := entity.(*SliceObjectValue)
	if sObjectValuePtrOK {
		impl := &TypeImpl{Name: sObjectValuePtr.GetName(), Value: model.TypeSliceValue, PkgPath: sObjectValuePtr.GetPkgPath(), IsPtr: true}
		impl.ElemType = &TypeImpl{Name: sObjectValuePtr.GetName(), Value: model.TypeStructValue, PkgPath: sObjectValuePtr.GetPkgPath(), IsPtr: true}

		ret = impl
		return
	}

	switch entity.(type) {
	case Field, SpecImpl, FieldValue, *Field, *SpecImpl, *FieldValue:
		err = cd.NewError(cd.UnExpected, "illegal entity value")
		return
	}

	ret, err = getEntityType(entity)
	if err != nil {
		log.Errorf("GetEntityType failed, getEntityType error:%s", err.Error())
	}
	return
}

func GetEntityValue(entity interface{}) (ret model.Value, err *cd.Result) {
	if entity == nil {
		err = cd.NewError(cd.UnExpected, "entity is nil")
		return
	}

	// 为了提升执行性能，以下判断条件的顺序不可以随便调整
	valueImplInfo, valueImplOK := entity.(ValueImpl)
	if valueImplOK {
		ret = &valueImplInfo
		return
	}

	valueImplPtr, valueImplPtrOK := entity.(*ValueImpl)
	if valueImplPtrOK {
		ret = valueImplPtr
		return
	}

	objectValue, objectValueOK := entity.(ObjectValue)
	if objectValueOK {
		ret = NewValue(&objectValue)
		return
	}

	sObjectValue, sObjectValueOK := entity.(SliceObjectValue)
	if sObjectValueOK {
		ret = NewValue(&sObjectValue)
		return
	}

	objectValuePtr, objectValuePtrOK := entity.(*ObjectValue)
	if objectValuePtrOK {
		ret = NewValue(objectValuePtr)
		return
	}

	sObjectValuePtr, sObjectValuePtrOK := entity.(*SliceObjectValue)
	if sObjectValuePtrOK {
		ret = NewValue(sObjectValuePtr)
		return
	}

	switch entity.(type) {
	case Object, Field, TypeImpl, SpecImpl, FieldValue, *Object, *Field, *TypeImpl, *SpecImpl, *FieldValue:
		err = cd.NewError(cd.UnExpected, "illegal entity value")
		return
	}

	entityType, entityErr := getEntityType(entity)
	if entityErr != nil {
		err = entityErr
		return
	}

	if entityType.IsBasic() {
		ret = NewValue(entity)
		return
	}
	if entityType.IsSlice() {
		entity, err = GetSliceObjectValue(entity)
	} else {
		entity, err = GetObjectValue(entity)
	}
	if err != nil {
		return
	}

	ret = NewValue(entity)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err *cd.Result) {
	var objectPtr *Object
	ptrVal, ptrOK := entity.(*Object)
	if ptrOK {
		objectPtr = ptrVal
	}

	infoVal, infoOK := entity.(Object)
	if infoOK {
		objectPtr = &infoVal
	}
	if objectPtr == nil {
		switch entity.(type) {
		case *ObjectValue, *SliceObjectValue, *Field, *TypeImpl, *ValueImpl, *SpecImpl, *FieldValue,
			ObjectValue, SliceObjectValue, Field, TypeImpl, ValueImpl, SpecImpl, FieldValue:
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal entity value, not object entity"))
		}
		if err != nil {
			return
		}

		objectPtr, err = GetObject(entity)
		if err != nil {
			return
		}
	}

	err = objectPtr.Verify()
	if err != nil {
		return
	}

	ret = objectPtr
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

	if !vVal.IsValid() {
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

func ElemDependValue(eVal interface{}) (ret []model.Value, err *cd.Result) {
	if eVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("eVal is nil"))
		log.Errorf("ElemDependValue failed, er:%s", err.Error())
		return
	}

	sliceObjectPtrValue, slicePtrOK := eVal.(*SliceObjectValue)
	if slicePtrOK {
		for idx := 0; idx < len(sliceObjectPtrValue.Values); idx++ {
			ret = append(ret, NewValue(sliceObjectPtrValue.Values[idx]))
		}
		return
	}

	listObjectValuePtr, listOK := eVal.([]*ObjectValue)
	if listOK {
		for idx := 0; idx < len(listObjectValuePtr); idx++ {
			ret = append(ret, NewValue(listObjectValuePtr[idx]))
		}
		return
	}

	rVal := reflect.Indirect(reflect.ValueOf(eVal))
	if rVal.Kind() != reflect.Slice {
		ret = append(ret, NewValue(eVal))
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		ret = append(ret, NewValue(rVal.Index(idx).Interface()))
	}
	return
}

func appendObjectSlice(sliceObjectValuePtr *SliceObjectValue, objectValuePtr *ObjectValue) (ret *SliceObjectValue, err *cd.Result) {
	if sliceObjectValuePtr.GetPkgKey() != objectValuePtr.GetPkgKey() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("mismatch slice value, slice pkgKey:%v, item pkgkey:%v", sliceObjectValuePtr.GetPkgKey(), objectValuePtr.GetPkgKey()))
		log.Errorf("appendObjectSlice failed, err:%s", err.Error())
		return
	}

	sliceObjectValuePtr.Values = append(sliceObjectValuePtr.Values, objectValuePtr)
	ret = sliceObjectValuePtr
	return
}

func appendBaseSlice(sliceValue interface{}, value interface{}) (ret interface{}, err *cd.Result) {
	sliceType, sliceErr := getEntityType(sliceValue)
	if sliceErr != nil {
		err = sliceErr
		return
	}

	valueType, valueErr := getEntityType(value)
	if valueErr != nil {
		err = valueErr
		return
	}
	if !sliceType.IsSlice() || !sliceType.IsBasic() || sliceType.Elem().GetPkgKey() != valueType.GetPkgKey() {
		err = cd.NewError(cd.UnExpected, "illegal slice value")
		return
	}

	if !sliceType.IsPtrType() {
		switch valueType.GetValue() {
		case model.TypeBooleanValue:
			ret = append(sliceValue.([]bool), value.(bool))
		case model.TypeBitValue:
			ret = append(sliceValue.([]int8), value.(int8))
		case model.TypeSmallIntegerValue:
			ret = append(sliceValue.([]int16), value.(int16))
		case model.TypeInteger32Value:
			ret = append(sliceValue.([]int32), value.(int32))
		case model.TypeBigIntegerValue:
			ret = append(sliceValue.([]int64), value.(int64))
		case model.TypeIntegerValue:
			ret = append(sliceValue.([]int), value.(int))
		case model.TypePositiveBitValue:
			ret = append(sliceValue.([]uint8), value.(uint8))
		case model.TypePositiveSmallIntegerValue:
			ret = append(sliceValue.([]uint16), value.(uint16))
		case model.TypePositiveInteger32Value:
			ret = append(sliceValue.([]uint32), value.(uint32))
		case model.TypePositiveBigIntegerValue:
			ret = append(sliceValue.([]uint64), value.(uint64))
		case model.TypePositiveIntegerValue:
			ret = append(sliceValue.([]uint), value.(uint))
		case model.TypeStringValue:
			ret = append(sliceValue.([]string), value.(string))
		case model.TypeDateTimeValue:
			ret = append(sliceValue.([]time.Time), value.(time.Time))
		default:
			err = cd.NewError(cd.UnExpected, "illegal value type")
		}

		return
	}

	switch valueType.GetValue() {
	case model.TypeBooleanValue:
		sliceVal := append(*sliceValue.(*[]bool), value.(bool))
		ret = &sliceVal
	case model.TypeBitValue:
		sliceVal := append(*sliceValue.(*[]int8), value.(int8))
		ret = &sliceVal
	case model.TypeSmallIntegerValue:
		sliceVal := append(*sliceValue.(*[]int16), value.(int16))
		ret = &sliceVal
	case model.TypeInteger32Value:
		sliceVal := append(*sliceValue.(*[]int32), value.(int32))
		ret = &sliceVal
	case model.TypeBigIntegerValue:
		sliceVal := append(*sliceValue.(*[]int64), value.(int64))
		ret = &sliceVal
	case model.TypeIntegerValue:
		sliceVal := append(*sliceValue.(*[]int), value.(int))
		ret = &sliceVal
	case model.TypePositiveBitValue:
		sliceVal := append(*sliceValue.(*[]uint8), value.(uint8))
		ret = &sliceVal
	case model.TypePositiveSmallIntegerValue:
		sliceVal := append(*sliceValue.(*[]uint16), value.(uint16))
		ret = &sliceVal
	case model.TypePositiveInteger32Value:
		sliceVal := append(*sliceValue.(*[]uint32), value.(uint32))
		ret = &sliceVal
	case model.TypePositiveBigIntegerValue:
		sliceVal := append(*sliceValue.(*[]uint64), value.(uint64))
		ret = &sliceVal
	case model.TypePositiveIntegerValue:
		sliceVal := append(*sliceValue.(*[]uint), value.(uint))
		ret = &sliceVal
	case model.TypeStringValue:
		sliceVal := append(*sliceValue.(*[]string), value.(string))
		ret = &sliceVal
	case model.TypeDateTimeValue:
		sliceVal := append(*sliceValue.(*[]time.Time), value.(time.Time))
		ret = &sliceVal
	default:
		err = cd.NewError(cd.UnExpected, "illegal value type")
	}
	if err != nil {
		return
	}
	return
}

func AppendSliceValue(sliceVal model.Value, vVal model.Value) (ret model.Value, err *cd.Result) {
	sliceObjectValuePtr, sliceObjectValueOK := sliceVal.Get().(*SliceObjectValue)
	objectValuePtr, objectValueOK := vVal.Get().(*ObjectValue)
	if sliceObjectValueOK && objectValueOK {
		sliceObjectValuePtr, err = appendObjectSlice(sliceObjectValuePtr, objectValuePtr)
		if err != nil {
			log.Errorf("AppendSliceValue failed, err:%s", err.Error())
			return
		}
		ret = NewValue(sliceObjectValuePtr)
		return
	}
	baseSliceVal, baseSliceErr := appendBaseSlice(sliceVal.Get(), vVal.Get())
	if baseSliceErr != nil {
		err = baseSliceErr
		return
	}

	ret = NewValue(baseSliceVal)
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
	if !tVal.IsValid() {
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
