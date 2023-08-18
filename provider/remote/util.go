package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
)

var _declareObjectSliceValue SliceObjectValue
var _declareObjectValue ObjectValue

func getBoolSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*bool
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]bool
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*bool
		ret = reflect.TypeOf(val)
		return
	}
	var val []bool
	ret = reflect.TypeOf(val)
	return
}

func getStringSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*string
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]string
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*string
		ret = reflect.TypeOf(val)
		return
	}
	var val []string
	ret = reflect.TypeOf(val)
	return
}

func getInt8Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*int8
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]int8
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*int8
		ret = reflect.TypeOf(val)
		return
	}
	var val []int8
	ret = reflect.TypeOf(val)
	return
}

func getInt16Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*int16
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]int16
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*int16
		ret = reflect.TypeOf(val)
		return
	}
	var val []int16
	ret = reflect.TypeOf(val)
	return
}

func getInt32Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*int32
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]int32
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*int32
		ret = reflect.TypeOf(val)
		return
	}
	var val []int32
	ret = reflect.TypeOf(val)
	return
}

func getIntSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*int
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]int
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*int
		ret = reflect.TypeOf(val)
		return
	}
	var val []int
	ret = reflect.TypeOf(val)
	return
}

func getInt64Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*int64
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]int64
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*int64
		ret = reflect.TypeOf(val)
		return
	}
	var val []int64
	ret = reflect.TypeOf(val)
	return
}

func getUInt8Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*uint8
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]uint8
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*uint8
		ret = reflect.TypeOf(val)
		return
	}
	var val []uint8
	ret = reflect.TypeOf(val)
	return
}

func getUInt16Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*uint16
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]uint16
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*uint16
		ret = reflect.TypeOf(val)
		return
	}
	var val []uint16
	ret = reflect.TypeOf(val)
	return
}

func getUInt32Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*uint32
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]uint32
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*uint32
		ret = reflect.TypeOf(val)
		return
	}
	var val []uint32
	ret = reflect.TypeOf(val)
	return
}

func getUIntSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*uint
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]uint
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*uint
		ret = reflect.TypeOf(val)
		return
	}
	var val []uint
	ret = reflect.TypeOf(val)
	return
}

func getUInt64Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*uint64
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]uint64
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*uint64
		ret = reflect.TypeOf(val)
		return
	}
	var val []uint64
	ret = reflect.TypeOf(val)
	return
}

func getFloat32Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*float32
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]float32
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*float32
		ret = reflect.TypeOf(val)
		return
	}
	var val []float32
	ret = reflect.TypeOf(val)
	return
}

func getFloat64Slice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	if tType.IsPtrType() {
		if eType.IsPtrType() {
			var val *[]*float64
			ret = reflect.TypeOf(val)
			return
		}
		var val *[]float64
		ret = reflect.TypeOf(val)
		return
	}
	if eType.IsPtrType() {
		var val []*float64
		ret = reflect.TypeOf(val)
		return
	}
	var val []float64
	ret = reflect.TypeOf(val)
	return
}

func getSliceType(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	switch eType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = getBoolSlice(tType)
	case model.TypeStringValue,
		model.TypeDateTimeValue:
		ret, err = getStringSlice(tType)
	case model.TypeBitValue:
		ret, err = getInt8Slice(tType)
	case model.TypeSmallIntegerValue:
		ret, err = getInt16Slice(tType)
	case model.TypeInteger32Value:
		ret, err = getInt32Slice(tType)
	case model.TypeIntegerValue:
		ret, err = getIntSlice(tType)
	case model.TypeBigIntegerValue:
		ret, err = getInt64Slice(tType)
	case model.TypePositiveBitValue:
		ret, err = getUInt8Slice(tType)
	case model.TypePositiveSmallIntegerValue:
		ret, err = getUInt16Slice(tType)
	case model.TypePositiveInteger32Value:
		ret, err = getUInt32Slice(tType)
	case model.TypePositiveIntegerValue:
		ret, err = getUIntSlice(tType)
	case model.TypePositiveBigIntegerValue:
		ret, err = getUInt64Slice(tType)
	case model.TypeFloatValue:
		ret, err = getFloat32Slice(tType)
	case model.TypeDoubleValue:
		ret, err = getFloat64Slice(tType)
	case model.TypeStructValue:
		ret = reflect.TypeOf(&_declareObjectSliceValue)
	default:
		err = fmt.Errorf("unexpected slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}

func getBool(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *bool
		ret = reflect.TypeOf(val)
		return
	}
	var val bool
	ret = reflect.TypeOf(val)
	return
}

func getSting(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *string
		ret = reflect.TypeOf(val)
		return
	}
	var val string
	ret = reflect.TypeOf(val)
	return
}

