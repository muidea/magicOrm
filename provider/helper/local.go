package helper

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

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
		lSubVal, lSubErr := lType.Elem().Interface(rValList[idx].Get())
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
	var objectValuePtr *remote.ObjectValue
	valuePtr, valuePtrOK := rVal.Get().(*remote.ObjectValue)
	if valuePtrOK {
		objectValuePtr = valuePtr
	}
	value, valueOK := rVal.Get().(remote.ObjectValue)
	if valueOK {
		objectValuePtr = &value
	}
	if objectValuePtr == nil {
		err = fmt.Errorf("illegal remote object value")
		log.Error(err)
		return
	}

	ret, err = toLocalValue(objectValuePtr, lType)
	return
}

func toStructSliceValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	var sliceObjectValuePtr *remote.SliceObjectValue
	valuePtr, valuePtrOK := rVal.Get().(*remote.SliceObjectValue)
	if valuePtrOK {
		sliceObjectValuePtr = valuePtr
	}
	value, valueOK := rVal.Get().(remote.SliceObjectValue)
	if valueOK {
		sliceObjectValuePtr = &value
	}
	if sliceObjectValuePtr == nil {
		err = fmt.Errorf("illegal remote slice value")
		return
	}

	ret, err = toLocalSliceValue(sliceObjectValuePtr, lType)
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
		lVal, lErr := lType.Interface(fieldVal.GetValue().Get())
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
