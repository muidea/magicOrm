package helper

import (
	"fmt"
	"reflect"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/utils"
	"log/slog"
)

const (
	ormTag         = "orm"
	viewTag        = "view"
	constraintsTag = "constraint"
)

func getEntityType(entity any) (ret *remote.TypeImpl, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.Unexpected, "illegal entity value")
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
		ret = &remote.TypeImpl{Name: sliceType.Name(), Value: typeVal, PkgPath: sliceType.PkgPath(), IsPtr: isPtr}

		sliceVal, sliceErr := utils.GetTypeEnum(sliceType)
		if sliceErr != nil {
			err = sliceErr
			return
		}
		if models.IsSliceType(sliceVal) {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal slice type, type:%s", sliceType.String()))
			return
		}

		ret.ElemType = &remote.TypeImpl{Name: sliceType.Name(), Value: sliceVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		return
	}

	ret = &remote.TypeImpl{Name: itemType.Name(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	return
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
		switch strings.TrimSpace(sv) {
		case models.DetailView:
			ret = append(ret, models.DetailView)
		case models.LiteView:
			ret = append(ret, models.LiteView)
		}
	}
	return
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

func type2Object(entityType reflect.Type) (ret *remote.Object, err *cd.Error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Slice {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = typeErr

		slog.Error("message")
		return
	}

	typeImpl = typeImpl.Elem().(*remote.TypeImpl)
	if !models.IsStructType(typeImpl.GetValue()) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal object type, must be a struct obj, type:%s", entityType.String()))

		slog.Error("message")
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

			slog.Error("message")
			return
		}
		if models.IsPrimaryField(fItem) {
			if hasPrimaryField {
				err = cd.NewError(cd.Unexpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldType.Name, impl.GetName()))

				slog.Error("message")
				return
			}

			hasPrimaryField = true
		}

		impl.Fields = append(impl.Fields, fItem)
	}

	if len(impl.Fields) == 0 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define orm field, struct name:%s", impl.GetName()))

		slog.Error("message")
		return
	}

	if !hasPrimaryField {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no define primary key field, struct name:%s", impl.GetName()))

		slog.Error("message")
		return
	}

	ret = impl
	return
}

// GetObject get object
func GetObject(entity any) (ret *remote.Object, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.Unexpected, "nil object value")
		return
	}

	rValue := reflect.ValueOf(entity)
	entityType := rValue.Type()
	objectPtr, objectErr := type2Object(entityType)
	if objectErr != nil {
		err = objectErr
		slog.Error("error occurred", "error", err.Error())
		return
	}

	err = models.VerifyModel(objectPtr)
	if err != nil {

		slog.Error("message")
		return
	}

	ret = objectPtr
	return
}

