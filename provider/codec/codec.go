package codec

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

type ElemDependValueFunc func(model.Value) ([]model.Value, error)

type Codec interface {
	Encode(vVal model.Value, tType model.Type) (ret interface{}, err error)
	Decode(val interface{}, tType model.Type) (ret model.Value, err error)
}

type impl struct {
	elemDependValue ElemDependValueFunc
}

func New(elemDependValue ElemDependValueFunc) Codec {
	return &impl{elemDependValue: elemDependValue}
}

func (s *impl) Encode(vVal model.Value, tType model.Type) (ret interface{}, err error) {
	if !tType.IsBasic() || vVal.IsNil() {
		err = fmt.Errorf("encode value failed, illegal value or type")
		return
	}

	switch tType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = s.encodeBool(vVal)
	case model.TypeDateTimeValue:
		ret, err = s.encodeDateTime(vVal)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.encodeFloat(vVal)
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
		ret, err = s.encodeInt(vVal)
	case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
		ret, err = s.encodeUint(vVal)
	case model.TypeSliceValue:
		ret, err = s.encodeSlice(vVal, tType)
	case model.TypeStringValue:
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

	switch tType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = s.decodeBool(val, tType)
	case model.TypeDateTimeValue:
		ret, err = s.decodeDateTime(val, tType)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.decodeFloat(val, tType)
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
		ret, err = s.decodeInt(val, tType)
	case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
		ret, err = s.decodeUint(val, tType)
	case model.TypeSliceValue:
		ret, err = s.decodeSlice(val, tType)
	case model.TypeStringValue:
		ret, err = s.decodeString(val, tType)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	if err != nil {
		return
	}

	//if tType.IsPtrType() && !ret.IsNil() {
	//	ret = ret.Addr()
	//}

	return
}
