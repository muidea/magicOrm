// Package helper 实现跨 Provider 转换，对外入口为 GetObject、GetObjectValue、GetSliceObjectValue、UpdateEntity、UpdateSliceEntity。
// 行为与设计文档 DESIGN-CONSISTENCY.md §7.4（7.4.2 Local→Remote、7.4.3 Remote→Local）、§5.5、§8.4 一致；
// 外部调用不得绕过上述 API 与 models.Model。
package helper

import (
	"fmt"
	"reflect"
	"strings"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/utils"
)

const (
	ormTag         = "orm"
	viewTag        = "view"
	constraintsTag = "constraint"
)

func getEntityType(entity any) (ret *remote.TypeImpl, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
		return
	}

	itemType := reflect.TypeOf(entity)
	return newType(itemType)
}

func newType(itemType reflect.Type) (ret *remote.TypeImpl, err *cd.Error) {
	isPtr := false
	if itemType.Kind() == reflect.Ptr {
		isPtr = true
		itemType = itemType.Elem()
	}

	typeVal, typeErr := utils.GetTypeEnum(itemType)
	if typeErr != nil {
		err = typeErr
		return
	}

	if models.IsSliceType(typeVal) {
		sliceType := itemType.Elem()
		slicePtr := false
		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem()
			slicePtr = true
		}
		ret = &remote.TypeImpl{Name: getRemoteTypeName(itemType, typeVal), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}

		sliceVal, sliceErr := utils.GetTypeEnum(sliceType)
		if sliceErr != nil {
			err = sliceErr
			return
		}
		if models.IsSliceType(sliceVal) {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal slice type, type:%s", sliceType.String()))
			return
		}

		ret.ElemType = &remote.TypeImpl{Name: getRemoteTypeName(sliceType, sliceVal), Value: sliceVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		return
	}

	ret = &remote.TypeImpl{Name: getRemoteTypeName(itemType, typeVal), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	return
}

func getRemoteTypeName(itemType reflect.Type, typeVal models.TypeDeclare) string {
	if typeVal == models.TypeBooleanValue {
		return models.TypeBooleanName
	}
	return itemType.Name()
}

func newSpec(tag reflect.StructTag) (ret *remote.SpecImpl, err *cd.Error) {
	ormSpec := tag.Get(ormTag)
	val, vErr := getOrmSpec(ormSpec)
	if vErr != nil {
		err = vErr
		return
	}

	viewSpec := tag.Get(viewTag)
	val.ViewDeclare = getViewItems(viewSpec)

	constraints := tag.Get(constraintsTag)
	if constraints != "" {
		val.Constraint = constraints
	}

	ret = &val
	return
}

func getOrmSpec(spec string) (ret remote.SpecImpl, err *cd.Error) {
	items := strings.Split(spec, " ")
	if len(items) < 1 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal spec value, val:%s", spec))
		return
	}

	ret = remote.SpecImpl{PrimaryKey: false, ValueDeclare: models.Customer}
	ret.FieldName = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case models.AutoIncrement:
			ret.ValueDeclare = models.AutoIncrement
		case models.UUID:
			ret.ValueDeclare = models.UUID
		case models.Snowflake:
			ret.ValueDeclare = models.Snowflake
		case models.DateTime:
			ret.ValueDeclare = models.DateTime
		case models.KeyTag:
			ret.PrimaryKey = true
		}
	}

	return
}

func getViewItems(spec string) (ret []models.ViewDeclare) {
	ret = []models.ViewDeclare{}
	items := strings.Split(spec, ",")
	for _, sv := range items {
		switch normalizeViewItem(sv) {
		case models.DetailView:
			ret = append(ret, models.DetailView)
		case models.LiteView:
			ret = append(ret, models.LiteView)
		}
	}
	return
}

func normalizeViewItem(spec string) string {
	return strings.TrimSpace(spec)
}

