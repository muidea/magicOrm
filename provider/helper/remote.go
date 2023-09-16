package helper

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote"
	pu "github.com/muidea/magicOrm/provider/util"
)

func newType(itemType reflect.Type) (ret *remote.TypeImpl, err error) {
	isPtr := false
	if itemType.Kind() == reflect.Ptr {
		isPtr = true
		itemType = itemType.Elem()
	}

	typeVal, typeErr := pu.GetTypeEnum(itemType)
	if typeErr != nil {
		err = typeErr
		return
	}

	if model.IsSliceType(typeVal) {
		sliceType := itemType.Elem()
		slicePtr := false
		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem()
			slicePtr = true
		}
		ret = &remote.TypeImpl{Name: sliceType.Name(), Value: typeVal, PkgPath: sliceType.PkgPath(), IsPtr: isPtr}

		sliceVal, sliceErr := pu.GetTypeEnum(sliceType)
		if sliceErr != nil {
			err = sliceErr
			return
		}
		if model.IsSliceType(sliceVal) {
			err = fmt.Errorf("illegal slice type, type:%s", sliceType.String())
			return
		}

		ret.ElemType = &remote.TypeImpl{Name: sliceType.Name(), Value: sliceVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		return
	}

	ret = &remote.TypeImpl{Name: itemType.Name(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	//ret.ElemType = &TypeImpl{Name: itemType.Name(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	return
}

func newSpec(tag reflect.StructTag) (ret *remote.SpecImpl, err error) {
	spec := tag.Get("orm")
	val, vErr := getSpec(spec)
	if vErr != nil {
		err = vErr
		return
	}

	ret = &val
	return
}

func getSpec(spec string) (ret remote.SpecImpl, err error) {
	items := strings.Split(spec, " ")
	if len(items) < 1 {
		err = fmt.Errorf("illegal spec value, val:%s", spec)
		return
	}

	ret = remote.SpecImpl{PrimaryKey: false, ValueDeclare: model.Customer}
	ret.FieldName = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case pu.Auto:
			ret.ValueDeclare = model.AutoIncrement
		case pu.UUID:
			ret.ValueDeclare = model.UUID
		case pu.SnowFlake:
			ret.ValueDeclare = model.SnowFlake
		case pu.DateTime:
			ret.ValueDeclare = model.DateTime
		case pu.Key:
			ret.PrimaryKey = true
		}
	}

	return
}

func getItemInfo(idx int, fieldType reflect.StructField) (ret *remote.Field, err error) {
	typeImpl, typeErr := newType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	specImpl, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	item := &remote.Field{}
	item.Index = idx
	item.Name = fieldType.Name
	if specImpl.GetFieldName() != "" {
		item.Name = specImpl.GetFieldName()
	}
	item.Type = typeImpl
	item.Spec = specImpl

	ret = item
	return
}

func getFieldName(fieldType reflect.StructField) (ret string, err error) {
	specPtr, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	fieldName := fieldType.Name
	if specPtr.GetFieldName() != "" {
		fieldName = specPtr.GetFieldName()
	}

	ret = fieldName
	return
}

// type2Object type2Object
func type2Object(entityType reflect.Type) (ret *remote.Object, err error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Interface {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal entity type, must be a struct obj, type:%s", entityType.String())
		return
	}
	if !model.IsStructType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}

	impl := &remote.Object{}
	impl.Name = entityType.Name()
	impl.PkgPath = entityType.PkgPath()
	impl.Fields = []*remote.Field{}

	hasPrimaryKey := false
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fItem, fErr := getItemInfo(idx, fieldType)
		if fErr != nil {
			err = fErr
			return
		}
		if fItem.IsPrimaryKey() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldType.Name, impl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		impl.Fields = append(impl.Fields, fItem)
	}

	if len(impl.Fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", impl.GetName())
		return
	}

	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", impl.GetName())
		return
	}

	ret = impl
	return
}

// GetObject get object
func GetObject(entity interface{}) (ret *remote.Object, err error) {
	entityType := reflect.ValueOf(entity).Type()
	ret, err = type2Object(entityType)
	if err != nil {
		log.Errorf("type2Object failed, raw type:%s, err:%s", entityType.String(), err.Error())
	}

	return
}

func getBasicValue(itemValue reflect.Value) (ret any, err error) {
	itemValue = reflect.Indirect(itemValue)
	switch itemValue.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		ret = itemValue.Interface()
	case reflect.Struct:
		if !itemValue.IsZero() {
			dtVal, dtOK := itemValue.Interface().(time.Time)
			if dtOK {
				ret = dtVal.Format(util.CSTLayout)
			} else {
				err = fmt.Errorf("illegal basic value, value type:%v", itemValue.Type().String())
			}
		} else {
			ret = ""
		}
	default:
		err = fmt.Errorf("illegal basic value, value type:%v", itemValue.Type().String())
	}

	return
}

