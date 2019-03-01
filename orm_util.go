package orm

import (
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) getItems(modelInfo model.Model) (ret []interface{}, err error) {
	items := []interface{}{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()
		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			return
		}
		if dependModel != nil {
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
