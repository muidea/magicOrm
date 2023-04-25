package remote

import (
	"fmt"
	"reflect"

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
	case util.TypeBooleanField:
		ret, err = getBoolSlice(tType)
	case util.TypeStringField,
		util.TypeDateTimeField:
		ret, err = getStringSlice(tType)
	case util.TypeBitField:
		ret, err = getInt8Slice(tType)
	case util.TypeSmallIntegerField:
		ret, err = getInt16Slice(tType)
	case util.TypeInteger32Field:
		ret, err = getInt32Slice(tType)
	case util.TypeIntegerField:
		ret, err = getIntSlice(tType)
	case util.TypeBigIntegerField:
		ret, err = getInt64Slice(tType)
	case util.TypePositiveBitField:
		ret, err = getUInt8Slice(tType)
	case util.TypePositiveSmallIntegerField:
		ret, err = getUInt16Slice(tType)
	case util.TypePositiveInteger32Field:
		ret, err = getUInt32Slice(tType)
	case util.TypePositiveIntegerField:
		ret, err = getUIntSlice(tType)
	case util.TypePositiveBigIntegerField:
		ret, err = getUInt64Slice(tType)
	case util.TypeFloatField:
		ret, err = getFloat32Slice(tType)
	case util.TypeDoubleField:
		ret, err = getFloat64Slice(tType)
	case util.TypeStructField:
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
	case util.TypeBooleanField:
		ret, err = getBool(tType)
	case util.TypeStringField,
		util.TypeDateTimeField:
		ret, err = getSting(tType)
	case util.TypeBitField:
		ret, err = getInt8(tType)
	case util.TypeSmallIntegerField:
		ret, err = getInt16(tType)
	case util.TypeInteger32Field:
		ret, err = getInt32(tType)
	case util.TypeIntegerField:
		ret, err = getInt(tType)
	case util.TypeBigIntegerField:
		ret, err = getInt64(tType)
	case util.TypePositiveBitField:
		ret, err = getUInt8(tType)
	case util.TypePositiveSmallIntegerField:
		ret, err = getUInt16(tType)
	case util.TypePositiveInteger32Field:
		ret, err = getUInt32(tType)
	case util.TypePositiveIntegerField:
		ret, err = getUInt(tType)
	case util.TypePositiveBigIntegerField:
		ret, err = getUInt64(tType)
	case util.TypeFloatField:
		ret, err = getFloat32(tType)
	case util.TypeDoubleField:
		ret, err = getFloat64(tType)
	case util.TypeStructField:
		ret = reflect.TypeOf(&_declareObjectValue)
	case util.TypeSliceField:
		ret, err = getSliceType(tType)
	default:
		err = fmt.Errorf("unexpect item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}

func getInitializeValue(tType model.Type) (ret reflect.Value) {
	cType, _ := getType(tType)
	if tType.IsPtrType() || !tType.IsBasic() {
		cType = cType.Elem()
	}

	cValue := reflect.New(cType).Elem()
	if !tType.IsBasic() {
		cValue.FieldByName("Name").SetString(tType.GetName())
		cValue.FieldByName("PkgPath").SetString(tType.GetPkgPath())
		cValue.FieldByName("IsPtr").SetBool(tType.IsPtrType())
		if util.IsSliceType(tType.GetValue()) {
			cValue.FieldByName("IsElemPtr").SetBool(tType.Elem().IsPtrType())
		}
	}

	if tType.IsPtrType() || !tType.IsBasic() {
		cValue = cValue.Addr()
	}

	ret = cValue
	return
}
