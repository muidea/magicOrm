package remote

import (
	"fmt"
	"reflect"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

var _declareObjectSliceValue SliceObjectValue
var _declareObjectValue ObjectValue

func getSliceInitValue(tType models.Type) (ret any) {
	eType := tType.Elem()
	switch eType.GetValue() {
	case models.TypeBooleanValue:
		ret = []bool{}
	case models.TypeStringValue,
		models.TypeDateTimeValue:
		ret = []string{}
	case models.TypeByteValue:
		ret = []int8{}
	case models.TypeSmallIntegerValue:
		ret = []int16{}
	case models.TypeInteger32Value:
		ret = []int32{}
	case models.TypeIntegerValue:
		ret = []int{}
	case models.TypeBigIntegerValue:
		ret = []int64{}
	case models.TypePositiveByteValue:
		ret = []uint8{}
	case models.TypePositiveSmallIntegerValue:
		ret = []uint16{}
	case models.TypePositiveInteger32Value:
		ret = []uint32{}
	case models.TypePositiveIntegerValue:
		ret = []uint{}
	case models.TypePositiveBigIntegerValue:
		ret = []uint64{}
	case models.TypeFloatValue:
		ret = []float32{}
	case models.TypeDoubleValue:
		ret = []float64{}
	default:
		err := fmt.Errorf("unexpected slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
		panic(err)
	}

	return
}

func getBasicInitValue(tType models.Type) (ret any) {
	switch tType.GetValue() {
	case models.TypeBooleanValue:
		ret = false
	case models.TypeStringValue,
		models.TypeDateTimeValue:
		ret = ""
	case models.TypeByteValue:
		ret = int8(0)
	case models.TypeSmallIntegerValue:
		ret = int16(0)
	case models.TypeInteger32Value:
		ret = int32(0)
	case models.TypeIntegerValue:
		ret = 0
	case models.TypeBigIntegerValue:
		ret = int64(0)
	case models.TypePositiveByteValue:
		ret = uint8(0)
	case models.TypePositiveSmallIntegerValue:
		ret = uint16(0)
	case models.TypePositiveInteger32Value:
		ret = uint32(0)
	case models.TypePositiveIntegerValue:
		ret = uint(0)
	case models.TypePositiveBigIntegerValue:
		ret = uint64(0)
	case models.TypeFloatValue:
		ret = float32(0.00)
	case models.TypeDoubleValue:
		ret = 0.00
	case models.TypeSliceValue:
		ret = getSliceInitValue(tType)
	default:
		err := fmt.Errorf("unexpected basic item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
		panic(err)
	}

	return
}

func getStructInitValue(tType models.Type) (ret *ObjectValue) {
	if models.IsStructType(tType.GetValue()) {
		valPtr := _declareObjectValue.Copy()
		valPtr.Name = tType.GetName()
		valPtr.PkgPath = tType.GetPkgPath()
		ret = valPtr
		return
	}

	return
}

func getSliceStructInitValue(tType models.Type) (ret *SliceObjectValue) {
	if models.IsSliceType(tType.GetValue()) {
		sliceVal := _declareObjectSliceValue.Copy()
		elemType := tType.Elem()
		sliceVal.Name = elemType.GetName()
		sliceVal.PkgPath = elemType.GetPkgPath()
		ret = sliceVal
		return
	}

	return
}

func getInitializeValue(tType models.Type) (ret any) {
	if !models.IsBasic(tType) {
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

func rewriteObjectValue(rawPtr *ObjectValue, srcPtr *ObjectValue) (err *cd.Error) {
	if rawPtr == nil || srcPtr == nil {
		return
	}
	rawPtr.Fields = srcPtr.Fields
	return
}

func rewriteSliceObjectValue(rawPtr *SliceObjectValue, srcPtr *SliceObjectValue) (err *cd.Error) {
	if rawPtr == nil || srcPtr == nil {
		return
	}

	rawPtr.Values = srcPtr.Values
	return
}

func appendBasicValue(sliceVal, val any) (ret any, err *cd.Error) {
	rVal := reflect.ValueOf(sliceVal)
	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "value is not slice")
		slog.Error("appendBasicValue: value is not slice")
		return
	}

	// Check if the types are compatible
	elemType := rVal.Type().Elem()
	valType := reflect.TypeOf(val)

	if !valType.AssignableTo(elemType) {
		err = cd.NewError(cd.Unexpected, "type mismatch, expected: "+elemType.String()+", got: "+valType.String())
		slog.Warn("appendBasicValue: type mismatch", "expected", elemType.String(), "got", valType.String())
		return
	}

	// Create a new slice with the appended value
	newSlice := reflect.Append(rVal, reflect.ValueOf(val))
	ret = newSlice.Interface()
	return
}

func convertValue(vType models.Type, val any) (ret any, err *cd.Error) {
	if val == nil {
		return
	}
	if !models.IsBasic(vType) && utils.IsReallyValidValue(val) {
		ret = val
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

func convertSliceValue(vType models.Type, val any) (ret any, err *cd.Error) {
	if val == nil {
		return
	}
	rVal := reflect.ValueOf(val)
	rVal = reflect.Indirect(rVal)
	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "value is not slice")
		slog.Error("convertSliceValue: value is not slice")
		return
	}
	if !models.IsBasic(vType) {
		for idx := 0; idx < rVal.Len(); idx++ {
			if !utils.IsReallyValidValue(rVal.Index(idx).Interface()) {
				break
			}
			if idx == rVal.Len()-1 {
				ret = val
				return
			}
		}
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

func isBasicCollectionValue(val any) bool {
	if val == nil {
		return false
	}

	rVal := reflect.ValueOf(val)
	if !rVal.IsValid() || rVal.Kind() != reflect.Slice {
		return false
	}
	if rVal.IsNil() {
		return false
	}
	if rVal.Type().Elem().Kind() != reflect.Interface {
		return false
	}
	if rVal.Len() == 0 {
		return true
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		itemVal := rVal.Index(idx)
		if !itemVal.IsValid() {
			return false
		}
		if itemVal.Kind() == reflect.Interface {
			if itemVal.IsNil() {
				return false
			}
			itemVal = itemVal.Elem()
		}
		if !utils.IsReallyValidValueForReflect(itemVal) {
			return false
		}
	}

	return true
}

func isSupportedBasicValue(val any) bool {
	return utils.IsReallyValidValue(val) || isBasicCollectionValue(val)
}

func isValid(val any) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case *ObjectValue:
		return v != nil
	case *SliceObjectValue:
		return v != nil
	case ObjectValue:
		return true
	case SliceObjectValue:
		return true
	default:
		return isSupportedBasicValue(val)
	}
}

func isZero(val any) (ret bool) {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case *ObjectValue:
		return v == nil || len(v.Fields) == 0 || !v.IsAssigned()
	case *SliceObjectValue:
		return v == nil || v.Values == nil
	case ObjectValue:
		return len(v.Fields) == 0 || !v.IsAssigned()
	case SliceObjectValue:
		return v.Values == nil
	default:
		if isBasicCollectionValue(val) {
			rVal := reflect.ValueOf(val)
			return !rVal.IsValid() || rVal.IsNil()
		}
		return utils.IsReallyZeroValue(val)
	}
}
