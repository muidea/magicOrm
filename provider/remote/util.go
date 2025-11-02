package remote

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
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
	case model.TypeByteValue:
		ret = []int8{}
	case model.TypeSmallIntegerValue:
		ret = []int16{}
	case model.TypeInteger32Value:
		ret = []int32{}
	case model.TypeIntegerValue:
		ret = []int{}
	case model.TypeBigIntegerValue:
		ret = []int64{}
	case model.TypePositiveByteValue:
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
	case model.TypeByteValue:
		ret = int8(0)
	case model.TypeSmallIntegerValue:
		ret = int16(0)
	case model.TypeInteger32Value:
		ret = int32(0)
	case model.TypeIntegerValue:
		ret = 0
	case model.TypeBigIntegerValue:
		ret = int64(0)
	case model.TypePositiveByteValue:
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

func getStructInitValue(tType model.Type) (ret *ObjectValue) {
	if model.IsStructType(tType.GetValue()) {
		valPtr := _declareObjectValue.Copy()
		valPtr.Name = tType.GetName()
		valPtr.PkgPath = tType.GetPkgPath()
		ret = valPtr
		return
	}

	return
}

func getSliceStructInitValue(tType model.Type) (ret *SliceObjectValue) {
	if model.IsSliceType(tType.GetValue()) {
		sliceVal := _declareObjectSliceValue.Copy()
		sliceVal.Name = tType.GetName()
		sliceVal.PkgPath = tType.GetPkgPath()
		ret = sliceVal
		return
	}

	return
}

func getInitializeValue(tType model.Type) (ret any) {
	if !model.IsBasic(tType) {
		if tType.GetValue().IsSliceType() {
			ret = getSliceStructInitValue(tType)
			return
		}

		ret = getStructInitValue(tType)
		return
	}

	ret = getBasicInitValue(tType)
	return
}

func rewriteObjectValue(rawPtr *ObjectValue, newPtr *ObjectValue) (err *cd.Error) {
	if rawPtr == nil || newPtr == nil {
		return
	}
	if rawPtr.PkgPath != newPtr.PkgPath {
		err = cd.NewError(cd.Unexpected, "illegal object value")
		return
	}

	rawPtr.Fields = newPtr.Fields
	return
}

func rewriteSliceObjectValue(rawPtr *SliceObjectValue, newPtr *SliceObjectValue) (err *cd.Error) {
	if rawPtr == nil || newPtr == nil {
		return
	}
	if rawPtr.PkgPath != newPtr.PkgPath {
		err = cd.NewError(cd.Unexpected, "illegal slice object value")
		return
	}

	rawPtr.Values = newPtr.Values
	return
}

func appendBasicValue(sliceVal, val any) (ret any, err *cd.Error) {
	rVal := reflect.ValueOf(sliceVal)
	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "value is not slice")
		log.Warnf("Append failed, value is not slice")
		return
	}

	// Check if the types are compatible
	elemType := rVal.Type().Elem()
	valType := reflect.TypeOf(val)

	if !valType.AssignableTo(elemType) {
		err = cd.NewError(cd.Unexpected, "type mismatch, expected: "+elemType.String()+", got: "+valType.String())
		log.Warnf("Append failed, type mismatch, expected: %s, got: %s", elemType, valType)
		return
	}

	// Create a new slice with the appended value
	newSlice := reflect.Append(rVal, reflect.ValueOf(val))
	ret = newSlice.Interface()
	return
}

func convertValue(vType model.Type, val any) (ret any, err *cd.Error) {
	if val == nil {
		return
	}

	vVal, vErr := vType.Interface(val)
	if vErr != nil {
		err = vErr
		return
	}

	ret = vVal.Get()
	return
}

func convertSliceValue(vType model.Type, val any) (ret any, err *cd.Error) {
	if val == nil {
		return
	}
	rVal := reflect.ValueOf(val)
	rVal = reflect.Indirect(rVal)
	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "value is not slice")
		log.Warnf("convertSliceValue failed, value is not slice")
		return
	}
	sliceVal := []any{}
	for idx := 0; idx < rVal.Len(); idx++ {
		val := rVal.Index(idx)
		vVal, vErr := vType.Interface(val.Interface())
		if vErr != nil {
			err = vErr
			return
		}
		sliceVal = append(sliceVal, vVal.Get())
	}

	ret = sliceVal
	return
}
