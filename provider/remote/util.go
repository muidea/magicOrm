package remote

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

var _declareObjectSliceValue SliceObjectValue
var _declareObjectValue ObjectValue

func getSliceInitValue(tType model.Type) (ret any) {
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
	default:
		err := fmt.Errorf("unexpected slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
		panic(err)
	}

	return
}

func getBasicInitValue(tType model.Type) (ret any) {
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
		ret = getSliceInitValue(tType)
	default:
		err := fmt.Errorf("unexpected basic item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
		panic(err)
	}

	return
}

func getStructInitValue(tType model.Type) (ret any) {
	if model.IsSliceType(tType.GetValue()) {
		sliceVal := _declareObjectSliceValue.Copy()
		sliceVal.Name = tType.GetName()
		sliceVal.PkgPath = tType.GetPkgPath()
		ret = sliceVal
		return
	}

	if model.IsStructType(tType.GetValue()) {
		valPtr := _declareObjectValue.Copy()
		valPtr.Name = tType.GetName()
		valPtr.PkgPath = tType.GetPkgPath()
		ret = valPtr
		return
	}

	return
}

func getInitializeValue(tType model.Type) (ret any) {
	if !tType.IsBasic() {
		ret = getStructInitValue(tType)
		return
	}

	ret = getBasicInitValue(tType)
	return
}
