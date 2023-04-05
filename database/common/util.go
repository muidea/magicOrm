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
	case util.TypeBooleanField, util.TypeBitField:
		val := int8(0)
		ret = &val
		break
	case util.TypeSmallIntegerField:
		val := int16(0)
		ret = &val
		break
	case util.TypeIntegerField:
		val := int(0)
		ret = &val
		break
	case util.TypeInteger32Field:
		val := int32(0)
		ret = &val
		break
	case util.TypeBigIntegerField:
		val := int64(0)
		ret = &val
		break
	case util.TypePositiveBitField:
		val := uint8(0)
		ret = &val
		break
	case util.TypePositiveSmallIntegerField:
		val := uint16(0)
		ret = &val
		break
	case util.TypePositiveIntegerField:
		val := uint(0)
		ret = &val
		break
	case util.TypePositiveInteger32Field:
		val := uint32(0)
		ret = &val
		break
	case util.TypePositiveBigIntegerField:
		val := uint64(0)
		ret = &val
		break
	case util.TypeFloatField:
		val := float32(0.00)
		ret = &val
		break
	case util.TypeDoubleField:
		val := 0.0000
		ret = &val
		break
	case util.TypeStringField, util.TypeDateTimeField:
		val := ""
		ret = &val
		break
	case util.TypeSliceField:
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