func getItemInfo(fieldType reflect.StructField) (ret *remote.Field, err *cd.Error) {
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
	item.Name = fieldType.Name
	if specImpl.GetFieldName() != "" {
		item.Name = specImpl.GetFieldName()
	}
	item.Type = typeImpl
	item.Spec = specImpl

	ret = item
	return
}

func getFieldName(fieldType reflect.StructField) (ret string, err *cd.Error) {
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

// reflectStructType 将 *T、[]T、[]*T 等类型解包为底层 struct 的 reflect.Type，供 type2Object 使用。
func reflectStructType(rt reflect.Type) reflect.Type {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() == reflect.Slice {
		rt = rt.Elem()
	}
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return rt
}

func type2Object(entityType reflect.Type) (ret *remote.Object, err *cd.Error) {
	entityType = reflectStructType(entityType)

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = typeErr
		slog.Error("type2Object newType failed", "entityType", entityType.String(), "error", err.Error())
		return
	}

	typeImpl = typeImpl.Elem().(*remote.TypeImpl)
	if !models.IsStructType(typeImpl.GetValue()) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal object type, must be a struct obj, type:%s", entityType.String()))
		slog.Error("type2Object: not a struct type", "entityType", entityType.String())
		return
	}

	impl := &remote.Object{}
	impl.Name = entityType.Name()
	impl.PkgPath = entityType.PkgPath()
	impl.Fields = []*remote.Field{}

	hasPrimaryField := false
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fItem, fErr := getItemInfo(fieldType)
		if fErr != nil {
			err = fErr
			slog.Error("type2Object getItemInfo failed", "struct", impl.GetName(), "field", fieldType.Name, "error", err.Error())
			return
		}
		if models.IsPrimaryField(fItem) {
			if hasPrimaryField {
				err = cd.NewError(cd.Unexpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldType.Name, impl.GetName()))
				slog.Error("type2Object: duplicate primary key", "struct", impl.GetName(), "field", fieldType.Name)
				return
			}

			hasPrimaryField = true
		}

		impl.Fields = append(impl.Fields, fItem)
	}

	if len(impl.Fields) == 0 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define orm field, struct name:%s", impl.GetName()))
		slog.Error("type2Object: no orm fields", "struct", impl.GetName())
		return
	}

	if !hasPrimaryField {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define primary key field, struct name:%s", impl.GetName()))
		slog.Error("type2Object: no primary key", "struct", impl.GetName())
		return
	}

	ret = impl
	return
}

// GetObject get object
func GetObject(entity any) (ret *remote.Object, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
		return
	}

	rValue := reflect.ValueOf(entity)
	entityType := rValue.Type()
	objectPtr, objectErr := type2Object(entityType)
	if objectErr != nil {
		err = objectErr
		slog.Error("GetObject type2Object failed", "entityType", entityType.String(), "error", err.Error())
		return
	}

	err = models.VerifyModel(objectPtr)
	if err != nil {
		slog.Error("GetObject VerifyModel failed", "error", err.Error())
		return
	}

	ret = objectPtr
	return
}

