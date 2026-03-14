package remote

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

func GetEntityType(entity any) (ret models.Type, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
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
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal entity, entity:%v", entity))
		return
	}
	return
}

func GetEntityValue(entity any) (ret models.Value, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
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
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal entity, entity:%v", entity))
		return
	}

	return
}

func GetEntityModel(entity any, valueValidator models.ValueValidator) (ret models.Model, err *cd.Error) {
	if entity == nil {
		err = cd.NewError(cd.IllegalParam, "entity is nil")
		slog.Error("GetEntityModel: entity is nil")
		return
	}

	var objectPtr *Object
	switch val := entity.(type) {
	case *Object:
		objectPtr = val
	case Object:
		objectPtr = &val
	default:
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal entity, entity:%v", entity))
		slog.Error("GetEntityModel: illegal entity type", "error", err.Error())
	}

	if err != nil {
		return
	}

	objectPtr.valueValidator = valueValidator
	ret = objectPtr
	return
}

func GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "vModel is nil")
		slog.Error("GetModelFilter: vModel is nil")
		return
	}
	objectPtr, objectOK := vModel.(*Object)
	if !objectOK {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal model, model:%v", vModel))
		slog.Error("GetModelFilter: model is not *Object", "error", err.Error())
		return
	}

	ret = NewFilter(objectPtr)
	return
}

func SetModelValue(vModel models.Model, vVal models.Value, disableValidator bool) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "vModel is nil")
		slog.Error("SetModelValue: vModel is nil")
		return
	}
	if vVal == nil {
		err = cd.NewError(cd.IllegalParam, "vVal is nil")
		slog.Error("SetModelValue: vVal is nil")
		return
	}
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("SetModelValue failed, illegal value, err:%v", errInfo))
			slog.Error("SetModelValue panic recovered", "error", err.Error())
			return
		}
	}()

	vObjectPtr := vModel.(*Object)
	switch val := vVal.Get().(type) {
	case *ObjectValue:
		err = assignObjectValue(vObjectPtr, val, disableValidator)
	default:
		if vVal.IsValid() {
			err = vObjectPtr.innerSetPrimaryFieldValue(val, disableValidator)
		} else {
			err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal model value, val:%v", val))
		}
		if err != nil {
			slog.Error("SetModelValue innerSetPrimaryFieldValue failed", "error", err.Error())
			return
		}
	}

	ret = vModel
	return
}

func assignObjectValue(vObjectPtr *Object, objectValuePtr *ObjectValue, disableValidator bool) (err *cd.Error) {
	for idx := range objectValuePtr.Fields {
		fieldVal := objectValuePtr.Fields[idx]
		if !fieldVal.Assigned && fieldVal.GetValue().IsZero() {
			continue
		}
		err = vObjectPtr.innerSetFieldValue(fieldVal.GetName(), fieldVal.Get(), disableValidator)
		if err != nil {
			slog.Error("assignObjectValue innerSetFieldValue failed", "field", fieldVal.GetName(), "error", err.Error())
			return
		}
	}

	return
}
