package common

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
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
	case util.TypeBooleanValue, util.TypeBitValue:
		val := int8(0)
		ret = &val
		break
	case util.TypeSmallIntegerValue:
		val := int16(0)
		ret = &val
		break
	case util.TypeIntegerValue:
		val := int(0)
		ret = &val
		break
	case util.TypeInteger32Value:
		val := int32(0)
		ret = &val
		break
	case util.TypeBigIntegerValue:
		val := int64(0)
		ret = &val
		break
	case util.TypePositiveBitValue:
		val := uint8(0)
		ret = &val
		break
	case util.TypePositiveSmallIntegerValue:
		val := uint16(0)
		ret = &val
		break
	case util.TypePositiveIntegerValue:
		val := uint(0)
		ret = &val
		break
	case util.TypePositiveInteger32Value:
		val := uint32(0)
		ret = &val
		break
	case util.TypePositiveBigIntegerValue:
		val := uint64(0)
		ret = &val
		break
	case util.TypeFloatValue:
		val := float32(0.00)
		ret = &val
		break
	case util.TypeDoubleValue:
		val := 0.0000
		ret = &val
		break
	case util.TypeStringValue, util.TypeDateTimeValue:
		val := ""
		ret = &val
		break
	case util.TypeSliceValue:
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
	isSlice := util.IsSliceType(fType.GetValue())

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
