package helper

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

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

	retVal, retErr := toLocalValue(remoteValue, entityValue.Type())
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal)
	return
}

func toLocalValue(objectValue *remote.ObjectValue, valueType reflect.Type) (ret reflect.Value, err error) {
	vType, vErr := local.GetType(valueType)
	if vErr != nil {
		err = vErr
		return
	}

	if objectValue.GetPkgKey() != vType.GetPkgKey() {
		err = fmt.Errorf("illegal object value, objectValue pkgKey:%s, entityType pkgKey:%s", objectValue.GetPkgKey(), vType.GetPkgKey())
		log.Errorf("toLocalValue failed, mismatch objectValue for value, err:%s", err)
		return
	}

	entityType := valueType
	if vType.IsPtrType() {
		entityType = entityType.Elem()
	}

	entityValue := reflect.New(entityType).Elem()
	for idx := 0; idx < len(objectValue.Fields); idx++ {
		curItem := objectValue.Fields[idx]
		if curItem.Get() == nil {
			continue
		}

		curField, curOK := entityType.FieldByName(curItem.GetName())
		if !curOK {
			continue
		}

		curFieldValue := entityValue.FieldByName(curItem.GetName())
		curFieldType := curField.Type
		vFieldType, vFieldErr := local.GetType(curFieldType)
		if vFieldErr != nil {
			err = vFieldErr
			log.Errorf("toLocalValue failed, field name:%s, local.GetType error, err:%s", curItem.GetName(), err.Error())
			return
		}

		for {
			if vFieldType.IsPtrType() {
				curFieldValue = reflect.Indirect(curFieldValue)
			}

			// for basic type
			if vFieldType.IsBasic() {
				curFieldValue.Set(reflect.ValueOf(curItem.Get()))
				break
			}

			// for struct type
			if model.IsStructType(vFieldType.GetValue()) {
				objPtr := curItem.Get().(*remote.ObjectValue)
				val, valErr := toLocalValue(objPtr, curFieldType)
				if valErr != nil {
					err = valErr
					log.Errorf("toLocalValue failed, field name:%s, err:%s", curItem.GetName(), err.Error())
					return
				}

				curFieldValue.Set(val)
				break
			}

			// for struct slice
			slicePtr := curItem.Get().(*remote.SliceObjectValue)
			val, valErr := toLocalSliceValue(slicePtr, curFieldType)
			if valErr != nil {
				err = valErr
				log.Errorf("toLocalSliceValue failed, field name:%s, err:%s", curItem.GetName(), err.Error())
				return
			}

			curFieldValue.Set(val)
			break
		}
	}

	if vType.IsPtrType() {
		entityValue = entityValue.Addr()
	}

	ret = entityValue
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

	entityValue = reflect.Indirect(entityValue)
	if entityValue.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal localEntity, must be a struct localEntity")
		return
	}
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal localEntity value, can't be set")
		return
	}

	retVal, retErr := toLocalSliceValue(remoteValue, entityValue.Type())
	if retErr != nil {
		err = retErr
		return
	}

	entityValue.Set(retVal)
	return
}

func toLocalSliceValue(sliceObjectValue *remote.SliceObjectValue, valueType reflect.Type) (ret reflect.Value, err error) {
	vType, vErr := local.GetType(valueType)
	if vErr != nil {
		err = vErr
		return
	}

	if sliceObjectValue.GetPkgKey() != vType.GetPkgKey() {
		err = fmt.Errorf("illegal slice object value, sliceObjectValue pkgKey:%s, sliceEntityType pkgKey:%s", sliceObjectValue.GetPkgKey(), vType.GetPkgKey())
		log.Errorf("toLocalSliceValue failed, mismatch objectValue for value, err:%s", err)
		return
	}

	sliceEntityType := valueType
	if vType.IsPtrType() {
		sliceEntityType = sliceEntityType.Elem()
	}

	sliceEntityValue := reflect.New(sliceEntityType).Elem()
	elemType := sliceEntityType.Elem()
	for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
		sliceItem := sliceObjectValue.Values[idx]
		elemVal, elemErr := toLocalValue(sliceItem, elemType)
		if elemErr != nil {
			err = fmt.Errorf("toLocalValue error [%v]", elemErr.Error())
			log.Errorf("toLocalSliceValue failed, sliceItem type:%s, elemVal type:%s, err:%s", sliceItem.GetName(), elemVal.Type().String(), err.Error())
			return
		}

		sliceEntityValue = reflect.Append(sliceEntityValue, elemVal)
	}

	ret = reflect.New(sliceEntityType).Elem()
	ret.Set(sliceEntityValue)
	if vType.IsPtrType() {
		ret = ret.Addr()
	}

	return
}
