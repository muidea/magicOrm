package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

func getModelFilter(vModel model.Model, provider provider.Provider, modelCodec codec.Codec) (ret model.Filter, err *cd.Result) {
	filterVal, filterErr := provider.GetModelFilter(vModel)
	if filterErr != nil {
		err = filterErr
		log.Errorf("getModelFilter failed, s.modelProvider.GetEntityFilter error:%s", err.Error())
		return
	}

	hasPKValue := false
	pkField := vModel.GetPrimaryField()
	if model.IsAssignedField(pkField) {
		pkVal, pkErr := modelCodec.PackedBasicFieldValue(pkField, pkField.GetValue())
		if pkErr != nil {
			err = pkErr
			log.Errorf("getModelFilter failed, modelCodec.PackedFieldValue error:%s", err.Error())
			return
		}

		err = filterVal.Equal(pkField.GetName(), pkVal)
		if err != nil {
			log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
			return
		}
		hasPKValue = true
	}

	if hasPKValue {
		ret = filterVal
		return
	}

	for _, field := range vModel.GetFields() {
		if model.IsPrimaryField(field) || !model.IsAssignedField(field) {
			continue
		}

		// 这里需要考虑普通值和Struct以及Slice Struct值的分开处理
		// basic, basic slice
		if model.IsBasicField(field) {
			//fieldVal, fieldErr := modelCodec.PackedBasicFieldValue(field, field.GetValue())
			//if fieldErr != nil {
			//	err = fieldErr
			//	log.Errorf("getModelFilter failed, modelCodec.PackedFieldValue error:%s", err.Error())
			//	return
			//}
			err = filterVal.Equal(field.GetName(), field.GetValue().Get())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// struct
		if model.IsStructField(field) {
			//fieldVal, fieldErr := modelCodec.PackedStructFieldValue(field, field.GetValue())
			//if fieldErr != nil {
			//	err = fieldErr
			//	log.Errorf("getModelFilter failed, modelCodec.PackedFieldValue error:%s", err.Error())
			//	return
			//}

			err = filterVal.Equal(field.GetName(), field.GetValue().Get())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// struct slice
		if model.IsSliceField(field) {
			//fieldVal, fieldErr := modelCodec.PackedSliceStructFieldValue(field, field.GetValue())
			//if fieldErr != nil {
			//	err = fieldErr
			//	log.Errorf("getModelFilter failed, modelCodec.PackedFieldValue error:%s", err.Error())
			//	return
			//}

			err = filterVal.In(field.GetName(), field.GetValue().Get())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.In error:%s", err.Error())
				return
			}

			continue
		}
	}

	ret = filterVal
	return
}
