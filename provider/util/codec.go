package util

import (
	"fmt"
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
)

type ElemDependValueFunc func(interface{}) ([]model.Value, *cd.Result)

type Codec interface {
	Encode(vVal model.Value, vType model.Type) (ret interface{}, err *cd.Result)
	Decode(val interface{}, vType model.Type) (ret model.Value, err *cd.Result)
}

type impl struct {
	elemDependValue ElemDependValueFunc
}

func New(elemDependValue ElemDependValueFunc) Codec {
	return &impl{elemDependValue: elemDependValue}
}

func (s *impl) Encode(vVal model.Value, vType model.Type) (ret interface{}, err *cd.Result) {
	if !vType.IsBasic() || vVal.IsNil() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encode value failed, illegal value or type"))
		return
	}

	switch vType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = GetInt8(vVal.Interface())
	case model.TypeBitValue:
		ret, err = GetInt8(vVal.Interface())
	case model.TypeSmallIntegerValue:
		ret, err = GetInt16(vVal.Interface())
	case model.TypeInteger32Value:
		ret, err = GetInt32(vVal.Interface())
	case model.TypeBigIntegerValue:
		ret, err = GetInt64(vVal.Interface())
	case model.TypeIntegerValue:
		ret, err = GetInt(vVal.Interface())
	case model.TypePositiveBitValue:
		ret, err = GetUint8(vVal.Interface())
	case model.TypePositiveSmallIntegerValue:
		ret, err = GetUint16(vVal.Interface())
	case model.TypePositiveInteger32Value:
		ret, err = GetUint32(vVal.Interface())
	case model.TypePositiveBigIntegerValue:
		ret, err = GetUint64(vVal.Interface())
	case model.TypePositiveIntegerValue:
		ret, err = GetUint(vVal.Interface())
	case model.TypeFloatValue:
		ret, err = GetFloat32(vVal.Interface())
	case model.TypeDoubleValue:
		ret, err = GetFloat64(vVal.Interface())
	case model.TypeStringValue:
		ret, err = GetString(vVal.Interface())
	case model.TypeDateTimeValue:
		ret, err = GetString(vVal.Interface())
	case model.TypeSliceValue:
		ret, err = s.encodeSlice(vVal, vType)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal type, type:%s", vType.GetName()))
	}

	return
}

// encodeSlice get slice value str
func (s *impl) encodeSlice(vVal model.Value, vType model.Type) (ret interface{}, err *cd.Result) {
	vals, valErr := s.elemDependValue(vVal.Interface())
	if valErr != nil {
		err = valErr
		return
	}
	if len(vals) == 0 {
		return
	}
	items := []interface{}{}
	for _, val := range vals {
		encodeVal, encodeErr := s.Encode(val, vType.Elem())
		if encodeErr != nil {
			err = encodeErr
			return
		}

		items = append(items, encodeVal)
	}

	ret = items
	return
}

func (s *impl) Decode(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	if !vType.IsBasic() {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal value type, type:%s", vType.GetName()))
		return
	}

	switch vType.GetValue() {
	case model.TypeBooleanValue,
		model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue,
		model.TypeFloatValue, model.TypeDoubleValue,
		model.TypeStringValue,
		model.TypeDateTimeValue:
		ret, err = vType.Interface(val)
	case model.TypeSliceValue:
		ret, err = s.decodeSlice(val, vType)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal type, type:%s", vType.GetName()))
	}

	if err != nil {
		return
	}

	return
}

// decodeSlice decode slice from string
func (s *impl) decodeSlice(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	if val == nil {
		ret, err = vType.Interface(nil)
		return
	}

	rVal := reflect.Indirect(reflect.ValueOf(val))
	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.UnExpected, "illegal value type, not slice value")
		return
	}

	sType := vType.Elem()
	switch sType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = s.decodeBoolSlice(rVal, vType)
	case model.TypeBitValue:
		ret, err = s.decodeInt8Slice(rVal, vType)
	case model.TypeSmallIntegerValue:
		ret, err = s.decodeInt16Slice(rVal, vType)
	case model.TypeInteger32Value:
		ret, err = s.decodeInt32Slice(rVal, vType)
	case model.TypeBigIntegerValue:
		ret, err = s.decodeInt64Slice(rVal, vType)
	case model.TypeIntegerValue:
		ret, err = s.decodeIntSlice(rVal, vType)
	case model.TypePositiveBitValue:
		ret, err = s.decodeUint8Slice(rVal, vType)
	case model.TypePositiveSmallIntegerValue:
		ret, err = s.decodeUint16Slice(rVal, vType)
	case model.TypePositiveInteger32Value:
		ret, err = s.decodeUint32Slice(rVal, vType)
	case model.TypePositiveBigIntegerValue:
		ret, err = s.decodeUint64Slice(rVal, vType)
	case model.TypePositiveIntegerValue:
		ret, err = s.decodeUintSlice(rVal, vType)
	case model.TypeFloatValue:
		ret, err = s.decodeFloatSlice(rVal, vType)
	case model.TypeDoubleValue:
		ret, err = s.decodeDoubleSlice(rVal, vType)
	case model.TypeStringValue:
		ret, err = s.decodeStringSlice(rVal, vType)
	case model.TypeDateTimeValue:
		ret, err = s.decodeDateTimeSlice(rVal, vType)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal slice element type, type pkgKey:%s", vType.GetPkgKey()))
	}

	return
}

func (s *impl) decodeBoolSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []bool{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawBool(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt8Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int8{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawInt8(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt16Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int16{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawInt16(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt32Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int32{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawInt32(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt64Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int64{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawInt64(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeIntSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawInt(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint8Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint8{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawUint8(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint16Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint16{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawUint16(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint32Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint32{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawUint32(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint64Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint64{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawUint64(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUintSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawUint(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeFloatSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []float32{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawFloat32(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeDoubleSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []float64{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawFloat64(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeStringSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []string{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawString(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeDateTimeSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []time.Time{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := GetRawDateTime(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		sliceVal = append(sliceVal, eVal)
	}

	ret, err = rType.Interface(sliceVal)
	return
}
