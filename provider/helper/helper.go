package helper

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type GetValueFunc func(interface{}) (model.Value, error)
type GetValueModelFunc func(model.Value, model.Type) (model.Model, error)
type ElemDependValueFunc func(model.Value) ([]model.Value, error)

type Helper interface {
	Encode(vVal model.Value, tType model.Type) (ret string, err error)
	Decode(val string, tType model.Type) (ret model.Value, err error)
}

type impl struct {
	getValueModel   GetValueModelFunc
	getValue        GetValueFunc
	elemDependValue ElemDependValueFunc
}

func New(getValue GetValueFunc, getValueModel GetValueModelFunc, elemDependValue ElemDependValueFunc) Helper {
	return &impl{getValue: getValue, getValueModel: getValueModel, elemDependValue: elemDependValue}
}

func (s *impl) Encode(vVal model.Value, tType model.Type) (ret string, err error) {
	switch tType.GetValue() {
	case util.TypeBooleanField:
		ret, err = s.encodeBoolValue(vVal, tType)
	case util.TypeDateTimeField:
		ret, err = s.encodeDateTimeValue(vVal, tType)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = s.encodeFloatValue(vVal, tType)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = s.encodeIntValue(vVal, tType)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = s.encodeUintValue(vVal, tType)
	case util.TypeSliceField:
		ret, err = s.encodeSliceValue(vVal, tType)
	case util.TypeStringField:
		ret, err = s.encodeStringValue(vVal, tType)
	case util.TypeStructField:
		ret, err = s.encodeStructValue(vVal, tType)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}

func (s *impl) Decode(val string, tType model.Type) (ret model.Value, err error) {
	switch tType.GetValue() {
	case util.TypeBooleanField:
		ret, err = s.decodeBoolValue(val, tType)
	case util.TypeDateTimeField:
		ret, err = s.decodeDateTimeValue(val, tType)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = s.decodeFloatValue(val, tType)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = s.decodeIntValue(val, tType)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = s.decodeUintValue(val, tType)
	case util.TypeSliceField:
		ret, err = s.decodeSliceValue(val, tType)
	case util.TypeStringField:
		ret, err = s.decodeStringValue(val, tType)
	case util.TypeStructField:
		ret, err = s.decodeStructValue(val, tType)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}
