package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// UpdateEntity update object value -> entity
func UpdateEntity(objectValue *ObjectValue, entity interface{}) (err error) {
	if !objectValue.IsAssigned() {
		return
	}

	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal entity value, must be a pointer entity")
		return
	}

	entityValue = reflect.Indirect(entityValue)
	if entityValue.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal entity, must be a struct entity")
		return
	}
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal entity value, can't be set")
		return
	}

	entityType, entityErr := newType(entityValue.Type())
	if entityErr != nil {
		err = entityErr
		return
	}

	retVal, retErr := updateStructValue(objectValue, entityType, entityValue)
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal)
	return
}

// UpdateSliceEntity update slice object value -> entitySlice
func UpdateSliceEntity(sliceObjectValue *SliceObjectValue, entitySlice interface{}) (err error) {
	if !sliceObjectValue.IsAssigned() {
		return
	}

	entitySliceVal := reflect.ValueOf(entitySlice)
	if entitySliceVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal slice entity value")
		return
	}
	entitySliceVal = reflect.Indirect(entitySliceVal)
	if entitySliceVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal entitySlice")
		return
	}
	if !entitySliceVal.CanSet() {
		err = fmt.Errorf("illegal entitySlice value, can't be set")
		return
	}

	entitySliceType, entitySliceErr := newType(entitySliceVal.Type())
	if entitySliceErr != nil {
		err = entitySliceErr
		return
	}

	retVal, retErr := updateSliceStructValue(sliceObjectValue, entitySliceType, entitySliceVal)
	if retErr != nil {
		err = retErr
		return
	}

	entitySliceVal.Set(retVal)
	return
}

func updateBasicValue(basicValue interface{}, tType model.Type, value reflect.Value) (ret reflect.Value, err error) {
	if util.IsNil(value) {
		return
	}

	value = reflect.Indirect(value)
	// special for dateTime
	if util.TypeDateTimeField == tType.Elem().GetValue() {
		rValue := reflect.ValueOf(basicValue)
		rValue = reflect.Indirect(rValue)
		if rValue.Kind() == reflect.Interface {
			rValue = rValue.Elem()
		}
		rValue = reflect.Indirect(rValue)
		if rValue.Kind() != reflect.String {
			err = fmt.Errorf("illegal dateTime value")
			return
		}
		rVal, rErr := decodeDateTime(rValue.String(), tType)
		if rErr != nil {
			err = rErr
			return
		}

		value.Set(rVal)
	} else {
		vVal, vErr := _helper.Decode(basicValue, tType)
		if vErr != nil {
			err = vErr
			return
		}
		rVal := reflect.Indirect(reflect.ValueOf(vVal.Get()))
		if rVal.Kind() == reflect.Interface {
			rVal = rVal.Elem()
		}

		rVal = reflect.Indirect(rVal)
		value.Set(rVal)
	}

	if tType.IsPtrType() {
		value = value.Addr()
	}

	ret = value
	return
}

func updateStructValue(objectValue *ObjectValue, vType model.Type, value reflect.Value) (ret reflect.Value, err error) {
	if vType.GetName() != objectValue.GetName() || vType.GetPkgPath() != objectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", objectValue.GetName(), vType.GetName())
		log.Errorf("mismatch objectValue for value, err:%s", err)
		return
	}

	entityValue := reflect.Indirect(value)
	entityType := entityValue.Type()
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		curItem := objectValue.Items[idx]
		if curItem.Value == nil {
			continue
		}
		curValue := entityValue.Field(idx)
		if util.IsNil(curValue) {
			continue
		}

		field := entityType.Field(idx)
		tType, tErr := newType(field.Type)
		if tErr != nil {
			err = tErr
			log.Errorf("illegal struct field, err:%s", err.Error())
			return
		}

		for {
			// for basic type
			if tType.IsBasic() {
				val, valErr := updateBasicValue(curItem.Value, tType, curValue)
				if valErr != nil {
					err = valErr
					log.Errorf("updateBasicValue failed, fieldName:%s, err:%s", field.Name, err.Error())
					return
				}

				curValue.Set(val)
				break
			}

			// for struct type
			if util.IsStructType(tType.GetValue()) {
				val, valErr := updateStructValue(curItem.Value.(*ObjectValue), tType, curValue)
				if valErr != nil {
					err = valErr
					log.Errorf("convertStructItemValue failed, fieldName:%s, err:%s", field.Name, err.Error())
					return
				}

				curValue.Set(val)
				break
			}

			// for struct slice
			val, valErr := updateSliceStructValue(curItem.Value.(*SliceObjectValue), tType, curValue)
			if valErr != nil {
				err = valErr
				log.Errorf("updateSliceStructValue failed, fieldName:%s, err:%s", field.Name, err.Error())
				return
			}

			curValue.Set(val)
			break
		}
	}

	if vType.IsPtrType() {
		entityValue = entityValue.Addr()
	}

	ret = entityValue
	return
}

func updateSliceStructValue(sliceObjectValue *SliceObjectValue, vType model.Type, value reflect.Value) (ret reflect.Value, err error) {
	if vType.GetName() != sliceObjectValue.GetName() || vType.GetPkgPath() != sliceObjectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, sliceObjectValue name:%s, entityType name:%s", sliceObjectValue.GetName(), vType.GetName())
		log.Errorf("mismatch sliceObjectValue for value, err:%s", err)
		return
	}

	value = reflect.Indirect(value)
	entityType := value.Type()
	entityValue := reflect.New(entityType).Elem()
	elemType := entityType.Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}
	sliceValue := sliceObjectValue.Values
	for idx := 0; idx < len(sliceValue); idx++ {
		sliceItem := sliceValue[idx]
		elemVal := reflect.New(elemType).Elem()
		elemVal, err = updateStructValue(sliceItem, vType.Elem(), elemVal)
		if err != nil {
			log.Errorf("updateStructValue failed, sliceItem type:%s, elemVal type:%s", sliceItem.GetName(), elemVal.Type().String())
			return
		}
		entityValue = reflect.Append(entityValue, elemVal)
	}

	value.Set(entityValue)
	if vType.IsPtrType() {
		value = value.Addr()
	}

	ret = value
	return
}
