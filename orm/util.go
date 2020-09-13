package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) getModelItems(modelInfo model.Model, builder builder.Builder) (ret []interface{}, err error) {
	var items []interface{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		itemVal, itemErr := builder.DeclareFieldValue(item)
		if itemVal == nil {
			continue
		}

		if itemErr != nil {
			err = itemErr
			return
		}

		items = append(items, itemVal)
	}
	ret = items

	return
}
