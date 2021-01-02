package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) getModelItems(modelInfo model.Model, builder builder.Builder) (ret []interface{}, err error) {
	var items []interface{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()
		if !fType.IsBasic() {
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

func (s *Orm) needStripSlashes(fType model.Type) bool {
	switch fType.GetValue() {
	case util.TypeStringField, util.TypeDateTimeField:
		return true
	}

	if !util.IsSliceType(fType.GetValue()) {
		return false
	}

	return fType.IsBasic()
}