func getInt8(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *int8
		ret = reflect.TypeOf(val)
		return
	}
	var val int8
	ret = reflect.TypeOf(val)
	return
}

func getInt16(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *int16
		ret = reflect.TypeOf(val)
		return
	}
	var val int16
	ret = reflect.TypeOf(val)
	return
}

func getInt32(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *int32
		ret = reflect.TypeOf(val)
		return
	}
	var val int32
	ret = reflect.TypeOf(val)
	return
}

func getInt(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *int
		ret = reflect.TypeOf(val)
		return
	}
	var val int
	ret = reflect.TypeOf(val)
	return
}

func getInt64(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *int64
		ret = reflect.TypeOf(val)
		return
	}
	var val int64
	ret = reflect.TypeOf(val)
	return
}

func getUInt8(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *uint8
		ret = reflect.TypeOf(val)
		return
	}
	var val uint8
	ret = reflect.TypeOf(val)
	return
}

func getUInt16(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *uint16
		ret = reflect.TypeOf(val)
		return
	}
	var val uint16
	ret = reflect.TypeOf(val)
	return
}

func getUInt32(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *uint32
		ret = reflect.TypeOf(val)
		return
	}
	var val uint32
	ret = reflect.TypeOf(val)
	return
}

func getUInt(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *uint
		ret = reflect.TypeOf(val)
		return
	}
	var val uint
	ret = reflect.TypeOf(val)
	return
}

func getUInt64(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *uint64
		ret = reflect.TypeOf(val)
		return
	}
	var val uint64
	ret = reflect.TypeOf(val)
	return
}

func getFloat32(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *float32
		ret = reflect.TypeOf(val)
		return
	}
	var val float32
	ret = reflect.TypeOf(val)
	return
}

func getFloat64(tType model.Type) (ret reflect.Type, err error) {
	if tType.IsPtrType() {
		var val *float64
		ret = reflect.TypeOf(val)
		return
	}
	var val float64
	ret = reflect.TypeOf(val)
	return
}

func getType(tType model.Type) (ret reflect.Type, err error) {
	switch tType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = getBool(tType)
	case model.TypeStringValue,
		model.TypeDateTimeValue:
		ret, err = getSting(tType)
	case model.TypeBitValue:
		ret, err = getInt8(tType)
	case model.TypeSmallIntegerValue:
		ret, err = getInt16(tType)
	case model.TypeInteger32Value:
		ret, err = getInt32(tType)
	case model.TypeIntegerValue:
		ret, err = getInt(tType)
	case model.TypeBigIntegerValue:
		ret, err = getInt64(tType)
	case model.TypePositiveBitValue:
		ret, err = getUInt8(tType)
	case model.TypePositiveSmallIntegerValue:
		ret, err = getUInt16(tType)
	case model.TypePositiveInteger32Value:
		ret, err = getUInt32(tType)
	case model.TypePositiveIntegerValue:
		ret, err = getUInt(tType)
	case model.TypePositiveBigIntegerValue:
		ret, err = getUInt64(tType)
	case model.TypeFloatValue:
		ret, err = getFloat32(tType)
	case model.TypeDoubleValue:
		ret, err = getFloat64(tType)
	case model.TypeStructValue:
		ret = reflect.TypeOf(&_declareObjectValue)
	case model.TypeSliceValue:
		ret, err = getSliceType(tType)
	default:
		err = fmt.Errorf("unexpected item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}

func getBasicValue(tType model.Type) (ret reflect.Value) {
	cType, cErr := getType(tType)
	if cErr != nil {
		log.Errorf("getBasicValue failed, err:%s", cErr.Error())
		return
	}
	cValue := reflect.New(cType).Elem()
	if tType.IsPtrType() {
		rVal := reflect.New(cType.Elem())
		cValue.Set(rVal)
	}

	ret = cValue
	return
}

func getStructValue(tType model.Type) (ret reflect.Value) {
	cType, cErr := getType(tType)
	if cErr != nil {
		log.Errorf("getStructValue failed, err:%s", cErr.Error())
		return
	}

	if cType.Kind() == reflect.Ptr {
		cType = cType.Elem()
	}

	cValue := reflect.New(cType).Elem()
	cValue.FieldByName("Name").SetString(tType.GetName())
	cValue.FieldByName("PkgPath").SetString(tType.GetPkgPath())

	ret = cValue
	return
}

func getInitializeValue(tType model.Type) (ret reflect.Value) {
	if !tType.IsBasic() {
		ret = getStructValue(tType)
		return
	}

	ret = getBasicValue(tType)
	return
}
