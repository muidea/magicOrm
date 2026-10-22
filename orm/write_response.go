package orm

import (
	"reflect"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type responseValuePointer struct {
	typeOf reflect.Type
	value  uintptr
	length int
}

// 在 Copy 之前检查原值图；复制后的对象身份不能用来识别输入指针环。
func validateWriteResponseValue(value reflect.Value, active map[responseValuePointer]bool, depth int) *cd.Error {
	if !value.IsValid() {
		return nil
	}
	if depth > 256 {
		return cd.NewError(cd.IllegalParam, "write response value exceeds nesting limit")
	}
	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			return validateWriteResponseValue(value.Elem(), active, depth+1)
		}
	case reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			return nil
		}
		key := responseValuePointer{typeOf: value.Type(), value: value.Pointer()}
		if value.Kind() == reflect.Slice {
			key.length = value.Len()
		}
		if active[key] {
			return cd.NewError(cd.IllegalParam, "cyclic write response value is not supported")
		}
		active[key] = true
		defer delete(active, key)
		if value.Kind() == reflect.Ptr {
			return validateWriteResponseValue(value.Elem(), active, depth+1)
		}
		for idx := 0; idx < value.Len(); idx++ {
			if err := validateWriteResponseValue(value.Index(idx), active, depth+1); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for idx := 0; idx < value.NumField(); idx++ {
			field := value.Type().Field(idx)
			// local Copy 也复制导出的 orm:"-" 内存字段，必须在这里检查其值图。
			if !field.IsExported() {
				continue
			}
			if err := validateWriteResponseValue(value.Field(idx), active, depth+1); err != nil {
				return err
			}
		}
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			if err := validateWriteResponseValue(iter.Key(), active, depth+1); err != nil {
				return err
			}
			if err := validateWriteResponseValue(iter.Value(), active, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildWriteResponseModel(vModel models.Model) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "write response model is nil")
		return
	}

	if err = validateWriteResponseValue(reflect.ValueOf(vModel.Interface(true)), map[responseValuePointer]bool{}, 0); err != nil {
		return
	}
	ret = vModel.Copy(models.DetailView)
	return
}

func projectWriteResponseModel(vModel models.Model, modelProvider provider.Provider) (ret models.Model, err *cd.Error) {
	responseModel, responseErr := buildWriteResponseModel(vModel)
	if responseErr != nil {
		err = responseErr
		return
	}

	ret, err = projectModelResponse(vModel, responseModel, false, modelProvider)
	return
}

func projectModelResponse(vModel, responseModel models.Model, responseByMask bool, modelProvider provider.Provider) (ret models.Model, err *cd.Error) {
	if vModel == nil || responseModel == nil {
		ret = vModel
		return
	}

	projectedModel := responseModel.Copy(models.OriginView)
	primaryField := vModel.GetPrimaryField()
	if primaryField != nil && (models.IsValidField(primaryField) || models.IsAssignedField(primaryField)) {
		assignProjectedFieldValue(projectedModel.GetPrimaryField(), primaryField.GetValue().Get())
	}

	for _, field := range projectedModel.GetFields() {
		if !fieldIncludedInResponse(responseModel, field, responseByMask) {
			continue
		}
		if models.IsPrimaryField(field) {
			continue
		}

		sourceField := vModel.GetField(field.GetName())
		if sourceField == nil {
			continue
		}
		if !models.IsValidField(sourceField) && !models.IsAssignedField(sourceField) {
			assignProjectedFieldValue(field, nil)
			continue
		}

		if err = projectResponseFieldValue(field, sourceField, modelProvider); err != nil {
			return
		}
	}

	ret = projectedModel
	return
}

func projectResponseFieldValue(dstField, srcField models.Field, modelProvider provider.Provider) (err *cd.Error) {
	if dstField == nil || srcField == nil {
		return
	}

	if models.IsBasicField(srcField) {
		assignProjectedFieldValue(dstField, srcField.GetValue().Get())
		return
	}

	if !models.IsValidField(srcField) {
		assignProjectedFieldValue(dstField, nil)
		return
	}

	if models.IsSliceField(srcField) {
		err = projectResponseSliceRelationValue(dstField, srcField, modelProvider)
		return
	}

	var projectedVal any
	projectedVal, err = projectRelationValue(srcField.GetType(), srcField.GetValue().Get(), modelProvider)
	if err != nil {
		return
	}

	assignProjectedFieldValue(dstField, projectedVal)
	return
}

func projectResponseSliceRelationValue(dstField, srcField models.Field, modelProvider provider.Provider) (err *cd.Error) {
	if dstField == nil || srcField == nil {
		return
	}

	if srcField.GetValue().Get() == nil {
		assignProjectedFieldValue(dstField, nil)
		return
	}

	dstField.Reset()
	elemType := srcField.GetType().Elem()
	for _, itemVal := range srcField.GetSliceValue() {
		if itemVal == nil || !itemVal.IsValid() {
			continue
		}

		var projectedVal any
		projectedVal, err = projectRelationValue(elemType, itemVal.Get(), modelProvider)
		if err != nil {
			return
		}
		if projectedVal == nil {
			continue
		}

		if err = dstField.AppendSliceValue(projectedVal); err != nil {
			return
		}
	}

	return
}

func projectRelationValue(relationType models.Type, rawVal any, modelProvider provider.Provider) (ret any, err *cd.Error) {
	if relationType == nil || rawVal == nil {
		return
	}

	relationModel, modelErr := modelProvider.GetTypeModel(relationType)
	if modelErr != nil {
		err = modelErr
		return
	}

	relationValue, valueErr := modelProvider.GetEntityValue(rawVal)
	if valueErr != nil {
		err = valueErr
		return
	}

	relationModel, err = modelProvider.SetModelValue(relationModel, relationValue)
	if err != nil {
		return
	}

	responseModel := relationModel.Copy(models.LiteView)
	projectedModel, projectErr := projectModelResponse(relationModel, responseModel, false, modelProvider)
	if projectErr != nil {
		err = projectErr
		return
	}

	ret = projectedModel.Interface(relationType.IsPtrType())
	return
}
