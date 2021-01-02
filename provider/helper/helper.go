package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type GetValueFunc func(reflect.Value) model.Value
type GetValueModelFunc func(model.Value, model.Type) (model.Model, error)

type Helper interface {
	Encode(vVal model.Value, tType model.Type) (ret string, err error)
	Decode(val string, tType model.Type) (ret model.Value, err error)
}

type impl struct {
	getValueModel GetValueModelFunc
	getValue      GetValueFunc
}

func New(getValue GetValueFunc, getValueModel GetValueModelFunc) Helper {
	return &impl{getValue: getValue, getValueModel: getValueModel}
}

func (s *impl) Encode(vVal model.Value, tType model.Type) (ret string, err error) {
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
	case util.TypeStructField:
		ret, err = s.encodeStructValue(vVal, tType)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}

func (s *impl) Decode(val string, tType model.Type) (ret model.Value, err error) {
	var tVal model.Value
	switch tType.GetValue() {
	case util.TypeBooleanField:
		tVal, err = s.decodeBoolValue(val)
	case util.TypeDateTimeField:
		tVal, err = s.decodeDateTimeValue(val)
	case util.TypeFloatField, util.TypeDoubleField:
		tVal, err = s.decodeFloatValue(val)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		tVal, err = s.decodeIntValue(val)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		tVal, err = s.decodeUintValue(val)
	case util.TypeSliceField:
		tVal, err = s.decodeSliceValue(val, tType)
	case util.TypeStringField:
		tVal, err = s.decodeStringValue(val)
	case util.TypeStructField:
		tVal, err = s.decodeStructValue(val, tType)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}
	if tType.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = tVal
	return
}
