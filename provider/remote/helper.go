package remote

import (
	"fmt"
	"reflect"
	"time"

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

	entityType, entityErr := newType(entitySliceVal.Type())
	if entityErr != nil {
		err = entityErr
		return
	}

	_, err = updateSliceStructValue(sliceObjectValue, entityType, entitySliceVal)
	return
}

func updateBasicValue(basicValue interface{}, tType model.Type, value reflect.Value) (ret reflect.Value, err error) {
	if util.IsNil(value) {
		return
	}

	assignFlag := false
	rVal := reflect.Indirect(reflect.ValueOf(basicValue))
	switch tType.GetValue() {
	case util.TypeBooleanField:
		if util.IsBool(rVal.Type()) {
			value.SetBool(rVal.Bool())
			assignFlag = true
		}
	case util.TypeDateTimeField:
		if util.IsString(rVal.Type()) {
			dtVal, dtErr := time.Parse("2006-01-02 15:04:05", rVal.String())
			if dtErr == nil {
				value.Set(reflect.ValueOf(dtVal))
				assignFlag = true
			}
		}
	case util.TypeFloatField, util.TypeDoubleField:
		if util.IsFloat(rVal.Type()) {
			value.SetFloat(rVal.Float())
			assignFlag = true
		}
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		if util.IsInteger(rVal.Type()) {
			value.SetInt(rVal.Int())
			assignFlag = true
		}
		if util.IsFloat(rVal.Type()) {
			value.SetInt(int64(rVal.Float()))
			assignFlag = true
		}
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		if util.IsUInteger(rVal.Type()) {
			value.SetUint(rVal.Uint())
			assignFlag = true
		}
		if util.IsFloat(rVal.Type()) {
			value.SetUint(uint64(rVal.Float()))
			assignFlag = true
		}
	case util.TypeStringField:
		if util.IsString(rVal.Type()) {
			value.SetString(rVal.String())
			assignFlag = true
		}
	default:
		err = fmt.Errorf("illegal basic value type,type:%s, value type:%s", tType.GetName(), rVal.Type().String())
	}

	if !assignFlag {
		err = fmt.Errorf("illegal basic value type,type:%s, value type:%s", tType.GetName(), rVal.Type().String())
		return
	}

	ret = value
	return
}

func updateSliceBasicValue(basicValue interface{}, tType model.Type, value reflect.Value) (ret reflect.Value, err error) {
	if util.IsNil(value) {
		return
	}

	tType = tType.Elem()
	sliceValue := reflect.ValueOf(basicValue)
	for idx := 0; idx < sliceValue.Len(); idx++ {
		val := sliceValue.Index(idx).Elem()
		switch tType.GetValue() {
		case util.TypeBooleanField:
			bVal := val.Bool()
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&bVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(bVal))
			}
		case util.TypeDateTimeField:
			strVal := val.String()
			dtVal, dtErr := time.Parse("2006-01-02 15:04:05", strVal)
			if dtErr != nil {
				err = fmt.Errorf("illegal dateTime value, val:%v", strVal)
			} else {
				if tType.IsPtrType() {
					value = reflect.Append(value, reflect.ValueOf(&dtVal))
				} else {
					value = reflect.Append(value, reflect.ValueOf(dtVal))
				}
			}
		case util.TypeFloatField:
			fVal := float32(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&fVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(fVal))
			}
		case util.TypeDoubleField:
			fVal := val.Float()
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&fVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(fVal))
			}
		case util.TypeBitField:
			iVal := int8(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&iVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(iVal))
			}
		case util.TypeSmallIntegerField:
			iVal := int16(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&iVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(iVal))
			}
		case util.TypeInteger32Field:
			iVal := int32(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&iVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(iVal))
			}
		case util.TypeIntegerField:
			iVal := int(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&iVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(iVal))
			}
		case util.TypeBigIntegerField:
			iVal := int64(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&iVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(iVal))
			}
		case util.TypePositiveBitField:
			uVal := uint8(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&uVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(uVal))
			}
		case util.TypePositiveSmallIntegerField:
			uVal := uint16(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&uVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(uVal))
			}
		case util.TypePositiveInteger32Field:
			uVal := uint32(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&uVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(uVal))
			}
		case util.TypePositiveIntegerField:
			uVal := uint(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&uVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(uVal))
			}
		case util.TypePositiveBigIntegerField:
			uVal := uint64(val.Float())
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&uVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(uVal))
			}
		case util.TypeStringField:
			strVal := val.String()
			if tType.IsPtrType() {
				value = reflect.Append(value, reflect.ValueOf(&strVal))
			} else {
				value = reflect.Append(value, reflect.ValueOf(strVal))
			}
		default:
			err = fmt.Errorf("invalud slice item type, type:%s", tType.GetName())
		}

		if err != nil {
			break
		}
	}

	if err != nil {
		return
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

	entityType := value.Type()
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		curItem := objectValue.Items[idx]
		if curItem.Value == nil {
			continue
		}
		curValue := reflect.Indirect(value.Field(idx))
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
			if util.IsBasicType(tType.GetValue()) {
				_, err = updateBasicValue(curItem.Value, tType, curValue)
				if err != nil {
					log.Errorf("updateBasicValue failed, fieldName:%s", field.Name)
					return
				}
				break
			}

			// for struct type
			if util.IsStructType(tType.GetValue()) {
				_, err = updateStructValue(curItem.Value.(*ObjectValue), tType, curValue)
				if err != nil {
					log.Errorf("convertStructItemValue failed, fieldName:%s", field.Name)
					return
				}
				break
			}

			// for basic slice
			if tType.IsBasic() {
				val, valErr := updateSliceBasicValue(curItem.Value, tType, curValue)
				if valErr != nil {
					err = valErr
					log.Errorf("updateSliceBasicValue failed, fieldName:%s", field.Name)
					return
				}

				curValue.Set(val)
				break
			}

			// for struct slice
			val, valErr := updateSliceStructValue(curItem.Value.(*SliceObjectValue), tType, curValue)
			if valErr != nil {
				err = valErr
				log.Errorf("updateSliceStructValue failed, fieldName:%s", field.Name)
				return
			}
			curValue.Set(val)
			break
		}
	}

	ret = value
	return
}

func updateSliceStructValue(sliceObjectValue *SliceObjectValue, vType model.Type, value reflect.Value) (ret reflect.Value, err error) {
	if vType.GetName() != sliceObjectValue.GetName() || vType.GetPkgPath() != sliceObjectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, sliceObjectValue name:%s, entityType name:%s", sliceObjectValue.GetName(), vType.GetName())
		log.Errorf("mismatch sliceObjectValue for value, err:%s", err)
		return
	}

	elemType := value.Type().Elem()
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
		if isPtr {
			elemVal = elemVal.Addr()
		}

		value = reflect.Append(value, elemVal)
	}

	ret = value
	return
}
