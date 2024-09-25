package codec

import (
	"fmt"
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type ElemDependValueFunc func(eVal interface{}) ([]model.Value, *cd.Result)

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

	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch vType.GetValue() {
	case model.TypeBooleanValue:
		ret, err = pu.GetRawInt8(val)
	case model.TypeBitValue:
		ret, err = pu.GetRawInt8(val)
	case model.TypeSmallIntegerValue:
		ret, err = pu.GetRawInt16(val)
	case model.TypeInteger32Value:
		ret, err = pu.GetRawInt32(val)
	case model.TypeBigIntegerValue:
		ret, err = pu.GetRawInt64(val)
	case model.TypeIntegerValue:
		ret, err = pu.GetRawInt(val)
	case model.TypePositiveBitValue:
		ret, err = pu.GetRawUint8(val)
	case model.TypePositiveSmallIntegerValue:
		ret, err = pu.GetRawUint16(val)
	case model.TypePositiveInteger32Value:
		ret, err = pu.GetRawUint32(val)
	case model.TypePositiveBigIntegerValue:
		ret, err = pu.GetRawUint64(val)
	case model.TypePositiveIntegerValue:
		ret, err = pu.GetRawUint(val)
	case model.TypeFloatValue:
		ret, err = pu.GetRawFloat32(val)
	case model.TypeDoubleValue:
		ret, err = pu.GetRawFloat64(val)
	case model.TypeStringValue:
		ret, err = pu.GetRawString(val)
	case model.TypeDateTimeValue:
		ret, err = pu.GetRawString(val)
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
		strVal, strErr := s.Encode(val, vType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		items = append(items, strVal)
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
	slicePtrVal := []*bool{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawBool(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt8Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int8{}
	slicePtrVal := []*int8{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawInt8(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt16Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int16{}
	slicePtrVal := []*int16{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawInt16(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt32Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int32{}
	slicePtrVal := []*int32{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawInt32(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeInt64Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int64{}
	slicePtrVal := []*int64{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawInt64(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeIntSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []int{}
	slicePtrVal := []*int{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawInt(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint8Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint8{}
	slicePtrVal := []*uint8{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawUint8(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint16Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint16{}
	slicePtrVal := []*uint16{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawUint16(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint32Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint32{}
	slicePtrVal := []*uint32{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawUint32(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUint64Slice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint64{}
	slicePtrVal := []*uint64{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawUint64(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeUintSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []uint{}
	slicePtrVal := []*uint{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawUint(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeFloatSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []float32{}
	slicePtrVal := []*float32{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawFloat32(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeDoubleSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []float64{}
	slicePtrVal := []*float64{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawFloat64(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeStringSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []string{}
	slicePtrVal := []*string{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawString(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}

func (s *impl) decodeDateTimeSlice(rVal reflect.Value, rType model.Type) (ret model.Value, err *cd.Result) {
	sliceVal := []time.Time{}
	slicePtrVal := []*time.Time{}
	for idx := 0; idx < rVal.Len(); idx++ {
		iVal := reflect.Indirect(rVal.Index(idx))
		if iVal.Kind() == reflect.Interface {
			iVal = iVal.Elem()
		}
		eVal, eErr := pu.GetRawDateTime(iVal)
		if eErr != nil {
			err = eErr
			return
		}
		if rType.Elem().IsPtrType() {
			slicePtrVal = append(slicePtrVal, &eVal)
			continue
		}

		sliceVal = append(sliceVal, eVal)
	}

	if rType.Elem().IsPtrType() {
		ret, err = rType.Interface(slicePtrVal)
		return
	}
	ret, err = rType.Interface(sliceVal)
	return
}
