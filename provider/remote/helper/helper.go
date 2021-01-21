package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type GetValueFunc func(interface{}) (model.Value, error)
type ElemDependValueFunc func(model.Value) ([]model.Value, error)

type Helper interface {
	Encode(vVal model.Value, tType model.Type) (ret string, err error)
	Decode(val interface{}, tType model.Type) (ret model.Value, err error)
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
		ret, err = s.encodeBool(vVal)
	case util.TypeDateTimeField:
		ret, err = s.encodeDateTime(vVal)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = s.encodeFloat(vVal)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = s.encodeInt(vVal)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = s.encodeUint(vVal)
	case util.TypeSliceField:
		ret, err = s.encodeSlice(vVal, tType)
	case util.TypeStringField:
		ret, err = s.encodeString(vVal)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}

func (s *impl) Decode(val interface{}, tType model.Type) (ret model.Value, err error) {
	if !tType.IsBasic() {
		err = fmt.Errorf("illegal value type, type:%s", tType.GetName())
		return
	}

	tVal := reflect.ValueOf(val)
	if tVal.Kind() == reflect.Interface {
		tVal = tVal.Elem()
	}
	tVal = reflect.Indirect(tVal)

	cValue, cErr := GetTypeValue(tType)
	if cErr != nil {
		err = cErr
		return
	}

	cValue, cErr = s.decodeInternal(tVal, tType, cValue)
	if cErr != nil {
		err = cErr
		return
	}
	ret, err = s.getValue(cValue.Interface())
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}

	return
}

func (s *impl) decodeInternal(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	switch tType.GetValue() {
	case util.TypeBooleanField:
		ret, err = s.decodeBool(tVal, tType, cVal)
	case util.TypeDateTimeField:
		ret, err = s.decodeDateTime(tVal, tType, cVal)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = s.decodeFloat(tVal, tType, cVal)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = s.decodeInt(tVal, tType, cVal)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = s.decodeUint(tVal, tType, cVal)
	case util.TypeSliceField:
		ret, err = s.decodeSlice(tVal, tType, cVal)
	case util.TypeStringField:
		ret, err = s.decodeString(tVal, tType, cVal)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}
