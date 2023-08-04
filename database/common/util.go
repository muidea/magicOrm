package common

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
)

type relationType int

const (
	relationInvalid = 0
	relationHas1v1  = 1
	relationHas1vn  = 2
	relationRef1v1  = 3
	relationRef1vn  = 4
)

func (s relationType) String() string {
	return fmt.Sprintf("%d", s)
}

func getFieldInitializeValue(field model.Field) (ret interface{}, err error) {
	fType := field.GetType()
	switch fType.GetValue() {
	case model.TypeBooleanValue, model.TypeBitValue:
		val := int8(0)
		ret = &val
		break
	case model.TypeSmallIntegerValue:
		val := int16(0)
		ret = &val
		break
	case model.TypeIntegerValue:
		val := int(0)
		ret = &val
		break
	case model.TypeInteger32Value:
		val := int32(0)
		ret = &val
		break
	case model.TypeBigIntegerValue:
		val := int64(0)
		ret = &val
		break
	case model.TypePositiveBitValue:
		val := uint8(0)
		ret = &val
		break
	case model.TypePositiveSmallIntegerValue:
		val := uint16(0)
		ret = &val
		break
	case model.TypePositiveIntegerValue:
		val := uint(0)
		ret = &val
		break
	case model.TypePositiveInteger32Value:
		val := uint32(0)
		ret = &val
		break
	case model.TypePositiveBigIntegerValue:
		val := uint64(0)
		ret = &val
		break
	case model.TypeFloatValue:
		val := float32(0.00)
		ret = &val
		break
	case model.TypeDoubleValue:
		val := 0.0000
		ret = &val
		break
	case model.TypeStringValue, model.TypeDateTimeValue:
		val := ""
		ret = &val
		break
	case model.TypeSliceValue:
		if fType.IsBasic() {
			val := ""
			ret = &val
		} else {
			err = fmt.Errorf("no support fileType, name:%s, type:%d", field.GetName(), fType.GetValue())
		}
	default:
		err = fmt.Errorf("no support fileType, name:%s, type:%d", field.GetName(), fType.GetValue())
	}

	return
}

func getFieldRelation(info model.Field) (ret relationType) {
	fType := info.GetType()
	if fType.IsBasic() {
		return
	}

	isPtr := fType.Elem().IsPtrType() || fType.IsPtrType()
	isSlice := model.IsSliceType(fType.GetValue())

	if !isPtr && !isSlice {
		ret = relationHas1v1
		return
	}

	if !isPtr && isSlice {
		ret = relationHas1vn
		return
	}

	if isPtr && !isSlice {
		ret = relationRef1v1
		return
	}

	ret = relationRef1vn
	return
}