func getFieldValue(fieldName string, itemType *remote.TypeImpl, itemValue reflect.Value) (ret *remote.FieldValue, err *cd.Error) {
	if itemType.IsPtrType() && itemValue.IsNil() {
		ret = &remote.FieldValue{Name: fieldName, Value: nil, Assigned: false}
		return
	}

	if models.IsBasic(itemType) {
		if itemType.IsPtrType() && !utils.IsReallyValidValueForReflect(itemValue) {
			ret = &remote.FieldValue{Name: fieldName, Value: nil, Assigned: false}
			return
		}

		itemVal, itemErr := local.EncodeValue(itemValue.Interface(), itemType)
		if itemErr != nil {
			err = itemErr
			slog.Error("getFieldValue EncodeValue failed", "field", fieldName, "pkgKey", itemType.GetPkgKey(), "error", itemErr.Error())
			return
		}

		assigned := false
		if itemType.IsPtrType() || models.IsSlice(itemType) {
			assigned = true
		}
		ret = &remote.FieldValue{Name: fieldName, Value: itemVal, Assigned: assigned}
		return
	}

	if models.IsSlice(itemType) {
		objVal, objErr := getSliceObjectValue(itemValue)
		if objErr != nil {
			err = objErr
			slog.Error("getFieldValue getSliceObjectValue failed", "field", fieldName, "error", err.Error())
			return
		}

		ret = &remote.FieldValue{Name: fieldName, Value: objVal}
		return
	}

	objVal, objErr := getObjectValue(itemValue)
	if objErr != nil {
		err = objErr
		slog.Error("getFieldValue getObjectValue failed", "field", fieldName, "error", err.Error())
		return
	}

	ret = &remote.FieldValue{Name: fieldName, Value: objVal}
	return
}

func getObjectValue(entityVal reflect.Value) (ret *remote.ObjectValue, err *cd.Error) {
	entityVal = reflect.Indirect(entityVal)
	entityType := entityVal.Type()
	objType, objErr := newType(entityType)
	if objErr != nil {
		err = objErr
		slog.Error("getObjectValue newType failed", "entityType", entityType.String(), "error", err.Error())
		return
	}
	if !models.IsStruct(objType) || models.IsSlice(objType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal entity value, entity type:%s", entityType.String()))
		slog.Error("getObjectValue: illegal entity type", "entityType", entityType.String())
		return
	}

	//!! must be String, not Name
	ret = &remote.ObjectValue{Name: objType.GetName(), PkgPath: objType.GetPkgPath(), Fields: []*remote.FieldValue{}}
	for idx := 0; idx < entityVal.NumField(); idx++ {
		fieldType := entityType.Field(idx)
		fieldName, fieldErr := getFieldName(fieldType)
		if fieldErr != nil {
			err = fieldErr
			slog.Error("getObjectValue getFieldName failed", "struct", objType.GetPkgKey(), "field", fieldType.Name, "error", err.Error())
			return
		}

		typePtr, typeErr := newType(fieldType.Type)
		if typeErr != nil {
			err = typeErr
			slog.Error("getObjectValue newType for field failed", "struct", objType.GetPkgKey(), "field", fieldName, "error", err.Error())
			return
		}

		val, valErr := getFieldValue(fieldName, typePtr, entityVal.Field(idx))
		if valErr != nil {
			err = valErr
			slog.Error("getObjectValue getFieldValue failed", "struct", objType.GetPkgKey(), "field", fieldName, "error", err.Error())
			return
		}

		specPtr, specErr := newSpec(fieldType.Tag)
		if specErr != nil {
			err = specErr
			slog.Error("getObjectValue newSpec failed", "struct", objType.GetPkgKey(), "field", fieldName, "error", specErr.Error())
		}

		if specPtr.IsPrimaryKey() && val.IsValid() {
			ret.ID = fmt.Sprintf("%v", val.GetValue().Get())
		}

		ret.Fields = append(ret.Fields, val)
	}

	return
}

// GetObjectValue get object value
func GetObjectValue(entity any) (ret *remote.ObjectValue, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
		return
	}

	objInfo, objOK := entity.(remote.Object)
	if objOK {
		ret = objInfo.Interface(true).(*remote.ObjectValue)
		return
	}
	objPtr, ptrOK := entity.(*remote.Object)
	if ptrOK {
		if objPtr == nil {
			err = cd.NewError(cd.IllegalParam, "entity is nil")
			return
		}
		ret = objPtr.Interface(true).(*remote.ObjectValue)
		return
	}

	valInfo, infoOK := entity.(remote.ObjectValue)
	if infoOK {
		ret = &valInfo
		return
	}
	valPtr, ptrOK := entity.(*remote.ObjectValue)
	if ptrOK {
		if valPtr == nil {
			err = cd.NewError(cd.IllegalParam, "entity is nil")
			return
		}
		ret = valPtr
		return
	}

	entityVal := reflect.ValueOf(entity)
	if entityVal.Kind() == reflect.Ptr && entityVal.IsNil() {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
		return
	}
	ret, err = getObjectValue(entityVal)
	return
}