func getFieldValue(fieldName string, itemType *remote.TypeImpl, itemValue reflect.Value) (ret *remote.FieldValue, err *cd.Error) {
	if itemType.IsPtrType() && itemValue.IsNil() {
		ret = &remote.FieldValue{Name: fieldName, Value: nil}
		return
	}

	if models.IsBasic(itemType) {
		if itemType.IsPtrType() && !utils.IsReallyValidValueForReflect(itemValue) {
			ret = &remote.FieldValue{Name: fieldName, Value: nil}
			return
		}

		itemVal, itemErr := local.EncodeValue(itemValue.Interface(), itemType)
		if itemErr != nil {
			err = itemErr
			slog.Error("error occurred", "error", fieldName, "pkgKey", itemType.GetPkgKey(), "itemErr", itemErr.Error())
			return
		}

		ret = &remote.FieldValue{Name: fieldName, Value: itemVal}
		return
	}

	if models.IsSlice(itemType) {
		objVal, objErr := getSliceObjectValue(itemValue)
		if objErr != nil {
			err = objErr

			slog.Error("error occurred", "error", err.Error())
			return
		}

		ret = &remote.FieldValue{Name: fieldName, Value: objVal}
		return
	}

	objVal, objErr := getObjectValue(itemValue)
	if objErr != nil {
		err = objErr

		slog.Error("error occurred", "error", err.Error())
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
		slog.Error("error occurred", "error", err.Error())
		return
	}
	if !models.IsStruct(objType) || models.IsSlice(objType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal entity value, entity type:%s", entityType.String()))

		slog.Error("message")
		return
	}

	//!! must be String, not Name
	ret = &remote.ObjectValue{Name: objType.GetName(), PkgPath: objType.GetPkgPath(), Fields: []*remote.FieldValue{}}
	for idx := 0; idx < entityVal.NumField(); idx++ {
		fieldType := entityType.Field(idx)
		fieldName, fieldErr := getFieldName(fieldType)
		if fieldErr != nil {
			err = fieldErr

			slog.Error("message")
			return
		}

		typePtr, typeErr := newType(fieldType.Type)
		if typeErr != nil {
			err = typeErr

			slog.Error("message")
			return
		}

		val, valErr := getFieldValue(fieldName, typePtr, entityVal.Field(idx))
		if valErr != nil {
			err = valErr

			slog.Error("message")
			return
		}

		specPtr, specErr := newSpec(fieldType.Tag)
		if specErr != nil {
			err = specErr

			slog.Error("message")
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
		err = cd.NewError(cd.Unexpected, "nil object value")
		return
	}

	objInfo, objOK := entity.(remote.Object)
	if objOK {
		ret = objInfo.Interface(true).(*remote.ObjectValue)
		return
	}
	objPtr, ptrOK := entity.(*remote.Object)
	if ptrOK {
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
		ret = valPtr
		return
	}

	entityVal := reflect.ValueOf(entity)
	ret, err = getObjectValue(entityVal)
	return
}

func getSliceObjectValue(sliceVal reflect.Value) (ret *remote.SliceObjectValue, err *cd.Error) {
	sliceType, sliceErr := newType(sliceVal.Type())
	if sliceErr != nil {
		err = sliceErr
		slog.Error("message")
		return
	}

	if !models.IsSliceType(sliceType.GetValue()) {
		err = cd.NewError(cd.Unexpected, "illegal slice object value")

		slog.Error("message")
		return
	}

	elemType := sliceType.Elem()
	if !models.IsStructType(elemType.GetValue()) {
		err = cd.NewError(cd.Unexpected, "illegal slice item type")

		slog.Error("message")
		return
	}

	sliceVal = reflect.Indirect(sliceVal)
	ret = &remote.SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), Values: []*remote.ObjectValue{}}
	for idx := 0; idx < sliceVal.Len(); idx++ {
		val := sliceVal.Index(idx)
		objVal, objErr := getObjectValue(val)
		if objErr != nil {
			err = objErr

			slog.Error("message")
			return
		}

		ret.Values = append(ret.Values, objVal)
	}

	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity any) (ret *remote.SliceObjectValue, err *cd.Error) {
	if sliceEntity == nil {
		err = cd.NewError(cd.Unexpected, "nil slice object value")
		return
	}

	valInfo, infoOK := sliceEntity.(remote.SliceObjectValue)
	if infoOK {
		ret = &valInfo
		return
	}
	valPtr, ptrOK := sliceEntity.(*remote.SliceObjectValue)
	if ptrOK {
		ret = valPtr
		return
	}

	sliceValue := reflect.ValueOf(sliceEntity)
	ret, err = getSliceObjectValue(sliceValue)
	return
}

