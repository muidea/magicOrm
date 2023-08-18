package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *impl) getModelFilter(vModel model.Model) (ret model.Filter, err error) {
	filterVal, filterErr := s.modelProvider.GetEntityFilter(vModel.Interface(true))
	if filterErr != nil {
		err = filterErr
		return
	}

	for _, val := range vModel.GetFields() {
		vType := val.GetType()
		vValue := val.GetValue()
		//if vValue.IsZero() {
		//	continue
		//}

		if !s.modelProvider.IsAssigned(vValue, vType) {
			continue
		}

		// if basic
		if model.IsBasicType(vType.Elem().GetValue()) {
			err = filterVal.Equal(val.GetName(), val.GetValue().Interface())
			if err != nil {
				return
			}

			continue
		}

		// if struct
		if model.IsStructType(vType.GetValue()) {
			err = filterVal.Equal(val.GetName(), val.GetValue().Interface())
			if err != nil {
				return
			}

			continue
		}

		// if struct slice
		err = filterVal.In(val.GetName(), val.GetValue().Interface())
		if err != nil {
			return
		}
	}

	ret = filterVal
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
	case model.TypeStringValue, model.TypeDateTimeValue:
		return true
	}

	if !model.IsSliceType(fType.GetValue()) {
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
