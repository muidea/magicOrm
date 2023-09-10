package codec

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

type ElemDependValueFunc func(model.Value) ([]model.Value, error)

type Codec interface {
	Encode(vVal model.Value, vType model.Type) (ret interface{}, err error)
	Decode(val interface{}, vType model.Type) (ret model.Value, err error)
}

type impl struct {
	elemDependValue ElemDependValueFunc
}

func New(elemDependValue ElemDependValueFunc) Codec {
	return &impl{elemDependValue: elemDependValue}
}

func (s *impl) Encode(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	if !vType.IsBasic() || vVal.IsNil() {
		err = fmt.Errorf("encode value failed, illegal value or type")
		return
	}

	switch vType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = s.encodeBool(vVal, vType)
	case model.TypeDateTimeValue:
		ret, err = s.encodeDateTime(vVal, vType)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.encodeFloat(vVal, vType)
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
		ret, err = s.encodeInt(vVal, vType)
	case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
		ret, err = s.encodeUint(vVal, vType)
	case model.TypeSliceValue:
		ret, err = s.encodeSlice(vVal, vType)
	case model.TypeStringValue:
		ret, err = s.encodeString(vVal, vType)
	default:
		err = fmt.Errorf("illegal type, type:%s", vType.GetName())
	}

	return
}

func (s *impl) Decode(val interface{}, vType model.Type) (ret model.Value, err error) {
	if !vType.IsBasic() {
		err = fmt.Errorf("illegal value type, type:%s", vType.GetName())
		return
	}

	switch vType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = s.decodeBool(val, vType)
	case model.TypeDateTimeValue:
		ret, err = s.decodeDateTime(val, vType)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.decodeFloat(val, vType)
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
		ret, err = s.decodeInt(val, vType)
	case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
		ret, err = s.decodeUint(val, vType)
	case model.TypeSliceValue:
		ret, err = s.decodeSlice(val, vType)
	case model.TypeStringValue:
		ret, err = s.decodeString(val, vType)
	default:
		err = fmt.Errorf("illegal type, type:%s", vType.GetName())
	}

	if err != nil {
		return
	}

	return
}
