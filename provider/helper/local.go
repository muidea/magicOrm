package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

func toBasicValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	lVal, lErr := lType.Interface(rVal.Get())
	if lErr != nil {
		err = lErr
		log.Errorf("toBasicValue failed, lType.Interface err:%v", err.Error())
		return
	}

	ret = lVal
	return
}

func toBasicSliceValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	if !model.IsSliceType(lType.GetValue()) {
		err = fmt.Errorf("illegal slice value, type pkgKey:%v", lType.GetPkgKey())
		log.Errorf("toBasicSliceValue failed, err:%v", err.Error())
		return
	}

	rValList, rErr := remote.ElemDependValue(rVal)
	if rErr != nil {
		err = rErr
		log.Errorf("toBasicSliceValue failed, remote.ElemDependValue err:%v", err.Error())
		return
	}

	lVal, _ := lType.Interface(nil)
	lSubType := lType.Elem()
	for idx := 0; idx < len(rValList); idx++ {
		lSubVal, lSubErr := toBasicValue(rValList[idx], lSubType)
		if lSubErr != nil {
			err = lSubErr
			log.Errorf("toBasicSliceValue failed, toBasicValue err:%v", err.Error())
			return
		}

		lVal, err = local.AppendSliceValue(lVal, lSubVal)
		if err != nil {
			log.Errorf("toBasicSliceValue failed, local.AppendSliceValue err:%v", err.Error())
			return
		}
	}

	ret = lVal
	return
}

func toLocalFieldValue(fieldVal *remote.FieldValue, lField model.Field) (err error) {
	lType := lField.GetType()
	// basic slice
	if model.IsBasicSlice(lType) {
		lVal, lErr := toBasicSliceValue(fieldVal.GetValue(), lType)
		if lErr != nil {
			err = lErr
			log.Errorf("toLocalFieldValue failed, toBasicSliceValue err:%v", err.Error())
			return
		}
		err = lField.SetValue(lVal)
		if err != nil {
			log.Errorf("toLocalFieldValue failed, basic slice lField.SetValue err:%v", err.Error())
		}
		return
	}

	// basic
	if lType.IsBasic() {
		//lVal, lErr := lType.Interface(fieldVal.GetValue().Get())
		lVal, lErr := toBasicValue(fieldVal.GetValue(), lType)
		if lErr != nil {
			err = lErr
			log.Errorf("toLocalFieldValue failed,toBasicValue err:%v", err.Error())
			return
		}
		err = lField.SetValue(lVal)
		if err != nil {
			log.Errorf("toLocalFieldValue failed, basic lField.SetValue err:%v", err.Error())
		}
		return
	}

	// struct slice
	if model.IsStructSlice(lType) {
		lVal, lErr := toStructSliceValue(fieldVal.GetValue(), lType)
		if lErr != nil {
			err = lErr
			log.Errorf("toLocalFieldValue failed, toStructSliceValue err:%v", err.Error())
			return
		}
		err = lField.SetValue(lVal)
		if err != nil {
			log.Errorf("toLocalFieldValue failed, struct slice lField.SetValue err:%v", err.Error())
		}
		return
	}

	// struct
	lVal, lErr := toStructValue(fieldVal.GetValue(), lType)
	if lErr != nil {
		err = lErr
		log.Errorf("toLocalFieldValue failed, toStructValue err:%v", err.Error())
		return
	}

	err = lField.SetValue(lVal)
	if err != nil {
		log.Errorf("toLocalFieldValue failed, struct lField.SetValue err:%v", err.Error())
	}
	return
}

func toLocalValue(rVal *remote.ObjectValue, lType model.Type) (ret model.Value, err error) {
	lVal, _ := lType.Interface(nil)
	lModel, lErr := local.GetEntityModel(lVal.Interface())
	if lErr != nil {
		err = lErr
		log.Errorf("toLocalValue failed, local.GetEntityModel err:%v", err.Error())
		return
	}

	if rVal.GetPkgKey() != lModel.GetPkgKey() {
		err = fmt.Errorf("mismatch pkgKey, remote object value pkgKey:%s, local model pkgKey:%s", rVal.GetPkgKey(), lModel.GetPkgKey())
		log.Errorf("toLocalValue failed, err:%v", err.Error())
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
			log.Errorf("toLocalValue failed, toLocalFieldValue err:%v", err.Error())
			return
		}
	}

	ret = local.NewValue(reflect.ValueOf(lModel.Interface(lType.IsPtrType())))
	return
}

