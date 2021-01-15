package helper

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type GetValueFunc func(interface{}) (model.Value, error)
type ElemDependValueFunc func(model.Value) ([]model.Value, error)

type Helper interface {
	Encode(vVal model.Value, tType model.Type) (ret string, err error)
	Decode(val string, tType model.Type) (ret model.Value, err error)
}

type impl struct {
	getValue        GetValueFunc
	elemDependValue ElemDependValueFunc
}

func New(getValue GetValueFunc, elemDependValue ElemDependValueFunc) Helper {
	return &impl{getValue: getValue, elemDependValue: elemDependValue}
}

func (s *impl) Encode(vVal model.Value, tType model.Type) (ret string, err error) {
	if !tType.IsBasic() {
		err = fmt.Errorf("illegal value type, type:%s", tType.GetName())
		return
	}

	switch tType.GetValue() {
	case util.TypeBooleanField:
		ret, err = s.encodeBoolValue(vVal)
	case util.TypeDateTimeField:
		ret, err = s.encodeDateTimeValue(vVal)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = s.encodeFloatValue(vVal)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = s.encodeIntValue(vVal)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = s.encodeUintValue(vVal)
	case util.TypeSliceField:
		ret, err = s.encodeSliceValue(vVal, tType)
	case util.TypeStringField:
		ret, err = s.encodeStringValue(vVal)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}

func (s *impl) Decode(val string, tType model.Type) (ret model.Value, err error) {
	if !tType.IsBasic() {
		err = fmt.Errorf("illegal value type, type:%s", tType.GetName())
		return
	}

	switch tType.GetValue() {
	case util.TypeBooleanField:
		ret, err = s.decodeBoolValue(val)
	case util.TypeDateTimeField:
		ret, err = s.decodeDateTimeValue(val)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = s.decodeFloatValue(val)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = s.decodeIntValue(val)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = s.decodeUintValue(val)
	case util.TypeSliceField:
		ret, err = s.decodeSliceValue(val, tType)
	case util.TypeStringField:
		ret, err = s.decodeStringValue(val)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}
