package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
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
	case util.TypeBooleanValue:
		ret, err = getBoolSlice(tType)
	case util.TypeStringValue,
		util.TypeDateTimeValue:
		ret, err = getStringSlice(tType)
	case util.TypeBitValue:
		ret, err = getInt8Slice(tType)
	case util.TypeSmallIntegerValue:
		ret, err = getInt16Slice(tType)
	case util.TypeInteger32Value:
		ret, err = getInt32Slice(tType)
	case util.TypeIntegerValue:
		ret, err = getIntSlice(tType)
	case util.TypeBigIntegerValue:
		ret, err = getInt64Slice(tType)
	case util.TypePositiveBitValue:
		ret, err = getUInt8Slice(tType)
	case util.TypePositiveSmallIntegerValue:
		ret, err = getUInt16Slice(tType)
	case util.TypePositiveInteger32Value:
		ret, err = getUInt32Slice(tType)
	case util.TypePositiveIntegerValue:
		ret, err = getUIntSlice(tType)
	case util.TypePositiveBigIntegerValue:
		ret, err = getUInt64Slice(tType)
	case util.TypeFloatValue:
		ret, err = getFloat32Slice(tType)
	case util.TypeDoubleValue:
		ret, err = getFloat64Slice(tType)
	case util.TypeStructValue:
		ret = reflect.TypeOf(&_declareObjectSliceValue)
	default:
		err = fmt.Errorf("unexpect slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
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
	case util.TypeBooleanValue:
		ret, err = getBool(tType)
	case util.TypeStringValue,
		util.TypeDateTimeValue:
		ret, err = getSting(tType)
	case util.TypeBitValue:
		ret, err = getInt8(tType)
	case util.TypeSmallIntegerValue:
		ret, err = getInt16(tType)
	case util.TypeInteger32Value:
		ret, err = getInt32(tType)
	case util.TypeIntegerValue:
		ret, err = getInt(tType)
	case util.TypeBigIntegerValue:
		ret, err = getInt64(tType)
	case util.TypePositiveBitValue:
		ret, err = getUInt8(tType)
	case util.TypePositiveSmallIntegerValue:
		ret, err = getUInt16(tType)
	case util.TypePositiveInteger32Value:
		ret, err = getUInt32(tType)
	case util.TypePositiveIntegerValue:
		ret, err = getUInt(tType)
	case util.TypePositiveBigIntegerValue:
		ret, err = getUInt64(tType)
	case util.TypeFloatValue:
		ret, err = getFloat32(tType)
	case util.TypeDoubleValue:
		ret, err = getFloat64(tType)
	case util.TypeStructValue:
		ret = reflect.TypeOf(&_declareObjectValue)
	case util.TypeSliceValue:
		ret, err = getSliceType(tType)
	default:
		err = fmt.Errorf("unexpect item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
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
	if util.IsSliceType(tType.GetValue()) {
		cValue.FieldByName("IsElemPtr").SetBool(tType.Elem().IsPtrType())
	}

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
