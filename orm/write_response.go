package orm

import (
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

func buildWriteResponseModel(vModel models.Model) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "write response model is nil")
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

	if srcField.GetValue().Get() == nil {
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
