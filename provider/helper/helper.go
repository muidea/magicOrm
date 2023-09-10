package helper

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/provider/util"
)

func newType(itemType reflect.Type) (ret *remote.TypeImpl, err error) {
	isPtr := false
	if itemType.Kind() == reflect.Ptr {
		isPtr = true
		itemType = itemType.Elem()
	}

	typeVal, typeErr := util.GetTypeEnum(itemType)
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

		sliceVal, sliceErr := util.GetTypeEnum(sliceType)
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
		case util.Auto:
			ret.ValueDeclare = model.AutoIncrement
		case util.UUID:
			ret.ValueDeclare = model.UUID
		case util.SnowFlake:
			ret.ValueDeclare = model.SnowFlake
		case util.DateTime:
			ret.ValueDeclare = model.DateTime
		case util.Key:
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

// GetObject GetObject
func GetObject(entity interface{}) (ret *remote.Object, err error) {
	entityType := reflect.ValueOf(entity).Type()
	ret, err = type2Object(entityType)
	if err != nil {
		log.Errorf("type2Object failed, raw type:%s, err:%s", entityType.String(), err.Error())
	}

	return
}

func getFieldValue(fieldName string, itemType *remote.TypeImpl, itemValue reflect.Value) (ret *remote.FieldValue, err error) {
	if itemType.IsPtrType() && itemValue.IsNil() {
		ret = &remote.FieldValue{Name: fieldName, Value: nil}
		return
	}

	if itemType.IsBasic() {
		ret = &remote.FieldValue{Name: fieldName, Value: itemValue.Interface()}
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

func getSliceFieldValue(fieldName string, itemType *remote.TypeImpl, itemValue reflect.Value) (ret *remote.FieldValue, err error) {
	ret = &remote.FieldValue{Name: fieldName}
	if itemValue.IsNil() {
		ret = &remote.FieldValue{Name: fieldName, Value: nil}
		return
	}

	elemType := itemType.Elem()
	if elemType.IsBasic() {
		ret = &remote.FieldValue{Name: fieldName, Value: itemValue.Interface()}
		return
	}

	sliceObjectVal := []*remote.ObjectValue{}
	rawVal := reflect.Indirect(itemValue)
	for idx := 0; idx < rawVal.Len(); idx++ {
		itemVal := rawVal.Index(idx)
		objVal, objErr := getObjectValue(itemVal)
		if objErr != nil {
			err = objErr
			log.Errorf("encodeDateTimeValue failed, err:%s", err.Error())
			return
		}

		sliceObjectVal = append(sliceObjectVal, objVal)
	}
	ret.Value = &remote.SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), Values: sliceObjectVal}
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

		if typePtr.GetValue() != model.TypeSliceValue {
			val, valErr := getFieldValue(fieldName, typePtr, entityVal.Field(idx))
			if valErr != nil {
				err = valErr
				log.Errorf("getFieldValue failed, field name:%s, err:%s", fieldType.Name, err.Error())
				return
			}
			ret.Fields = append(ret.Fields, val)
		} else {
			val, valErr := getSliceFieldValue(fieldName, typePtr, entityVal.Field(idx))
			if valErr != nil {
				err = valErr
				log.Errorf("getSliceFieldValue failed, field name:%s, err:%s", fieldType.Name, err.Error())
				return
			}
			ret.Fields = append(ret.Fields, val)
		}
	}

	return
}

// GetObjectValue get object value
func GetObjectValue(entity interface{}) (ret *remote.ObjectValue, err error) {
	entityVal := reflect.ValueOf(entity)
	ret, err = getObjectValue(entityVal)
	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity interface{}) (ret *remote.SliceObjectValue, err error) {
	sliceValue := reflect.ValueOf(sliceEntity)
	sliceType, sliceErr := newType(sliceValue.Type())
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

	ret = &remote.SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), Values: []*remote.ObjectValue{}}
	sliceValue = reflect.Indirect(sliceValue)
	for idx := 0; idx < sliceValue.Len(); idx++ {
		val := sliceValue.Index(idx)
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
