package provider

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/util"
)

// UpdateEntity update object value -> entity
func UpdateEntity(objectValue *remote.ObjectValue, entity interface{}) (err error) {
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

	retVal, retErr := getObjectValue(objectValue, entityValue.Type())
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal)
	return
}

// UpdateSliceEntity update slice object value -> entitySlice
func UpdateSliceEntity(sliceObjectValue *remote.SliceObjectValue, entitySlice interface{}) (err error) {
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

	retVal, retErr := getSliceStructValue(sliceObjectValue, entitySliceVal.Type())
	if retErr != nil {
		err = retErr
		return
	}

	entitySliceVal.Set(retVal)
	return
}

func getSliceStructValue(sliceObjectValue *remote.SliceObjectValue, valueType reflect.Type) (ret reflect.Value, err error) {
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
		elemVal, elemErr := getObjectValue(sliceItem, elemType)
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

func getObjectValue(objectValue *remote.ObjectValue, valueType reflect.Type) (ret reflect.Value, err error) {
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
	items := objectValue.Items
	fieldNum := len(objectValue.Items)
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
			if util.IsStructType(vFieldType.GetValue()) {
				objPtr := curItem.Value.(*remote.ObjectValue)
				val, valErr := getObjectValue(objPtr, curFieldType)
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
			val, valErr := getSliceStructValue(slicePtr, curFieldType)
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
