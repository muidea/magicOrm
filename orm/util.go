package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

func (s *impl) getModelFilter(vModel model.Model, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Result) {
	filterVal, filterErr := s.modelProvider.GetModelFilter(vModel, viewSpec)
	if filterErr != nil {
		err = filterErr
		log.Errorf("getModelFilter failed, s.modelProvider.GetEntityFilter error:%s", err.Error())
		return
	}

	for _, val := range vModel.GetFields() {
		fType := val.GetType()
		fValue := val.GetValue()
		if fValue.IsZero() {
			continue
		}

		// if basic
		if model.IsBasicType(fType.GetValue()) {
			err = filterVal.Equal(val.GetName(), val.GetValue().Interface())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// if struct
		if model.IsStructType(fType.GetValue()) {
			// 为了避免自己引用或关联自己
			if fType.GetPkgKey() == vModel.GetPkgKey() {
				vValue := vModel.GetPrimaryField().GetValue()
				if util.IsSameValue(fValue.Interface(), vValue.Interface()) {
					continue
				}
			}

			err = filterVal.Equal(val.GetName(), fValue.Interface())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// if slice
		err = filterVal.In(val.GetName(), fValue.Interface())
		if err != nil {
			log.Errorf("getModelFilter failed, filterVal.In error:%s", err.Error())
			return
		}
	}

	ret = filterVal
	return
}

func (s *impl) getModelFieldsScanDestPtr(vModel model.Model, builder builder.Builder) (ret []any, err *cd.Result) {
	items := []any{}
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || !fValue.IsValid() {
			continue
		}

		itemVal, itemErr := builder.GetFieldScanDest(field)
		if itemErr != nil {
			err = itemErr
			log.Errorf("getModelFieldsScanDestPtr failed, builder.GetFieldScanDest error:%s", err.Error())
			return
		}

		items = append(items, itemVal)
	}
	ret = items

	return
}

func (s *impl) getModelPKFieldScanDestPtr(vModel model.Model, builder builder.Builder) (ret any, err *cd.Result) {
	itemVal, itemErr := builder.GetFieldScanDest(vModel.GetPrimaryField())
	if itemErr != nil {
		err = itemErr
		log.Errorf("getModelPKFieldScanDestPtr failed, builder.GetFieldScanDest error:%s", err.Error())
		return
	}

	ret = itemVal
	return
}
