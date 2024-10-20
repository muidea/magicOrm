package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/util"
)

func getModelFilter(vModel model.Model, provider provider.Provider, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Result) {
	filterVal, filterErr := provider.GetModelFilter(vModel, viewSpec)
	if filterErr != nil {
		err = filterErr
		log.Errorf("getModelFilter failed, s.modelProvider.GetEntityFilter error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if fValue.IsZero() {
			continue
		}

		// if basic
		if field.IsBasic() {
			err = filterVal.Equal(field.GetName(), field.GetValue().Interface().Value())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}

		// if slice
		if field.IsSlice() {
			err = filterVal.In(field.GetName(), fValue.Interface().Value())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.In error:%s", err.Error())
				return
			}

			continue
		}

		// if struct
		if field.IsStruct() {
			// 为了避免自己引用或关联自己
			if fType.GetPkgKey() == vModel.GetPkgKey() {
				vValue := vModel.GetPrimaryField().GetValue()
				if util.IsSameValue(fValue.Interface(), vValue.Interface()) {
					continue
				}
			}

			err = filterVal.Equal(field.GetName(), fValue.Interface().Value())
			if err != nil {
				log.Errorf("getModelFilter failed, filterVal.Equal error:%s", err.Error())
				return
			}

			continue
		}
	}

	ret = filterVal
	return
}
