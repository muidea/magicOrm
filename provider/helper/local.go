package helper

import (
	"fmt"
	"reflect"
	"time"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

func toBasicValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	lVal := lType.Interface()
	lRawVal := reflect.Indirect(lVal.Get().(reflect.Value))
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	switch lType.GetValue() {
	case model.TypeBooleanValue:
		lRawVal.SetBool(rRawVal.Bool())
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
		switch rVal.Get().(type) {
		case int8, int16, int32, int, int64:
			lRawVal.SetInt(rRawVal.Int())
		case float64:
			lRawVal.SetInt(int64(rRawVal.Float()))
		default:
			err = fmt.Errorf("illegal int, value:%v", rVal.Get())
		}
	case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
		switch rVal.Get().(type) {
		case uint8, uint16, uint32, uint, uint64:
			lRawVal.SetUint(rRawVal.Uint())
		case float64:
			lRawVal.SetUint(uint64(rRawVal.Float()))
		default:
			err = fmt.Errorf("illegal uint, value:%v", rVal.Get())
		}
	case model.TypeFloatValue, model.TypeDoubleValue:
		lRawVal.SetFloat(rRawVal.Float())
	case model.TypeStringValue:
		lRawVal.SetString(rRawVal.String())
	case model.TypeDateTimeValue:
		switch rVal.Get().(type) {
		case string:
			dtVal, _ := time.Parse(util.CSTLayout, rRawVal.String())
			lRawVal.Set(reflect.ValueOf(dtVal))
		default:
			err = fmt.Errorf("illegal dateTime, value:%v", rVal.Get())
		}
	default:
		err = fmt.Errorf("illegal basic local type, type:%s", lType.GetPkgKey())
	}

	if err != nil {
		return
	}

	if lType.IsPtrType() {
		ret = lVal.Addr()
		return
	}

	ret = lVal
	return
}

func toBasicSliceValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	if !model.IsSliceType(lType.GetValue()) {
		err = fmt.Errorf("illegal local type value, type pkgKey:%v", lType.GetPkgKey())
		log.Error(err)
		return
	}

	lVal := lType.Interface()
	rValList, rErr := remote.ElemDependValue(rVal)
	if rErr != nil {
		err = rErr
		return
	}

	for idx := 0; idx < len(rValList); idx++ {
		lSubVal, lSubErr := toBasicValue(rValList[idx], lType.Elem())
		if lSubErr != nil {
			err = lSubErr
			return
		}

		lVal, err = local.AppendSliceValue(lVal, lSubVal)
		if err != nil {
			return
		}
	}

	if lType.IsPtrType() {
		ret = lVal.Addr()
		return
	}

	ret = lVal
	return
}

func toStructValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	lModel, lErr := local.GetEntityModel(lType.Interface().Addr().Interface())
	if lErr != nil {
		err = lErr
		log.Error(err)
		return
	}

	objectValuePtr, objectValueOK := rVal.Get().(*remote.ObjectValue)
	if objectValueOK {
		ret, err = toLocalValue(objectValuePtr, lModel, lType.IsPtrType())
		return
	}

	err = fmt.Errorf("illegal remote value")
	return
}

func toStructSliceValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	sliceObjectValuePtr, sliceObjectValueOK := rVal.Get().(*remote.SliceObjectValue)
	if sliceObjectValueOK {
		ret, err = toLocalSliceValue(sliceObjectValuePtr, lType)
		return
	}

	err = fmt.Errorf("illegal remote slice value")
	return
}

// UpdateEntity update object value -> entity
func UpdateEntity(remoteValue *remote.ObjectValue, localEntity any) (err error) {
	if !remoteValue.IsAssigned() {
		return
	}

	entityValue := reflect.ValueOf(localEntity)
	if entityValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal localEntity value, must be a pointer localEntity")
		return
	}

	entityValue = reflect.Indirect(entityValue)
	if entityValue.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal localEntity, must be a struct localEntity")
		return
	}
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal localEntity value, can't be set")
		return
	}

	localModel, localErr := local.GetEntityModel(localEntity)
	if localErr != nil {
		err = localErr
		return
	}

	retVal, retErr := toLocalValue(remoteValue, localModel, true)
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal.Get().(reflect.Value))
	return
}

