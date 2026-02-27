package local

import (
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

func checkEntityType(entity any) (ret *TypeImpl, err *cd.Error) {
	typeImplPtr, typeImplErr := NewType(reflect.TypeOf(entity))
	if typeImplErr != nil {
		err = typeImplErr
		slog.Error("error occurred", "error", "operation failed")
		return
	}
	if !models.IsStruct(typeImplPtr.Elem()) {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		slog.Error("error occurred", "error", "operation failed")
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
		slog.Error("error occurred", "error", "operation failed")
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
		slog.Error("error occurred", "error", "operation failed")
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
		return
	}

	ret = implPtr
	return
}

func GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true)))
	ret = newFilter(valuePtr)
	return
}

func SetModelValue(vModel models.Model, vVal models.Value, disableValidator bool) (ret models.Model, err *cd.Error) {
	valImplPtr, valImplOK := vVal.(*ValueImpl)
	if !valImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		slog.Error("error occurred", "error", "operation failed")
		return
	}
	valueModel, valueModelErr := getValueModel(valImplPtr.value, models.OriginView)
	if valueModelErr != nil {
		err = valueModelErr
		slog.Error("error occurred", "error", "operation failed")
		return
	}

	vModelImplPtr := vModel.(*objectImpl)
	fields := valueModel.GetFields()
	for _, field := range fields {
		if !models.IsValidField(field) {
			continue
		}

		err = vModelImplPtr.innerSetFieldValue(field.GetName(), field.GetValue().Get(), disableValidator)
		if err != nil {
			slog.Error("error occurred", "error", err.Error())
			return
		}
	}

	ret = vModel
	return
}
