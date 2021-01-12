package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) getModelFilter(vModel model.Model) (ret model.Filter, err error) {
	filter := &queryFilter{params: map[string]model.FilterItem{}, modelProvider: s.modelProvider}

	fields := vModel.GetFields()
	for _, item := range fields {
		vVal := item.GetValue()
		vType := item.GetType()
		if !s.modelProvider.IsAssigned(vVal, vType) {
			continue
		}

		filter.equalInternal(item.GetName(), vVal)
	}

	ret = filter
	return
}

func (s *Orm) getFieldFilter(vField model.Field) (ret model.Filter, err error) {
	filter := &queryFilter{params: map[string]model.FilterItem{}, modelProvider: s.modelProvider}
	vVal := vField.GetValue()
	if !vVal.IsNil() {
		filter.equalInternal(vField.GetName(), vField.GetValue())
	}

	ret = filter
	return
}

func (s *Orm) getModelItems(modelInfo model.Model, builder builder.Builder) (ret []interface{}, err error) {
	var items []interface{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()
		if item.GetValue().IsNil() || !fType.IsBasic() {
			continue
		}

		itemVal, itemErr := builder.DeclareFieldValue(item)
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

func (s *Orm) stripSlashes(fType model.Type, val interface{}) interface{} {
	if !s.needStripSlashes(fType) {
		return val
	}

	strPtr, strOK := val.(*string)
	if !strOK {
		return val
	}

	strVal := util.StripSlashes(*strPtr)
	return &strVal
}