func toLocalSliceValue(sliceObjectValue *remote.SliceObjectValue, lType model.Type) (ret model.Value, err error) {
	if sliceObjectValue.GetPkgKey() != lType.GetPkgKey() {
		err = fmt.Errorf("mismatch objectValue for value, sliceObjectValue pkgKey:%s, sliceEntityType pkgKey:%s", sliceObjectValue.GetPkgKey(), lType.GetPkgKey())
		log.Errorf("toLocalSliceValue failed, err:%s", err)
		return
	}

	sliceEntityValue, _ := lType.Interface(nil)
	for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
		sliceItem := sliceObjectValue.Values[idx]
		lVal, lErr := toLocalValue(sliceItem, lType.Elem())
		if lErr != nil {
			err = fmt.Errorf("sliceItem name:%s, error [%v]", sliceItem.GetName(), lErr.Error())
			log.Errorf("toLocalSliceValue failed, toLocalValue err:%s", err.Error())
			return
		}

		sliceEntityValue, err = local.AppendSliceValue(sliceEntityValue, lVal)
		if err != nil {
			log.Errorf("toLocalSliceValue failed, local.AppendSliceValue err:%s", err.Error())
			return
		}
	}

	ret = sliceEntityValue
	return
}

func toStructValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	objectValuePtr, objectValueOK := rVal.Get().(*remote.ObjectValue)
	if objectValuePtr == nil || !objectValueOK {
		err = fmt.Errorf("illegal remote object value")
		log.Errorf("toStructValue failed, error:%s", err.Error())
		return
	}

	ret, err = toLocalValue(objectValuePtr, lType)
	if err != nil {
		log.Errorf("toStructValue failed, toLocalValue error:%s", err.Error())
	}
	return
}

func toStructSliceValue(rVal model.Value, lType model.Type) (ret model.Value, err error) {
	sliceObjectValuePtr, sliceObjectValueOK := rVal.Get().(*remote.SliceObjectValue)
	if sliceObjectValuePtr == nil || !sliceObjectValueOK {
		err = fmt.Errorf("illegal remote slice value")
		log.Errorf("toStructSliceValue failed, err:%s", err.Error())
		return
	}

	ret, err = toLocalSliceValue(sliceObjectValuePtr, lType)
	if err != nil {
		log.Errorf("toStructSliceValue failed, toLocalSliceValue error:%s", err.Error())
	}
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
		log.Errorf("UpdateEntity failed, error:%s", err.Error())
		return
	}

	entityValue = reflect.Indirect(entityValue)
	if entityValue.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal localEntity, must be a struct localEntity")
		return
	}
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal localEntity value, can't be set")
		log.Errorf("UpdateEntity failed, error:%s", err.Error())
		return
	}

	localType, localErr := local.GetEntityType(localEntity)
	if localErr != nil {
		err = localErr
		log.Errorf("UpdateEntity failed, local.GetEntityType error:%s", err.Error())
		return
	}

	retVal, retErr := toLocalValue(remoteValue, localType)
	if retErr != nil {
		err = retErr
		log.Errorf("UpdateEntity failed, toLocalValue error:%s", err.Error())
		return
	}

	entityValue.Set(reflect.Indirect(retVal.Get().(reflect.Value)))
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
		log.Errorf("UpdateSliceEntity failed, error:%s", err.Error())
		return
	}

	vType, vErr := local.GetType(entityValue.Type())
	if vErr != nil {
		err = vErr
		log.Errorf("UpdateSliceEntity failed, local.GetType error:%s", err.Error())
		return
	}

	entityValue = reflect.Indirect(entityValue)
	if entityValue.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal localEntity, must be a struct localEntity")
		log.Errorf("UpdateSliceEntity failed, error:%s", err.Error())
		return
	}
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal localEntity value, can't be set")
		log.Errorf("UpdateSliceEntity failed, error:%s", err.Error())
		return
	}

	retVal, retErr := toLocalSliceValue(remoteValue, vType)
	if retErr != nil {
		err = retErr
		log.Errorf("UpdateSliceEntity failed, toLocalSliceValue error:%s", err.Error())
		return
	}

	entityValue.Set(retVal.Get().(reflect.Value))
	return
}
