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

func toBoolValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	var bVal bool
	switch rRawVal.Kind() {
	case reflect.Bool:
		bVal = rRawVal.Bool()
	default:
		err = fmt.Errorf("illegal remote bool value[%v]", rVal.Get())
	}
	if err != nil {
		return
	}

	ret, err = lType.Interface(bVal)
	return
}

func toIntValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	var iVal int64
	switch rRawVal.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		iVal = rRawVal.Int()
	case reflect.Float32, reflect.Float64:
		iVal = int64(rRawVal.Float())
	default:
		err = fmt.Errorf("illegal remote int value[%v]", rVal.Get())
	}
	if err != nil {
		return
	}

	ret, err = lType.Interface(iVal)
	return
}

func toUintValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	var uiVal uint64
	switch rRawVal.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		uiVal = rRawVal.Uint()
	case reflect.Float32, reflect.Float64:
		uiVal = uint64(rRawVal.Float())
	default:
		err = fmt.Errorf("illegal remote uint value[%v]", rVal.Get())
	}
	if err != nil {
		return
	}

	ret, err = lType.Interface(uiVal)
	return
}

func toFloatValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	var fVal float64
	switch rRawVal.Kind() {
	case reflect.Float32, reflect.Float64:
		fVal = rRawVal.Float()
	default:
		err = fmt.Errorf("illegal remote float value[%v]", rVal.Get())
	}
	if err != nil {
		return
	}

	ret, err = lType.Interface(fVal)
	return
}

func toStringValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	var strVal string
	switch rRawVal.Kind() {
	case reflect.String:
		strVal = rRawVal.String()
	default:
		err = fmt.Errorf("illegal remote string value[%v]", rVal.Get())
	}
	if err != nil {
		return
	}

	ret, err = lType.Interface(strVal)
	return
}

func toDateTimeValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	rRawVal := reflect.Indirect(reflect.ValueOf(rVal.Get()))
	var dtVal time.Time
	switch rRawVal.Kind() {
	case reflect.String:
		dtVal, err = time.Parse(util.CSTLayout, rRawVal.String())
	default:
		err = fmt.Errorf("illegal remote datetime value[%v]", rVal.Get())
	}
	if err != nil {
		return
	}

	lVal, _ := lType.Interface(nil)
	lRawVal := reflect.Indirect(lVal.Get().(reflect.Value))
	lRawVal.Set(reflect.ValueOf(dtVal))

	ret = lVal
	return
}

func toBasicValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	switch lType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = toBoolValue(rVal, lType)
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
		ret, err = toIntValue(rVal, lType)
	case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
		ret, err = toUintValue(rVal, lType)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = toFloatValue(rVal, lType)
	case model.TypeStringValue:
		ret, err = toStringValue(rVal, lType)
	case model.TypeDateTimeValue:
		ret, err = toDateTimeValue(rVal, lType)
	default:
		err = fmt.Errorf("illegal basic local type, type:%s", lType.GetPkgKey())
	}

	return
}

func toBasicSliceValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	if !model.IsSliceType(lType.GetValue()) {
		err = fmt.Errorf("illegal local slice type, type pkgKey:%v", lType.GetPkgKey())
		log.Error(err)
		return
	}

	lVal, _ := lType.Interface(nil)
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

	ret = lVal
	return
}

func toStructValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	objectValuePtr, objectValueOK := rVal.Get().(*remote.ObjectValue)
	if objectValueOK {
		ret, err = toLocalValue(objectValuePtr, lType)
		return
	}

	err = fmt.Errorf("illegal remote object value")
	log.Error(err)
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

	localType, localErr := local.GetEntityType(localEntity)
	if localErr != nil {
		err = localErr
		return
	}

	retVal, retErr := toLocalValue(remoteValue, localType)
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(reflect.Indirect(retVal.Get().(reflect.Value)))
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

func toLocalValue(rVal *remote.ObjectValue, lType model.Type) (ret model.Value, err error) {
	lVal, _ := lType.Interface(nil)
	lModel, lErr := local.GetEntityModel(lVal.Interface())
	if lErr != nil {
		err = lErr
		log.Error(err)
		return
	}

	if rVal.GetPkgKey() != lModel.GetPkgKey() {
		err = fmt.Errorf("mismatch pkgKey, remote object value pkgKey:%s, local model pkgKey:%s", rVal.GetPkgKey(), lModel.GetPkgKey())
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

	ret = local.NewValue(reflect.ValueOf(lModel.Interface(lType.IsPtrType())))
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

	sliceEntityValue, _ := lType.Interface(nil)
	for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
		sliceItem := sliceObjectValue.Values[idx]
		lVal, lErr := toLocalValue(sliceItem, lType.Elem())
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

	ret = sliceEntityValue
	return
}