func getSliceObjectValue(sliceVal reflect.Value) (ret *remote.SliceObjectValue, err *cd.Error) {
	sliceType, sliceErr := newType(sliceVal.Type())
	if sliceErr != nil {
		err = sliceErr
		slog.Error("getSliceObjectValue newType failed", "sliceType", sliceVal.Type().String(), "error", err.Error())
		return
	}

	if !models.IsSliceType(sliceType.GetValue()) {
		err = cd.NewError(cd.Unexpected, "illegal slice object value")
		slog.Error("getSliceObjectValue: not a slice type", "sliceType", sliceType.GetPkgKey())
		return
	}

	elemType := sliceType.Elem()
	if !models.IsStructType(elemType.GetValue()) {
		err = cd.NewError(cd.Unexpected, "illegal slice item type")
		slog.Error("getSliceObjectValue: slice element not struct", "sliceType", sliceType.GetPkgKey())
		return
	}

	sliceVal = reflect.Indirect(sliceVal)
	ret = &remote.SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath()}
	if sliceVal.Kind() == reflect.Slice && sliceVal.IsNil() {
		return
	}

	ret.Values = []*remote.ObjectValue{}
	for idx := 0; idx < sliceVal.Len(); idx++ {
		val := sliceVal.Index(idx)
		// 设计 5.4：不考虑 []*T 中 item 为 nil；显式拒绝并返回错误，避免 getObjectValue 对零 Value 调用 Type() 导致 panic
		if elemType.IsPtrType() && val.IsNil() {
			err = cd.NewError(cd.IllegalParam, "slice element is nil, not supported for []*T")
			slog.Error("getSliceObjectValue: nil slice element", "sliceType", sliceType.GetPkgKey(), "index", idx, "error", err.Error())
			return
		}
		objVal, objErr := getObjectValue(val)
		if objErr != nil {
			err = objErr
			slog.Error("getSliceObjectValue failed", "sliceType", sliceType.GetPkgKey(), "index", idx, "error", err.Error())
			return
		}

		ret.Values = append(ret.Values, objVal)
	}

	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity any) (ret *remote.SliceObjectValue, err *cd.Error) {
	if sliceEntity == nil {
		err = cd.NewError(cd.IllegalParam, "slice entity is nil")
		return
	}

	valInfo, infoOK := sliceEntity.(remote.SliceObjectValue)
	if infoOK {
		ret = &valInfo
		return
	}
	valPtr, ptrOK := sliceEntity.(*remote.SliceObjectValue)
	if ptrOK {
		if valPtr == nil {
			err = cd.NewError(cd.IllegalParam, "slice entity is nil")
			return
		}
		ret = valPtr
		return
	}

	sliceValue := reflect.ValueOf(sliceEntity)
	if sliceValue.Kind() == reflect.Ptr && sliceValue.IsNil() {
		err = cd.NewError(cd.IllegalParam, "slice entity is nil")
		return
	}
	ret, err = getSliceObjectValue(sliceValue)
	return
}

func UpdateEntity(remoteValuePtr *remote.ObjectValue, localEntity any) (err *cd.Error) {
	if remoteValuePtr == nil {
		err = cd.NewError(cd.IllegalParam, "remote object value is nil")
		slog.Error("UpdateEntity: remote object value is nil")
		return
	}
	if localEntity == nil {
		err = cd.NewError(cd.IllegalParam, "local entity is nil")
		slog.Error("UpdateEntity: local entity is nil")
		return
	}

	localModel, localErr := local.GetEntityModel(localEntity, nil)
	if localErr != nil {
		err = localErr
		slog.Error("UpdateEntity GetEntityModel failed", "error", err.Error())
		return
	}

	err = updateLocalModel(remoteValuePtr, localModel)
	return
}

