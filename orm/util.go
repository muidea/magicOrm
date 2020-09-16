package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) isCommonType(fType model.Type) bool {
	if fType.Depend() == nil {
		return true
	}

	return util.IsBasicType(fType.Depend().GetValue())
}

func (s *Orm) getModelItems(modelInfo model.Model, builder builder.Builder) (ret []interface{}, err error) {
	var items []interface{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()
		if !s.isCommonType(fType) {
			continue
		}

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
