package orm

import (
	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

func shouldUseImplicitQueryCondition(field models.Field) bool {
	if field == nil || models.IsPrimaryField(field) || !models.IsAssignedField(field) {
		return false
	}

	// Slice fields are ambiguous in Query(model):
	// business code often assigns []/relation-slice to express the response shape
	// rather than an actual WHERE condition. Complex slice filtering should use
	// explicit Filter operators instead of implicit model-to-filter conversion.
	if models.IsSliceField(field) {
		return false
	}

	return true
}

func getModelFilter(vModel models.Model, provider provider.Provider, modelCodec codec.Codec) (ret models.Filter, err *cd.Error) {
	filterVal, filterErr := provider.GetModelFilter(vModel)
	if filterErr != nil {
		err = filterErr
		slog.Error("getModelFilter GetModelFilter failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	hasPKValue := false
	pkField := vModel.GetPrimaryField()
	if models.IsAssignedField(pkField) {
		pkVal, pkErr := modelCodec.PackedBasicFieldValue(pkField, pkField.GetValue())
		if pkErr != nil {
			err = pkErr
			slog.Error("getModelFilter PackedBasicFieldValue failed", "field", pkField.GetName(), "error", err.Error())
			return
		}

		err = filterVal.Equal(pkField.GetName(), pkVal)
		if err != nil {
			slog.Error("getModelFilter Equal failed", "field", pkField.GetName(), "error", err.Error())
			return
		}
		hasPKValue = true
	}

	if hasPKValue {
		ret = filterVal
		return
	}

	for _, field := range vModel.GetFields() {
		if !shouldUseImplicitQueryCondition(field) {
			continue
		}

		// 这里需要考虑普通值和Struct以及Slice Struct值的分开处理
		// basic, basic slice
		if models.IsBasicField(field) {
			//fieldVal, fieldErr := modelCodec.PackedBasicFieldValue(field, field.GetValue())
			//if fieldErr != nil {
			//	err = fieldErr
			//	slog.Error("operation failed", "error", "operation failed")
			//	return
			//}
			err = filterVal.Equal(field.GetName(), field.GetValue().Get())
			if err != nil {
				slog.Error("getModelFilter Equal failed", "field", field.GetName(), "error", err.Error())
				return
			}

			continue
		}

		// struct
		if models.IsStructField(field) {
			//fieldVal, fieldErr := modelCodec.PackedStructFieldValue(field, field.GetValue())
			//if fieldErr != nil {
			//	err = fieldErr
			//	slog.Error("operation failed", "error", "operation failed")
			//	return
			//}

			err = filterVal.Equal(field.GetName(), field.GetValue().Get())
			if err != nil {
				slog.Error("getModelFilter Equal struct failed", "field", field.GetName(), "error", err.Error())
				return
			}

			continue
		}

		// struct slice
		if models.IsSliceField(field) {
			//fieldVal, fieldErr := modelCodec.PackedSliceStructFieldValue(field, field.GetValue())
			//if fieldErr != nil {
			//	err = fieldErr
			//	slog.Error("operation failed", "error", "operation failed")
			//	return
			//}

			err = filterVal.In(field.GetName(), field.GetValue().Get())
			if err != nil {
				slog.Error("getModelFilter In failed", "field", field.GetName(), "error", err.Error())
				return
			}

			continue
		}
	}

	ret = filterVal
	return
}