func UpdateEntity(remoteValuePtr *remote.ObjectValue, localEntity any) (err *cd.Error) {
	if remoteValuePtr == nil {
		err = cd.NewError(cd.Unexpected, "illegal remote object value")
		slog.Error("error occurred", "error", err.Error())
		return
	}

	localModel, localErr := local.GetEntityModel(localEntity, nil)
	if localErr != nil {
		err = localErr

		slog.Error("message")
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
				slog.Error("error occurred", "val", fieldValue.Name, "error", err.Error())
				return
			}
			continue
		}
		if models.IsSliceField(localField) {
			err = updateSliceStructField(fieldValue.Get(), localField)
			if err != nil {

				slog.Error("message")
				return
			}

			continue
		}
		if models.IsStructField(localField) {
			err = updateStructField(fieldValue.Get(), localField)
			if err != nil {

				slog.Error("message")
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
		slog.Error("error occurred", "val", localField.GetName(), "error", err.Error())
		return
	}

	localField.Reset()
	for _, objectValuePtr := range sliceObjectValuePtr.Values {
		elemType := localField.GetType().Elem()
		localSubVal, _ := elemType.Interface(nil)
		localSubModel, localSubErr := local.GetValueModel(localSubVal)
		if localSubErr != nil {
			err = localSubErr

			slog.Error("error occurred", "error", err.Error())
			return
		}
		err = updateLocalModel(objectValuePtr, localSubModel)
		if err != nil {

			slog.Error("error occurred", "error", err.Error())
			return
		}
		err = localField.AppendSliceValue(localSubModel.Interface(elemType.IsPtrType()))
		if err != nil {

			slog.Error("error occurred", "error", err.Error())
			return
		}
	}

	return
}

func updateStructField(val any, vField models.Field) (err *cd.Error) {
	if val == nil {
		return
	}

	objectValuePtr, objectValueOK := val.(*remote.ObjectValue)
	if !objectValueOK {
		err = cd.NewError(cd.Unexpected, "illegal object value")
		slog.Error("error occurred", "val", vField.GetName(), "error", err.Error())
		return
	}

	elemType := vField.GetType().Elem()
	localFileVal, _ := elemType.Interface(nil)
	localModelVal, localModelErr := local.GetValueModel(localFileVal)
	if localModelErr != nil {
		err = localModelErr

		slog.Error("error occurred", "error", err.Error())
		return

	}
	err = updateLocalModel(objectValuePtr, localModelVal)
	if err != nil {

		slog.Error("error occurred", "error", err.Error())
		return
	}

	vField.SetValue(localModelVal.Interface(elemType.IsPtrType()))
	return
}

func UpdateSliceEntity(remoteSliceValuePtr *remote.SliceObjectValue, localSliceValue any) (err *cd.Error) {
	if remoteSliceValuePtr == nil {
		err = cd.NewError(cd.Unexpected, "illegal remote slice object value")
		slog.Error("error occurred", "error", err.Error())
		return
	}
	if localSliceValue == nil {
		err = cd.NewError(cd.Unexpected, "illegal local entity value")

		slog.Error("message")
		return
	}

	localTypePtr, localTypeErr := local.NewType(reflect.TypeOf(localSliceValue))
	if localTypeErr != nil {
		err = localTypeErr

		slog.Error("message")
		return
	}
	if !models.IsSlice(localTypePtr) || !localTypePtr.IsPtrType() {
		err = cd.NewError(cd.Unexpected, "illegal local entity type")

		slog.Error("message")
		return
	}

	localSliceReflect := reflect.Indirect(reflect.ValueOf(localSliceValue))
	localValuePtr := local.NewValue(localSliceReflect)
	for _, val := range remoteSliceValuePtr.Values {
		elemType := localTypePtr.Elem()
		localItemVal, _ := elemType.Interface(nil)
		localItemModel, localItemErr := local.GetValueModel(localItemVal)
		if localItemErr != nil {
			err = localItemErr

			slog.Error("message")
			return
		}
		err = updateLocalModel(val, localItemModel)
		if err != nil {

			slog.Error("message")
			return
		}

		localRawVal := localItemModel.Interface(elemType.IsPtrType())
		err = localValuePtr.Append(reflect.ValueOf(localRawVal))
		if err != nil {

			slog.Error("message")
			return
		}
	}
	return
}