func toLocalFieldValue(fieldVal *remote.FieldValue, lField model.Field) (err error) {
	lType := lField.GetType()
	if lType.IsBasic() && model.IsSliceType(lType.GetValue()) {
		lVal, lErr := toBasicSliceValue(fieldVal.GetValue(), lType)
		if lErr != nil {
			err = lErr
			return
		}
		err = lField.SetValue(lVal)
		return
	}

	if lType.IsBasic() {
		lVal, lErr := toBasicValue(fieldVal.GetValue(), lType)
		if lErr != nil {
			err = lErr
			return
		}
		err = lField.SetValue(lVal)
		return
	}

	if model.IsSliceType(lType.GetValue()) {
		lVal, lErr := toStructSliceValue(fieldVal.GetValue(), lType)
		if lErr != nil {
			err = lErr
			return
		}
		err = lField.SetValue(lVal)
		return
	}

	lVal, lErr := toStructValue(fieldVal.GetValue(), lType)
	if lErr != nil {
		err = lErr
		return
	}

	err = lField.SetValue(lVal)
	return
}

func toLocalValue(rVal *remote.ObjectValue, lModel model.Model, ptrValue bool) (ret model.Value, err error) {
	if rVal.GetPkgKey() != lModel.GetPkgKey() {
		err = fmt.Errorf("mismatch pkgKey, remote value pkgKey:%s, local model pkgKey:%s", rVal.GetPkgKey(), lModel.GetPkgKey())
		return
	}

	for idx := 0; idx < len(rVal.Fields); idx++ {
		fieldVal := rVal.Fields[idx]
		if fieldVal.IsNil() {
			continue
		}

		lField := lModel.GetField(fieldVal.GetName())
		if lField == nil {
			continue
		}

		err = toLocalFieldValue(fieldVal, lField)
		if err != nil {
			return
		}
	}

	ret = local.NewValue(reflect.ValueOf(lModel.Interface(ptrValue)))
	return
}

// UpdateSliceEntity update slice object value -> entitySlice
func UpdateSliceEntity(remoteValue *remote.SliceObjectValue, localEntity any) (err error) {
	if !remoteValue.IsAssigned() {
		return
	}

	entityValue := reflect.ValueOf(localEntity)
	if entityValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal localEntity value, must be a pointer localEntity")
		return
	}

	vType, vErr := local.GetType(entityValue.Type())
	if vErr != nil {
		err = vErr
		return
	}

	entityValue = reflect.Indirect(entityValue)
	if entityValue.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal localEntity, must be a struct localEntity")
		return
	}
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal localEntity value, can't be set")
		return
	}

	retVal, retErr := toLocalSliceValue(remoteValue, vType)
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal.Get().(reflect.Value))
	return
}

func toLocalSliceValue(sliceObjectValue *remote.SliceObjectValue, lType model.Type) (ret model.Value, err error) {
	if sliceObjectValue.GetPkgKey() != lType.GetPkgKey() {
		err = fmt.Errorf("illegal slice object value, sliceObjectValue pkgKey:%s, sliceEntityType pkgKey:%s", sliceObjectValue.GetPkgKey(), lType.GetPkgKey())
		log.Errorf("toLocalSliceValue failed, mismatch objectValue for value, err:%s", err)
		return
	}

	sliceEntityValue := lType.Interface()
	elemVal := lType.Elem().Interface().Addr().Interface()
	lModel, lErr := local.GetEntityModel(elemVal)
	if lErr != nil {
		err = lErr
		return
	}

	for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
		sliceItem := sliceObjectValue.Values[idx]
		lVal, lErr := toLocalValue(sliceItem, lModel, lType.Elem().IsPtrType())
		if lErr != nil {
			err = fmt.Errorf("toLocalValue error [%v]", lErr.Error())
			log.Errorf("toLocalSliceValue failed, err:%s", sliceItem.GetName(), err.Error())
			return
		}

		sliceEntityValue, err = local.AppendSliceValue(sliceEntityValue, lVal)
		if err != nil {
			return
		}
	}

	if lType.IsPtrType() {
		ret = sliceEntityValue.Addr()
		return
	}

	ret = sliceEntityValue
	return
}
