package provider

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/util"
)

// UpdateLocalEntity update object value -> entity
func UpdateLocalEntity(remoteValue *remote.ObjectValue, localEntity interface{}) (err error) {
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

	retVal, retErr := getLocalValue(remoteValue, entityValue.Type())
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal)
	return
}

func getLocalValue(objectValue *remote.ObjectValue, valueType reflect.Type) (ret reflect.Value, err error) {
	vType, vErr := local.GetType(valueType)
	if vErr != nil {
		err = vErr
		return
	}

	if vType.GetName() != objectValue.GetName() || vType.GetPkgPath() != objectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", objectValue.GetName(), vType.GetName())
		log.Errorf("mismatch objectValue for value, err:%s", err)
		return
	}
	entityType := valueType
	if vType.IsPtrType() {
		entityType = entityType.Elem()
	}

	entityValue := reflect.New(entityType).Elem()
	items := objectValue.Fields
	fieldNum := len(objectValue.Fields)
	if fieldNum == 0 {
		return
	}

	for idx := 0; idx < fieldNum; idx++ {
		curItem := items[idx]
		if curItem.Value == nil {
			continue
		}

		curFieldType := entityType.Field(idx).Type
		curFieldValue := entityValue.Field(idx)
		vFieldType, curErr := local.GetType(curFieldType)
		if curErr != nil {
			err = curErr
			log.Errorf("illegal struct curField, err:%s", err.Error())
			return
		}

		curField := entityType.Field(idx)
		for {
			// for basic type
			if vFieldType.IsBasic() {
				val, valErr := local.GetHelper().Decode(curItem.Value, vFieldType)
				if valErr != nil {
					err = valErr
					log.Errorf("updateBasicValue failed, fieldName:%s, err:%s", curField.Name, err.Error())
					return
				}

				if !val.IsNil() {
					curFieldValue.Set(val.Get())
				}
				break
			}

			// for struct type
			if model.IsStructType(vFieldType.GetValue()) {
				objPtr := curItem.Value.(*remote.ObjectValue)
				val, valErr := getLocalValue(objPtr, curFieldType)
				if valErr != nil {
					err = valErr
					log.Errorf("convertStructItemValue failed, fieldName:%s, err:%s", curField.Name, err.Error())
					return
				}

				if !util.IsNil(val) {
					curFieldValue.Set(val)
				}
				break
			}

			// for struct slice
			slicePtr := curItem.Value.(*remote.SliceObjectValue)
			val, valErr := getLocalSliceValue(slicePtr, curFieldType)
			if valErr != nil {
				err = valErr
				log.Errorf("updateSliceStructValue failed, fieldName:%s, err:%s", curField.Name, err.Error())
				return
			}

			if !util.IsNil(val) {
				curFieldValue.Set(val)
			}
			break
		}
	}

	if vType.IsPtrType() {
		entityValue = entityValue.Addr()
	}

	ret = entityValue
	return
}

// UpdateLocalSliceEntity update slice object value -> entitySlice
func UpdateLocalSliceEntity(sliceObjectValue *remote.SliceObjectValue, sliceEntity interface{}) (err error) {
	if !sliceObjectValue.IsAssigned() {
		return
	}

	entityVal := reflect.ValueOf(sliceEntity)
	if entityVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal slice entity value")
		return
	}
	entityVal = reflect.Indirect(entityVal)
	if entityVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal sliceEntity")
		return
	}
	if !entityVal.CanSet() {
		err = fmt.Errorf("illegal sliceEntity value, can't be set")
		return
	}

	retVal, retErr := getLocalSliceValue(sliceObjectValue, entityVal.Type())
	if retErr != nil {
		err = retErr
		return
	}

	entityVal.Set(retVal)
	return
}

func getLocalSliceValue(sliceObjectValue *remote.SliceObjectValue, valueType reflect.Type) (ret reflect.Value, err error) {
	vType, vErr := local.GetType(valueType)
	if vErr != nil {
		err = vErr
		return
	}

	if vType.GetName() != sliceObjectValue.GetName() || vType.GetPkgPath() != sliceObjectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, sliceObjectValue name:%s, entityType name:%s", sliceObjectValue.GetName(), vType.GetName())
		log.Errorf("mismatch sliceObjectValue for value, err:%s", err)
		return
	}

	entityType := valueType
	if vType.IsPtrType() {
		entityType = entityType.Elem()
	}

	entityValue := reflect.New(entityType).Elem()
	sliceValue := sliceObjectValue.Values
	sliceSize := len(sliceValue)
	if sliceSize == 0 {
		return
	}

	elemType := entityType.Elem()
	for idx := 0; idx < sliceSize; idx++ {
		sliceItem := sliceValue[idx]
		elemVal, elemErr := getLocalValue(sliceItem, elemType)
		if elemErr != nil {
			err = elemErr
			log.Errorf("updateStructValue failed, sliceItem type:%s, elemVal type:%s", sliceItem.GetName(), elemVal.Type().String())
			return
		}
		entityValue = reflect.Append(entityValue, elemVal)
	}

	ret = reflect.New(entityType).Elem()
	ret.Set(entityValue)
	if vType.IsPtrType() {
		ret = ret.Addr()
	}

	return
}