func updateLocalModel(remoteValuePtr *remote.ObjectValue, localModel models.Model) (err *cd.Error) {
	for _, fieldValue := range remoteValuePtr.Fields {
		localField := localModel.GetField(fieldValue.Name)
		if localField == nil || !fieldValue.IsValid() {
			continue
		}

		if models.IsBasicField(localField) {
			rVal, rErr := local.DecodeValue(fieldValue.Get(), localField.GetType())
			if rErr != nil {
				err = rErr
				return
			}

			err = localField.SetValue(rVal)
			if err != nil {
				slog.Error("updateLocalModel SetValue failed", "field", fieldValue.Name, "error", err.Error())
				return
			}
			continue
		}
		if models.IsSliceField(localField) {
			err = updateSliceStructField(fieldValue.Get(), localField)
			if err != nil {
				slog.Error("updateLocalModel updateSliceStructField failed", "model", localModel.GetPkgKey(), "field", fieldValue.Name, "error", err.Error())
				return
			}

			continue
		}
		if models.IsStructField(localField) {
			err = updateStructField(fieldValue.Get(), localField)
			if err != nil {
				slog.Error("updateLocalModel updateStructField failed", "model", localModel.GetPkgKey(), "field", fieldValue.Name, "error", err.Error())
				return
			}

			continue
		}
	}

	return
}

func updateSliceStructField(val any, localField models.Field) (err *cd.Error) {
	if val == nil {
		return
	}

	sliceObjectValuePtr, sliceObjectValueOK := val.(*remote.SliceObjectValue)
	if !sliceObjectValueOK {
		err = cd.NewError(cd.Unexpected, "illegal slice object value")
		slog.Error("updateSliceStructField: invalid slice value", "field", localField.GetName(), "error", err.Error())
		return
	}

	if sliceObjectValuePtr.Values == nil {
		return
	}

	err = assignLocalSliceField(localField, true)
	if err != nil {
		slog.Error("updateSliceStructField assignLocalSliceField failed", "field", localField.GetName(), "error", err.Error())
		return
	}

	for _, objectValuePtr := range sliceObjectValuePtr.Values {
		elemType := localField.GetType().Elem()
		localSubVal, _ := elemType.Interface(nil)
		localSubModel, localSubErr := local.GetValueModel(localSubVal)
		if localSubErr != nil {
			err = localSubErr
			slog.Error("updateSliceStructField GetValueModel failed", "field", localField.GetName(), "error", err.Error())
			return
		}
		err = updateLocalModel(objectValuePtr, localSubModel)
		if err != nil {
			slog.Error("updateSliceStructField updateLocalModel failed", "field", localField.GetName(), "error", err.Error())
			return
		}
		err = localField.AppendSliceValue(localSubModel.Interface(elemType.IsPtrType()))
		if err != nil {
			slog.Error("updateSliceStructField AppendSliceValue failed", "field", localField.GetName(), "error", err.Error())
			return
		}
	}

	return
}

func assignLocalSliceField(localField models.Field, assigned bool) (err *cd.Error) {
	valueType := reflect.TypeOf(localField.GetValue().Get())
	if valueType == nil {
		err = cd.NewError(cd.Unexpected, "local slice field type is nil")
		return
	}

	switch valueType.Kind() {
	case reflect.Ptr:
		sliceType := valueType.Elem()
		if sliceType.Kind() != reflect.Slice {
			err = cd.NewError(cd.Unexpected, "local field is not pointer to slice")
			return
		}
		if !assigned {
			return localField.SetValue(reflect.Zero(valueType).Interface())
		}

		sliceVal := reflect.MakeSlice(sliceType, 0, 0)
		slicePtr := reflect.New(sliceType)
		slicePtr.Elem().Set(sliceVal)
		return localField.SetValue(slicePtr.Interface())
	case reflect.Slice:
		if !assigned {
			return localField.SetValue(reflect.Zero(valueType).Interface())
		}
		return localField.SetValue(reflect.MakeSlice(valueType, 0, 0).Interface())
	default:
		err = cd.NewError(cd.Unexpected, "local field is not slice")
		return
	}
}