func getBasicSliceValue(itemValue reflect.Value) (ret any, err error) {
	itemValue = reflect.Indirect(itemValue)
	switch itemValue.Kind() {
	case reflect.Slice:
	default:
		err = fmt.Errorf("illegal basic slice value, value type:%v", itemValue.Type().String())
	}
	if err != nil {
		log.Errorf("getBasicSliceValue failed, err:%s", err.Error())
		return
	}

	if itemValue.IsNil() {
		return
	}

	subValList := []any{}
	for idx := 0; idx < itemValue.Len(); idx++ {
		subVal, subErr := getBasicValue(itemValue.Index(idx))
		if subErr != nil {
			err = subErr
			return
		}

		subValList = append(subValList, subVal)
	}

	ret = subValList
	return
}

func getFieldValue(fieldName string, itemType *remote.TypeImpl, itemValue reflect.Value) (ret *remote.FieldValue, err error) {
	if itemType.IsPtrType() && itemValue.IsNil() {
		ret = &remote.FieldValue{Name: fieldName, Value: nil}
		return
	}

	if !model.IsSliceType(itemType.GetValue()) {
		if itemType.IsBasic() {
			itemVal, itemErr := getBasicValue(itemValue)
			if itemErr != nil {
				err = itemErr
				return
			}

			ret = &remote.FieldValue{Name: fieldName, Value: itemVal}
			return
		}

		objVal, objErr := getObjectValue(itemValue)
		if objErr != nil {
			err = objErr
			log.Errorf("GetObjectValue failed, raw type:%s, err:%s", itemType.GetName(), err.Error())
			return
		}

		ret = &remote.FieldValue{Name: fieldName, Value: objVal}
		return
	}

	if itemType.IsBasic() {
		itemVal, itemErr := getBasicSliceValue(itemValue)
		if itemErr != nil {
			err = itemErr
			return
		}

		ret = &remote.FieldValue{Name: fieldName, Value: itemVal}
		return
	}

	objVal, objErr := getSliceObjectValue(itemValue)
	if objErr != nil {
		err = objErr
		log.Errorf("getSliceObjectValue failed, raw type:%s, err:%s", itemType.GetName(), err.Error())
		return
	}

	ret = &remote.FieldValue{Name: fieldName, Value: objVal}
	return
}

func getObjectValue(entityVal reflect.Value) (ret *remote.ObjectValue, err error) {
	entityVal = reflect.Indirect(entityVal)
	entityType := entityVal.Type()
	objType, objErr := newType(entityType)
	if objErr != nil {
		err = objErr
		return
	}
	if !model.IsStructType(objType.GetValue()) {
		err = fmt.Errorf("illegal entity, entity type:%s", entityType.String())
		return
	}

	//!! must be String, not Name
	ret = &remote.ObjectValue{Name: objType.GetName(), PkgPath: objType.GetPkgPath(), Fields: []*remote.FieldValue{}}
	fieldNum := entityVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fieldName, fieldErr := getFieldName(fieldType)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("get entity failed, field name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		typePtr, typeErr := newType(fieldType.Type)
		if typeErr != nil {
			err = typeErr
			log.Errorf("get entity type failed, field name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		val, valErr := getFieldValue(fieldName, typePtr, entityVal.Field(idx))
		if valErr != nil {
			err = valErr
			log.Errorf("getFieldValue failed, field name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		ret.Fields = append(ret.Fields, val)
	}

	return
}

// GetObjectValue get object value
func GetObjectValue(entity interface{}) (ret *remote.ObjectValue, err error) {
	entityVal := reflect.ValueOf(entity)
	ret, err = getObjectValue(entityVal)
	return
}

func getSliceObjectValue(sliceVal reflect.Value) (ret *remote.SliceObjectValue, err error) {
	if pu.IsNil(sliceVal) {
		return
	}

	sliceType, sliceErr := newType(sliceVal.Type())
	if sliceErr != nil {
		err = fmt.Errorf("get slice object type failed, err:%s", err.Error())
		log.Errorf("GetSliceObjectValue failed, slice type name:%s", sliceType.GetName())
		return
	}

	if !model.IsSliceType(sliceType.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		log.Errorf("illegal slice type, slice type name:%s", sliceType.GetName())
		return
	}

	elemType := sliceType.Elem()
	if !model.IsStructType(elemType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		log.Errorf("illegal slice elem type, type%s", elemType.GetName())
		return
	}

	sliceVal = reflect.Indirect(sliceVal)
	ret = &remote.SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), Values: []*remote.ObjectValue{}}
	for idx := 0; idx < sliceVal.Len(); idx++ {
		val := sliceVal.Index(idx)
		objVal, objErr := getObjectValue(val)
		if objErr != nil {
			err = objErr
			log.Errorf("GetObjectValue failed, type%s, err:%s", val.Type().String(), err.Error())
			return
		}

		ret.Values = append(ret.Values, objVal)
	}

	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity interface{}) (ret *remote.SliceObjectValue, err error) {
	sliceValue := reflect.ValueOf(sliceEntity)
	ret, err = getSliceObjectValue(sliceValue)
	return
}
