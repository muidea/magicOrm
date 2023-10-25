package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) getModelFilter(vModel model.Model) (ret model.Filter, re *cd.Result) {
	filterVal, filterErr := s.modelProvider.GetModelFilter(vModel)
	if filterErr != nil {
		re = cd.NewError(cd.UnExpected, filterErr.Error())
		log.Errorf("getModelFilter failed, s.modelProvider.GetEntityFilter error:%s", filterErr.Error())
		return
	}

	for _, val := range vModel.GetFields() {
		vType := val.GetType()
		vValue := val.GetValue()
		if vValue.IsZero() {
			continue
		}

		// if basic
		if model.IsBasicType(vType.GetValue()) {
			err := filterVal.Equal(val.GetName(), val.GetValue().Interface())
			if err != nil {
				re = cd.NewError(cd.UnExpected, err.Error())
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// if struct
		if model.IsStructType(vType.GetValue()) {
			err := filterVal.Equal(val.GetName(), val.GetValue().Interface())
			if err != nil {
				re = cd.NewError(cd.UnExpected, err.Error())
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// if slice
		err := filterVal.In(val.GetName(), val.GetValue().Interface())
		if err != nil {
			re = cd.NewError(cd.UnExpected, err.Error())
			log.Errorf("getModelFilter failed, filterVal.In error:%s", err.Error())
			return
		}
	}

	ret = filterVal
	return
}

func (s *impl) getModelFieldsScanDestPtr(vModel model.Model, builder builder.Builder) (ret []any, re *cd.Result) {
	items := []any{}
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		itemVal, itemErr := builder.GetFieldScanDest(field)
		if itemErr != nil {
			re = cd.NewError(cd.UnExpected, itemErr.Error())
			log.Errorf("getModelFieldsScanDestPtr failed, builder.GetFieldScanDest error:%s", itemErr.Error())
			return
		}

		items = append(items, itemVal)
	}
	ret = items

	return
}

func (s *impl) getModelPKFieldScanDestPtr(vModel model.Model, builder builder.Builder) (ret any, re *cd.Result) {
	itemVal, itemErr := builder.GetFieldScanDest(vModel.GetPrimaryField())
	if itemErr != nil {
		re = cd.NewError(cd.UnExpected, itemErr.Error())
		log.Errorf("getModelPKFieldScanDestPtr failed, builder.GetFieldScanDest error:%s", itemErr.Error())
		return
	}

	ret = itemVal
	return
}