func updateStructField(val any, vField models.Field) (err *cd.Error) {
	if val == nil {
		return
	}

	objectValuePtr, objectValueOK := val.(*remote.ObjectValue)
	if !objectValueOK {
		err = cd.NewError(cd.Unexpected, "illegal object value")
		slog.Error("updateStructField: invalid object value", "field", vField.GetName(), "error", err.Error())
		return
	}

	elemType := vField.GetType().Elem()
	localFileVal, _ := elemType.Interface(nil)
	localModelVal, localModelErr := local.GetValueModel(localFileVal)
	if localModelErr != nil {
		err = localModelErr
		slog.Error("updateStructField GetValueModel failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	err = updateLocalModel(objectValuePtr, localModelVal)
	if err != nil {
		slog.Error("updateStructField updateLocalModel failed", "field", vField.GetName(), "error", err.Error())
		return
	}

	vField.SetValue(localModelVal.Interface(elemType.IsPtrType()))
	return
}

func UpdateSliceEntity(remoteSliceValuePtr *remote.SliceObjectValue, localSliceValue any) (err *cd.Error) {
	if remoteSliceValuePtr == nil {
		err = cd.NewError(cd.IllegalParam, "remote slice object value is nil")
		slog.Error("UpdateSliceEntity: remote slice object value is nil")
		return
	}
	if localSliceValue == nil {
		err = cd.NewError(cd.IllegalParam, "local slice entity is nil")
		slog.Error("UpdateSliceEntity: local slice entity is nil")
		return
	}

	localTypePtr, localTypeErr := local.NewType(reflect.TypeOf(localSliceValue))
	if localTypeErr != nil {
		err = localTypeErr
		slog.Error("UpdateSliceEntity NewType failed", "error", err.Error())
		return
	}
	if !models.IsSlice(localTypePtr) || !localTypePtr.IsPtrType() {
		err = cd.NewError(cd.Unexpected, "illegal local entity type")
		slog.Error("UpdateSliceEntity: illegal local entity type", "pkgKey", localTypePtr.GetPkgKey())
		return
	}

	localSliceReflect := reflect.Indirect(reflect.ValueOf(localSliceValue))
	if remoteSliceValuePtr.Values == nil {
		return
	}

	localSliceReflect.Set(reflect.MakeSlice(localSliceReflect.Type(), 0, 0))
	localValuePtr := local.NewValue(localSliceReflect)
	for idx, val := range remoteSliceValuePtr.Values {
		elemType := localTypePtr.Elem()
		localItemVal, _ := elemType.Interface(nil)
		localItemModel, localItemErr := local.GetValueModel(localItemVal)
		if localItemErr != nil {
			err = localItemErr
			slog.Error("UpdateSliceEntity GetValueModel failed", "sliceType", remoteSliceValuePtr.Name, "index", idx, "error", err.Error())
			return
		}
		err = updateLocalModel(val, localItemModel)
		if err != nil {
			slog.Error("UpdateSliceEntity updateLocalModel failed", "sliceType", remoteSliceValuePtr.Name, "index", idx, "error", err.Error())
			return
		}

		localRawVal := localItemModel.Interface(elemType.IsPtrType())
		err = localValuePtr.Append(reflect.ValueOf(localRawVal))
		if err != nil {
			slog.Error("UpdateSliceEntity Append failed", "sliceType", remoteSliceValuePtr.Name, "index", idx, "error", err.Error())
			return
		}
	}
	return
}
