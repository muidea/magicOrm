package remote

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

var _declareObjectSliceValue SliceObjectValue
var _declareObjectValue ObjectValue

func getSliceType(tType model.Type) (ret any) {
	eType := tType.Elem()
	switch eType.GetValue() {
	case model.TypeBooleanValue:
		ret = []bool{}
	case model.TypeStringValue,
		model.TypeDateTimeValue:
		ret = []string{}
	case model.TypeBitValue:
		ret = []int8{}
	case model.TypeSmallIntegerValue:
		ret = []int16{}
	case model.TypeInteger32Value:
		ret = []int32{}
	case model.TypeIntegerValue:
		ret = []int{}
	case model.TypeBigIntegerValue:
		ret = []int64{}
	case model.TypePositiveBitValue:
		ret = []uint8{}
	case model.TypePositiveSmallIntegerValue:
		ret = []uint16{}
	case model.TypePositiveInteger32Value:
		ret = []uint32{}
	case model.TypePositiveIntegerValue:
		ret = []uint{}
	case model.TypePositiveBigIntegerValue:
		ret = []uint64{}
	case model.TypeFloatValue:
		ret = []float32{}
	case model.TypeDoubleValue:
		ret = []float64{}
	case model.TypeStructValue:
		ret = &_declareObjectSliceValue
	default:
		err := fmt.Errorf("unexpected slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
		panic(err)
	}

	return
}

func getBasicValue(tType model.Type) (ret any) {
	switch tType.GetValue() {
	case model.TypeBooleanValue:
		ret = false
	case model.TypeStringValue,
		model.TypeDateTimeValue:
		ret = ""
	case model.TypeBitValue:
		ret = int8(0)
	case model.TypeSmallIntegerValue:
		ret = int16(0)
	case model.TypeInteger32Value:
		ret = int32(0)
	case model.TypeIntegerValue:
		ret = 0
	case model.TypeBigIntegerValue:
		ret = int64(0)
	case model.TypePositiveBitValue:
		ret = uint8(0)
	case model.TypePositiveSmallIntegerValue:
		ret = uint16(0)
	case model.TypePositiveInteger32Value:
		ret = uint32(0)
	case model.TypePositiveIntegerValue:
		ret = uint(0)
	case model.TypePositiveBigIntegerValue:
		ret = uint64(0)
	case model.TypeFloatValue:
		ret = float32(0.00)
	case model.TypeDoubleValue:
		ret = 0.00
	case model.TypeSliceValue:
		ret = getSliceType(tType)
	default:
		err := fmt.Errorf("unexpected basic item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
		panic(err)
	}

	return
}

func getStructValue(tType model.Type) (ret any) {
	if model.IsSliceType(tType.GetValue()) {
		_declareObjectSliceValue.Name = tType.GetName()
		_declareObjectSliceValue.PkgPath = tType.GetPkgPath()

		ret = _declareObjectSliceValue.Copy()
		return
	}

	if model.IsStructType(tType.GetValue()) {
		_declareObjectValue.Name = tType.GetName()
		_declareObjectValue.PkgPath = tType.GetPkgPath()
		ret = _declareObjectValue.Copy()
		return
	}

	return
}

func getInitializeValue(tType model.Type) (ret any) {
	if !tType.IsBasic() {
		ret = getStructValue(tType)
		return
	}

	ret = getBasicValue(tType)
	return
}
