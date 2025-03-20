package remote

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
)

func GetEntityType(entity any) (ret model.Type, err *cd.Result) {
	if entity == nil {
		err = cd.NewResult(cd.UnExpected, "entity is nil")
		return
	}

	switch val := entity.(type) {
	case *Object:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   model.TypeStructValue,
			IsPtr:   true,
		}
	case *ObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   model.TypeStructValue,
			IsPtr:   true,
		}
	case *SliceObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   model.TypeSliceValue,
			IsPtr:   true,
			ElemType: &TypeImpl{
				Name:    val.Name,
				PkgPath: val.PkgPath,
				Value:   model.TypeStructValue,
				IsPtr:   true,
			},
		}
	case Object:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   model.TypeStructValue,
			IsPtr:   true,
		}
	case ObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   model.TypeStructValue,
			IsPtr:   true,
		}
	case SliceObjectValue:
		ret = &TypeImpl{
			Name:    val.Name,
			PkgPath: val.PkgPath,
			Value:   model.TypeSliceValue,
			IsPtr:   true,
			ElemType: &TypeImpl{
				Name:    val.Name,
				PkgPath: val.PkgPath,
				Value:   model.TypeStructValue,
				IsPtr:   true,
			},
		}
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal entity, entity:%v", entity))
		return
	}
	return
}

func GetEntityValue(entity any) (ret model.Value, err *cd.Result) {
	if entity == nil {
		err = cd.NewResult(cd.UnExpected, "entity is nil")
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
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal entity, entity:%v", entity))
		//log.Errorf("GetEntityValue failed, err:%s", err.Error())
		return
	}

	return
}

func GetEntityModel(entity any) (ret model.Model, err *cd.Result) {
	if entity == nil {
		err = cd.NewResult(cd.UnExpected, "entity is nil")
		return
	}

	switch val := entity.(type) {
	case *Object:
		ret = val
	case Object:
		ret = &val
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal entity, entity:%v", entity))
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
	}

	return
}

func GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Result) {
	objectPtr, objectOK := vModel.(*Object)
	if !objectOK {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal model, model:%v", vModel))
		log.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	ret = NewFilter(objectPtr)
	return
}

func SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err *cd.Result) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("SetModelValue failed, illegal value, err:%v", errInfo))
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
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal model value, val:%v", val))
		}
		if err != nil {
			log.Errorf("SetModelValue failed, err:%s", err.Error())
			return
		}
	}

	ret = vModel
	return
}

func assignObjectValue(vModel model.Model, objectValuePtr *ObjectValue) (err *cd.Result) {
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
