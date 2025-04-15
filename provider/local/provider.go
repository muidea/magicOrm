package local

import (
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
)

func checkEntityType(entity any) (ret *TypeImpl, err *cd.Error) {
	typeImplPtr, typeImplErr := NewType(reflect.TypeOf(entity))
	if typeImplErr != nil {
		err = typeImplErr
		log.Errorf("checkEntityType failed, illegal entity type, err:%s", err.Error())
		return
	}
	if !model.IsStruct(typeImplPtr.Elem()) {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		log.Errorf("checkEntityType failed, illegal entity type, err:%s", err.Error())
		return
	}

	ret = typeImplPtr
	return
}

func GetEntityType(entity any) (ret model.Type, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	typeImplPtr, typeImplErr := checkEntityType(entity)
	if typeImplErr != nil {
		err = typeImplErr
		log.Errorf("GetEntityType failed, err:%s", err.Error())
		return
	}

	ret = typeImplPtr
	return
}

func GetEntityValue(entity any) (ret model.Value, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		return
	}

	_, err = checkEntityType(entity)
	if err != nil {
		log.Errorf("GetEntityValue failed, err:%s", err.Error())
		return
	}

	ret = NewValue(reflect.ValueOf(entity))
	return
}

func GetEntityModel(entity any) (ret model.Model, err *cd.Error) {
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

	implPtr, implErr := getValueModel(reallyVal, model.OriginView)
	if implErr != nil {
		err = implErr
		return
	}

	ret = implPtr
	return
}

func GetValueModel(vVal model.Value) (ret model.Model, err *cd.Error) {
	if vVal == nil {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		return
	}
	valueImplPtr, valueImplOK := vVal.(*ValueImpl)
	if !valueImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		return
	}

	implPtr, implErr := getValueModel(valueImplPtr.value, model.MetaView)
	if implErr != nil {
		err = implErr
		return
	}

	ret = implPtr
	return
}

func GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Error) {
	valuePtr := NewValue(reflect.ValueOf(vModel.Interface(true)))
	ret = newFilter(valuePtr)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err *cd.Error) {
	valImplPtr, valImplOK := vVal.(*ValueImpl)
	if !valImplOK {
		err = cd.NewError(cd.IllegalParam, "value is invalid")
		log.Errorf("SetModelValue failed, err:%s", err.Error())
		return
	}
	valueModel, valueModelErr := getValueModel(valImplPtr.value, model.OriginView)
	if valueModelErr != nil {
		err = valueModelErr
		log.Errorf("SetModelValue failed, err:%s", err.Error())
		return
	}

	fields := valueModel.GetFields()
	for _, field := range fields {
		if !model.IsValidField(field) {
			continue
		}

		err = vModel.SetFieldValue(field.GetName(), field.GetValue().Get())
		if err != nil {
			log.Errorf("SetModelValue failed, set field:%s value err:%s", field.GetName(), err.Error())
			return
		}
	}

	ret = vModel
	return
}
