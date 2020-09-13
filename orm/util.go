package orm

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) getModelItems(modelInfo model.Model) (ret []interface{}, err error) {
	var items []interface{}
	fields := modelInfo.GetFields()
	for _, item := range fields {
		fType := item.GetType()
		if !s.isCommonType(fType) {
			continue
		}

		itemVal, itemErr := s.getInitValue(fType)
		if itemErr != nil {
			err = itemErr
			return
		}

		items = append(items, itemVal)
	}
	ret = items

	return
}

func (s *Orm) isCommonType(vType model.Type) bool {
	if vType.Depend() != nil {
		vType = vType.Depend()
	}

	return util.IsBasicType(vType.GetValue())
}

func (s *Orm) getInitValue(vType model.Type) (ret interface{}, err error) {
	switch vType.GetValue() {
	case util.TypeBooleanField,
		util.TypeBitField, util.TypeSmallIntegerField, util.TypeIntegerField, util.TypeInteger32Field, util.TypeBigIntegerField:
		val := int64(0)
		ret = &val
		break
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveIntegerField, util.TypePositiveInteger32Field, util.TypePositiveBigIntegerField:
		val := uint64(0)
		ret = &val
		break
	case util.TypeStringField, util.TypeDateTimeField:
		val := ""
		ret = &val
		break
	case util.TypeFloatField, util.TypeDoubleField:
		val := 0.00
		ret = &val
		break
	case util.TypeStructField:
		val := 0
		ret = &val
	case util.TypeSliceField:
		if util.IsBasicType(vType.Depend().GetValue()) {
			val := ""
			ret = &val
		} else {
			err = fmt.Errorf("no support fileType, name:%s, value:%d", vType.GetName(), vType.GetValue())
		}
	default:
		err = fmt.Errorf("no support fileType, name:%s, value:%d", vType.GetName(), vType.GetValue())
	}

	return
}
