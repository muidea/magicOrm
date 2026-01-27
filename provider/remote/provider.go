package remote

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
)

func GetEntityType(entity any) (ret models.Type, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.Unexpected, "entity is nil")
		return
	}

	switch val := entity.(type) {
	case *Object:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   models.TypeStructValue,
			IsPtr:   true,
		}
	case *ObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   models.TypeStructValue,
			IsPtr:   true,
		}
	case *SliceObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   models.TypeSliceValue,
			IsPtr:   true,
			ElemType: &TypeImpl{
				Name:    val.Name,
				PkgPath: val.PkgPath,
				Value:   models.TypeStructValue,
				IsPtr:   true,
			},
		}
	case Object:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   models.TypeStructValue,
			IsPtr:   true,
		}
	case ObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   models.TypeStructValue,
			IsPtr:   true,
		}
	case SliceObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   models.TypeSliceValue,
			IsPtr:   true,
			ElemType: &TypeImpl{
				Name:    val.Name,
				PkgPath: val.PkgPath,
				Value:   models.TypeStructValue,
				IsPtr:   true,
			},
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal entity, entity:%v", entity))
		return
	}
	return
}

func GetEntityValue(entity any) (ret models.Value, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.Unexpected, "entity is nil")
		return
	}

	switch val := entity.(type) {
	case *ObjectValue:
		ret = &ValueImpl{
			value: val,
		}
	case *SliceObjectValue:
		ret = &ValueImpl{
			value: val,
		}
	case ObjectValue:
		ret = &ValueImpl{
			value: &val,
		}
	case SliceObjectValue:
		ret = &ValueImpl{
			value: &val,
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal entity, entity:%v", entity))
		//log.Errorf("GetEntityValue failed, err:%s", err.Error())
		return
	}

	return
}

func GetEntityModel(entity any, valueValidator models.ValueValidator) (ret models.Model, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.Unexpected, "entity is nil")
		return
	}

	var objectPtr *Object
	switch val := entity.(type) {
	case *Object:
		objectPtr = val
	case Object:
		objectPtr = &val
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal entity, entity:%v", entity))
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
	}

	if err != nil {
		return
	}

	objectPtr.valueValidator = valueValidator
	ret = objectPtr
	return
}

func GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error) {
	objectPtr, objectOK := vModel.(*Object)
	if !objectOK {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal model, model:%v", vModel))
		log.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	ret = NewFilter(objectPtr)
	return
}

func SetModelValue(vModel models.Model, vVal models.Value) (ret models.Model, err *cd.Error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("SetModelValue failed, illegal value, err:%v", errInfo))
			log.Errorf("SetModelValue failed, err:%s", err.Error())
			return
		}
	}()

	switch val := vVal.Get().(type) {
	case *ObjectValue:
		err = assignObjectValue(vModel, val)
	default:
		if vVal.IsValid() {
			err = vModel.SetPrimaryFieldValue(val)
		} else {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal model value, val:%v", val))
		}
		if err != nil {
			log.Errorf("SetModelValue failed, err:%s", err.Error())
			return
		}
	}

	ret = vModel
	return
}

func assignObjectValue(vModel models.Model, objectValuePtr *ObjectValue) (err *cd.Error) {
	for idx := range objectValuePtr.Fields {
		fieldVal := objectValuePtr.Fields[idx]
		err = vModel.SetFieldValue(fieldVal.GetName(), fieldVal.Get())
		if err != nil {
			log.Errorf("assignObjectValue failed, err:%s", err.Error())
			return
		}
	}

	return
}
