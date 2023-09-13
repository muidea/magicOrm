package remote

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote/codec"
	"github.com/muidea/magicOrm/provider/util"
)

var _codec codec.Codec

func init() {
	_codec = codec.New(ElemDependValue)
}

func GetCodec() codec.Codec {
	return _codec
}

func GetEntityType(entity interface{}) (ret model.Type, err error) {
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

	err = fmt.Errorf("illegal entity, entity:%v", entity)
	return
}

func GetEntityValue(entity interface{}) (ret model.Value, err error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = fmt.Errorf("%v", errInfo)
		}
	}()

	ret = NewValue(entity)
	return
}

func GetEntityModel(entity interface{}) (ret model.Model, err error) {
	objPtr, ok := entity.(*Object)
	if !ok {
		err = fmt.Errorf("illegal entity value, not object entity")
		return
	}

	err = objPtr.Verify()
	if err != nil {
		return
	}

	ret = objPtr
	return
}

func GetModelFilter(vModel model.Model) (ret model.Filter, err error) {
	objectImpl, objectOK := vModel.(*Object)
	if !objectOK {
		err = fmt.Errorf("invalid model value")
		return
	}

	ret = NewFilter(objectImpl)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = fmt.Errorf("SetModelValue failed, illegal value, err:%v", errInfo)
			return
		}
	}()

	if vVal.IsNil() {
		ret = vModel
		return
	}

	rVal := vVal.Interface().(*ObjectValue)
	if rVal.GetPkgKey() != vModel.GetPkgKey() {
		err = fmt.Errorf("illegal model value, mode PkgKey:%s, value PkgKey:%s", vModel.GetPkgKey(), rVal.GetPkgKey())
		return
	}
	for idx := 0; idx < len(rVal.Fields); idx++ {
		fieldVal := rVal.Fields[idx]
		if fieldVal.IsNil() {
			continue
		}

		err = vModel.SetFieldValue(fieldVal.GetName(), fieldVal.GetValue())
	}

	ret = vModel
	return
}

func elemSlice[T any](valSlice []T) (ret []model.Value) {
	for idx := 0; idx < len(valSlice); idx++ {
		ret = append(ret, NewValue(valSlice[idx]))
	}

	return
}

func ElemDependValue(vVal model.Value) (ret []model.Value, err error) {
	if vVal.IsNil() {
		err = fmt.Errorf("illegal value")
		return
	}

	sliceObjectValue, sliceOK := vVal.Get().(*SliceObjectValue)
	if sliceOK {
		for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
			ret = append(ret, NewValue(sliceObjectValue.Values[idx]))
		}
		return
	}

	objectValue, objectOK := vVal.Get().(*ObjectValue)
	if objectOK {
		ret = append(ret, NewValue(objectValue))
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

func AppendSliceValue(sliceVal model.Value, vVal model.Value) (ret model.Value, err error) {
	sliceValuePtr, sliceValueOK := sliceVal.Get().(*SliceObjectValue)
	if !sliceValueOK {
		err = fmt.Errorf("illegal slice value")
		return
	}

	valuePtr, valueOK := vVal.Get().(*ObjectValue)
	if !valueOK {
		err = fmt.Errorf("illegal item value")
		return

	}
	if sliceValuePtr.GetPkgKey() != valuePtr.GetPkgKey() {
		err = fmt.Errorf("mismatch slice value, slice pkgKey:%v, item pkgkey:%v", sliceValuePtr.GetPkgKey(), valuePtr.GetPkgKey())
		return
	}

	sliceValuePtr.Values = append(sliceValuePtr.Values, valuePtr)
	ret = NewValue(sliceValuePtr)
	return
}

func encodeModel(vVal model.Value, vType model.Type, mCache model.Cache, codec codec.Codec) (ret interface{}, err error) {
	tModel := mCache.Fetch(vType.GetPkgKey())
	if tModel == nil {
		err = fmt.Errorf("illegal value type,type:%s", vType.GetName())
		return
	}

	if vVal.IsBasic() {
		pkField := tModel.GetPrimaryField()
		pkType := pkField.GetType()
		ret, err = codec.Encode(vVal, pkType)
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
		tVal, _ = tType.Interface(nil)
	}

	ret, err = codec.Encode(tVal, tType)
	return
}

func encodeSliceModel(tVal model.Value, tType model.Type, mCache model.Cache, codec codec.Codec) (ret interface{}, err error) {
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
		strVal, strErr := encodeModel(v, tType.Elem(), mCache, codec)
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
		ret, err = _codec.Encode(tVal, tType)
		return
	}
	if model.IsStructType(tType.GetValue()) {
		ret, err = encodeModel(tVal, tType, mCache, _codec)
		return
	}

	ret, err = encodeSliceModel(tVal, tType, mCache, _codec)
	return
}

func DecodeValue(tVal interface{}, tType model.Type, mCache model.Cache) (ret model.Value, err error) {
	if tType.IsBasic() {
		ret, err = _codec.Decode(tVal, tType)
		return
	}

	objPtr, objOK := tVal.(*ObjectValue)
	if objOK {
		if objPtr.GetPkgKey() == tType.GetPkgKey() && model.IsStructType(tType.GetValue()) {
			ret, err = GetEntityValue(tVal)
			return
		}
	}

	sObjPtr, sObjOK := tVal.(*SliceObjectValue)
	if sObjOK {
		if sObjPtr.GetPkgKey() == tType.GetPkgKey() && model.IsSliceType(tType.GetValue()) {
			ret, err = GetEntityValue(tVal)
			return
		}
	}

	err = fmt.Errorf("unexpected type, type name:%s", tType.GetName())
	return
}

func GetValue(valueDeclare model.ValueDeclare) (ret model.Value) {
	var rVal interface{}
	switch valueDeclare {
	case model.SnowFlake:
		rVal = util.GetNewSnowFlakeID()
	case model.UUID:
		rVal = util.GetNewUUID()
	case model.DateTime:
		rVal = util.GetCurrentDateTime()
	}
	if rVal != nil {
		ret = NewValue(rVal)
		return
	}

	ret = &NilValue
	return
}
