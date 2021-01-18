package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *impl) getModelFilter(vModel model.Model) (ret model.Filter, err error) {
	filter := &queryFilter{params: map[string]model.FilterItem{}, modelProvider: s.modelProvider}

	for _, field := range vModel.GetFields() {
		vVal := field.GetValue()
		vType := field.GetType()
		if !s.modelProvider.IsAssigned(vVal, vType) {
			continue
		}

		filter.equalInternal(field.GetName(), vVal)
	}

	ret = filter
	return
}

func (s *impl) getFieldFilter(vField model.Field) (ret model.Filter, err error) {
	filter := &queryFilter{params: map[string]model.FilterItem{}, modelProvider: s.modelProvider}
	vVal := vField.GetValue()
	if !vVal.IsNil() {
		filter.equalInternal(vField.GetName(), vField.GetValue())
	}

	ret = filter
	return
}

func (s *impl) getInitializeValue(vModel model.Model, builder builder.Builder) (ret []interface{}, err error) {
	var items []interface{}
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		if !fType.IsBasic() {
			continue
		}

		itemVal, itemErr := builder.GetInitializeValue(field)
		if itemErr != nil {
			err = itemErr
			return
		}

		items = append(items, itemVal)
	}
	ret = items

	return
}

func (s *impl) needStripSlashes(fType model.Type) bool {
	switch fType.GetValue() {
	case util.TypeStringField, util.TypeDateTimeField:
		return true
	}

	if !util.IsSliceType(fType.GetValue()) {
		return false
	}

	return fType.IsBasic()
}

func (s *impl) stripSlashes(fType model.Type, val interface{}) interface{} {
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
