package orm

import (
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *orm) getItems(modelInfo model.Model) (ret []interface{}, err error) {
	items := []interface{}{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()
		depend := fType.Depend()
		if depend != nil && !util.IsBasicType(depend.GetValue()) {
			continue
		}

		itemVal, itemErr := util.GetBasicTypeInitValue(fType.GetValue())
		if itemErr != nil {
			err = itemErr
			return
		}

		items = append(items, itemVal)
	}
	ret = items

	return
}
