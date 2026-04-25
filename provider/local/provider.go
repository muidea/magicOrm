package local

import (
	"reflect"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

func checkEntityType(entity any) (ret *TypeImpl, err *cd.Error) {
	entityType := reflect.TypeOf(entity)
	typeImplPtr, typeImplErr := NewType(entityType)
	if typeImplErr != nil {
		err = typeImplErr
		slog.Error("checkEntityType NewType failed", "entityType", entityType.String(), "error", err.Error())
		return
	}
	if !models.IsStruct(typeImplPtr.Elem()) {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		slog.Error("checkEntityType: not a struct type", "entityType", entityType.String())
		return
	}

	ret = typeImplPtr
	return
}

func GetEntityType(entity any) (ret models.Type, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	typeImplPtr, typeImplErr := checkEntityType(entity)
	if typeImplErr != nil {
		err = typeImplErr
		slog.Error("GetEntityType checkEntityType failed", "error", err.Error())
		return
	}

	ret = typeImplPtr
	return
}

func GetEntityValue(entity any) (ret models.Value, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	_, err = checkEntityType(entity)
	if err != nil {
		slog.Error("GetEntityValue checkEntityType failed", "error", err.Error())
		return
	}

	ret = NewValue(reflect.ValueOf(entity))
	return
}

func GetEntityModel(entity any, valueValidator models.ValueValidator) (ret models.Model, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	reallyVal := reflect.Indirect(reflect.ValueOf(entity))
	if !reallyVal.CanSet() {
		newVal := reflect.New(reallyVal.Type()).Elem()
		newVal.Set(reallyVal)
		reallyVal = newVal
	}

	implPtr, implErr := getValueModel(reallyVal, models.OriginView)
	if implErr != nil {
		err = implErr
		slog.Error("GetEntityModel getValueModel failed", "entityType", reallyVal.Type().String(), "error", err.Error())
		return
	}

	implPtr.valueValidator = valueValidator
	ret = implPtr
	return
}

func GetValueModel(vVal models.Value) (ret models.Model, err *cd.Error) {
	if vVal == nil {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		return
	}
	valueImplPtr, valueImplOK := vVal.(*ValueImpl)
	if !valueImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		return
	}

	implPtr, implErr := getValueModel(valueImplPtr.value, models.MetaView)
	if implErr != nil {
		err = implErr
		slog.Error("GetValueModel getValueModel failed", "error", err.Error())
		return
	}

	ret = implPtr
	return
}

func GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true)))
	ret = newFilter(valuePtr, vModel)
	return
}

func SetModelValue(vModel models.Model, vVal models.Value, disableValidator bool) (ret models.Model, err *cd.Error) {
	valImplPtr, valImplOK := vVal.(*ValueImpl)
	if !valImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		slog.Error("SetModelValue: value is not *ValueImpl")
		return
	}
	valueModel, valueModelErr := getValueModel(valImplPtr.value, models.OriginView)
	if valueModelErr != nil {
		err = valueModelErr
		slog.Error("SetModelValue getValueModel failed", "error", err.Error())
		return
	}

	vModelImplPtr := vModel.(*objectImpl)
	fields := valueModel.GetFields()
	for _, field := range fields {
		if !models.IsValidField(field) && !models.IsAssignedField(field) {
			continue
		}

		err = vModelImplPtr.innerSetFieldValue(field.GetName(), field.GetValue().Get(), disableValidator)
		if err != nil {
			slog.Error("SetModelValue innerSetFieldValue failed", "model", vModel.GetPkgKey(), "field", field.GetName(), "error", err.Error())
			return
		}
	}

	ret = vModel
	return
}
