package remote

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
)

type ValueConvertMap map[model.TypeDeclare]func(reflect.Value, model.Type) (any, *cd.Result)

var encodeValueConvertMap ValueConvertMap
var encodeValueConvertSliceMap ValueConvertMap

var decodeValueConvertMap ValueConvertMap
var decodeConvertSliceMap ValueConvertMap

func init() {
	encodeValueConvertMap = ValueConvertMap{
		model.TypeBooleanValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int8Val, int8Err := utils.ConvertToInt8(vVal)
			if int8Err != nil {
				err = int8Err
				return
			}
			if vType.IsPtrType() {
				ret = &int8Val
			} else {
				ret = int8Val
			}
			return
		},
		model.TypeBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int8Val, int8Err := utils.ConvertToInt8(vVal)
			if int8Err != nil {
				err = int8Err
				return
			}
			if vType.IsPtrType() {
				ret = &int8Val
			} else {
				ret = int8Val
			}
			return
		},
		model.TypeSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int16Val, int16Err := utils.ConvertToInt16(vVal)
			if int16Err != nil {
				err = int16Err
				return
			}
			if vType.IsPtrType() {
				ret = &int16Val
			} else {
				ret = int16Val
			}
			return
		},
		model.TypeInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int32Val, int32Err := utils.ConvertToInt32(vVal)
			if int32Err != nil {
				err = int32Err
				return
			}
			if vType.IsPtrType() {
				ret = &int32Val
			} else {
				ret = int32Val
			}
			return
		},
		model.TypeBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int64Val, int64Err := utils.ConvertToInt64(vVal)
			if int64Err != nil {
				err = int64Err
				return
			}
			if vType.IsPtrType() {
				ret = &int64Val
			} else {
				ret = int64Val
			}
			return
		},
		model.TypeIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			intVal, intErr := utils.ConvertToInt(vVal)
			if intErr != nil {
				err = intErr
				return
			}
			if vType.IsPtrType() {
				ret = &intVal
			} else {
				ret = intVal
			}
			return
		},
		model.TypePositiveBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint8Val, uint8Err := utils.ConvertToUint8(vVal)
			if uint8Err != nil {
				err = uint8Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint8Val
			} else {
				ret = uint8Val
			}
			return
		},
		model.TypePositiveSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint16Val, uint16Err := utils.ConvertToUint16(vVal)
			if uint16Err != nil {
				err = uint16Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint16Val
			} else {
				ret = uint16Val
			}
			return
		},
		model.TypePositiveInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint32Val, uint32Err := utils.ConvertToUint32(vVal)
			if uint32Err != nil {
				err = uint32Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint32Val
			} else {
				ret = uint32Val
			}
			return
		},
		model.TypePositiveBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint64Val, uint64Err := utils.ConvertToUint64(vVal)
			if uint64Err != nil {
				err = uint64Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint64Val
			} else {
				ret = uint64Val
			}
			return
		},
		model.TypePositiveIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uintVal, uintErr := utils.ConvertToUint(vVal)
			if uintErr != nil {
				err = uintErr
				return
			}
			if vType.IsPtrType() {
				ret = &uintVal
			} else {
				ret = uintVal
			}
			return
		},
		model.TypeFloatValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float32Val, float32Err := utils.ConvertToFloat32(vVal)
			if float32Err != nil {
				err = float32Err
				return
			}
			if vType.IsPtrType() {
				ret = &float32Val
			} else {
				ret = float32Val
			}
			return
		},
		model.TypeDoubleValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float64Val, float64Err := utils.ConvertToFloat64(vVal)
			if float64Err != nil {
				err = float64Err
				return
			}
			if vType.IsPtrType() {
				ret = &float64Val
			} else {
				ret = float64Val
			}
			return
		},
		model.TypeStringValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			strVal, strErr := utils.ConvertToString(vVal)
			if strErr != nil {
				err = strErr
				return
			}
			if vType.IsPtrType() {
				ret = &strVal
			} else {
				ret = strVal
			}
			return
		},
		model.TypeDateTimeValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			strVal, strErr := utils.ConvertToString(vVal)
			if strErr != nil {
				err = strErr
				return
			}
			if vType.IsPtrType() {
				ret = &strVal
			} else {
				ret = strVal
			}
			return
		},
		model.TypeSliceValue: encodeSliceValue,
	}

	encodeValueConvertSliceMap = ValueConvertMap{
		model.TypeBooleanValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			boolSlice, boolErr := encodeSliceTemplate(vVal, vType.Elem(), int8(0))
			if boolErr != nil {
				err = boolErr
				return
			}
			if vType.IsPtrType() {
				ret = &boolSlice
			} else {
				ret = boolSlice
			}
			return
		},
		model.TypeBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int8Slice, int8Err := encodeSliceTemplate(vVal, vType.Elem(), int8(0))
			if int8Err != nil {
				err = int8Err
				return
			}
			if vType.IsPtrType() {
				ret = &int8Slice
			} else {
				ret = int8Slice
			}
			return
		},
		model.TypeSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int16Slice, int16Err := encodeSliceTemplate(vVal, vType.Elem(), int16(0))
			if int16Err != nil {
				err = int16Err
				return
			}
			if vType.IsPtrType() {
				ret = &int16Slice
			} else {
				ret = int16Slice
			}
			return
		},
		model.TypeInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int32Slice, int32Err := encodeSliceTemplate(vVal, vType.Elem(), int32(0))
			if int32Err != nil {
				err = int32Err
				return
			}
			if vType.IsPtrType() {
				ret = &int32Slice
			} else {
				ret = int32Slice
			}
			return
		},
		model.TypeBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int64Slice, int64Err := encodeSliceTemplate(vVal, vType.Elem(), int64(0))
			if int64Err != nil {
				err = int64Err
				return
			}
			if vType.IsPtrType() {
				ret = &int64Slice
			} else {
				ret = int64Slice
			}
			return
		},
		model.TypeIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			intSlice, intErr := encodeSliceTemplate(vVal, vType.Elem(), int(0))
			if intErr != nil {
				err = intErr
				return
			}
			if vType.IsPtrType() {
				ret = &intSlice
			} else {
				ret = intSlice
			}
			return
		},
		model.TypePositiveBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint8Slice, uint8Err := encodeSliceTemplate(vVal, vType.Elem(), uint8(0))
			if uint8Err != nil {
				err = uint8Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint8Slice
			} else {
				ret = uint8Slice
			}
			return
		},
		model.TypePositiveSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint16Slice, uint16Err := encodeSliceTemplate(vVal, vType.Elem(), uint16(0))
			if uint16Err != nil {
				err = uint16Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint16Slice
			} else {
				ret = uint16Slice
			}
			return
		},
		model.TypePositiveInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint32Slice, uint32Err := encodeSliceTemplate(vVal, vType.Elem(), uint32(0))
			if uint32Err != nil {
				err = uint32Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint32Slice
			} else {
				ret = uint32Slice
			}
			return
		},
		model.TypePositiveBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint64Slice, uint64Err := encodeSliceTemplate(vVal, vType.Elem(), uint64(0))
			if uint64Err != nil {
				err = uint64Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint64Slice
			} else {
				ret = uint64Slice
			}
			return
		},
		model.TypePositiveIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uintSlice, uintErr := encodeSliceTemplate(vVal, vType.Elem(), uint(0))
			if uintErr != nil {
				err = uintErr
				return
			}
			if vType.IsPtrType() {
				ret = &uintSlice
			} else {
				ret = uintSlice
			}
			return
		},
		model.TypeFloatValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float32Slice, float32Err := encodeSliceTemplate(vVal, vType.Elem(), float32(0))
			if float32Err != nil {
				err = float32Err
				return
			}
			if vType.IsPtrType() {
				ret = &float32Slice
			} else {
				ret = float32Slice
			}
			return
		},
		model.TypeDoubleValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float64Slice, float64Err := encodeSliceTemplate(vVal, vType.Elem(), float64(0))
			if float64Err != nil {
				err = float64Err
				return
			}
			if vType.IsPtrType() {
				ret = &float64Slice
			} else {
				ret = float64Slice
			}
			return
		},
		model.TypeStringValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			strSlice, strErr := encodeSliceTemplate(vVal, vType.Elem(), "")
			if strErr != nil {
				err = strErr
				return
			}
			if vType.IsPtrType() {
				ret = &strSlice
			} else {
				ret = strSlice
			}
			return
		},
		model.TypeDateTimeValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			timeSlice, timeErr := encodeSliceTemplate(vVal, vType.Elem(), "")
			if timeErr != nil {
				err = timeErr
				return
			}
			if vType.IsPtrType() {
				ret = &timeSlice
			} else {
				ret = timeSlice
			}
			return
		},
	}

	decodeValueConvertMap = ValueConvertMap{
		model.TypeBooleanValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			boolVal, boolErr := utils.ConvertToBool(vVal)
			if boolErr != nil {
				err = boolErr
				return
			}
			if vType.IsPtrType() {
				ret = &boolVal
			} else {
				ret = boolVal
			}
			return
		},
		model.TypeBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int8Val, int8Err := utils.ConvertToInt8(vVal)
			if int8Err != nil {
				err = int8Err
				return
			}
			if vType.IsPtrType() {
				ret = &int8Val
			} else {
				ret = int8Val
			}
			return
		},
		model.TypeSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int16Val, int16Err := utils.ConvertToInt16(vVal)
			if int16Err != nil {
				err = int16Err
				return
			}
			if vType.IsPtrType() {
				ret = &int16Val
			} else {
				ret = int16Val
			}
			return
		},
		model.TypeInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int32Val, int32Err := utils.ConvertToInt32(vVal)
			if int32Err != nil {
				err = int32Err
				return
			}
			if vType.IsPtrType() {
				ret = &int32Val
			} else {
				ret = int32Val
			}
			return
		},
		model.TypeBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int64Val, int64Err := utils.ConvertToInt64(vVal)
			if int64Err != nil {
				err = int64Err
				return
			}
			if vType.IsPtrType() {
				ret = &int64Val
			} else {
				ret = int64Val
			}
			return
		},
		model.TypeIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			intVal, intErr := utils.ConvertToInt(vVal)
			if intErr != nil {
				err = intErr
				return
			}
			if vType.IsPtrType() {
				ret = &intVal
			} else {
				ret = intVal
			}
			return
		},
		model.TypePositiveBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint8Val, uint8Err := utils.ConvertToUint8(vVal)
			if uint8Err != nil {
				err = uint8Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint8Val
			} else {
				ret = uint8Val
			}
			return
		},
		model.TypePositiveSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint16Val, uint16Err := utils.ConvertToUint16(vVal)
			if uint16Err != nil {
				err = uint16Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint16Val
			} else {
				ret = uint16Val
			}
			return
		},
		model.TypePositiveInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint32Val, uint32Err := utils.ConvertToUint32(vVal)
			if uint32Err != nil {
				err = uint32Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint32Val
			} else {
				ret = uint32Val
			}
			return
		},
		model.TypePositiveBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint64Val, uint64Err := utils.ConvertToUint64(vVal)
			if uint64Err != nil {
				err = uint64Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint64Val
			} else {
				ret = uint64Val
			}
			return
		},
		model.TypePositiveIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uintVal, uintErr := utils.ConvertToUint(vVal)
			if uintErr != nil {
				err = uintErr
				return
			}
			if vType.IsPtrType() {
				ret = &uintVal
			} else {
				ret = uintVal
			}
			return
		},
		model.TypeFloatValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float32Val, float32Err := utils.ConvertToFloat32(vVal)
			if float32Err != nil {
				err = float32Err
				return
			}
			if vType.IsPtrType() {
				ret = &float32Val
			} else {
				ret = float32Val
			}
			return
		},
		model.TypeDoubleValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float64Val, float64Err := utils.ConvertToFloat64(vVal)
			if float64Err != nil {
				err = float64Err
				return
			}
			if vType.IsPtrType() {
				ret = &float64Val
			} else {
				ret = float64Val
			}
			return
		},
		model.TypeStringValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			strVal, strErr := utils.ConvertToString(vVal)
			if strErr != nil {
				err = strErr
				return
			}
			if vType.IsPtrType() {
				ret = &strVal
			} else {
				ret = strVal
			}
			return
		},
		model.TypeDateTimeValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			strVal, strErr := utils.ConvertToString(vVal)
			if strErr != nil {
				err = strErr
				return
			}
			if vType.IsPtrType() {
				ret = &strVal
			} else {
				ret = strVal
			}
			return
		},
		model.TypeSliceValue: decodeSliceValue,
	}

	decodeConvertSliceMap = ValueConvertMap{
		model.TypeBooleanValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			boolSlice, boolErr := decodeSliceTemplate(vVal, vType.Elem(), false)
			if boolErr != nil {
				err = boolErr
				return
			}
			if vType.IsPtrType() {
				ret = &boolSlice
			} else {
				ret = boolSlice
			}
			return
		},
		model.TypeBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int8Slice, int8Err := decodeSliceTemplate(vVal, vType.Elem(), int8(0))
			if int8Err != nil {
				err = int8Err
				return
			}
			if vType.IsPtrType() {
				ret = &int8Slice
			} else {
				ret = int8Slice
			}
			return
		},
		model.TypeSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int16Slice, int16Err := decodeSliceTemplate(vVal, vType.Elem(), int16(0))
			if int16Err != nil {
				err = int16Err
				return
			}
			if vType.IsPtrType() {
				ret = &int16Slice
			} else {
				ret = int16Slice
			}
			return
		},
		model.TypeInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int32Slice, int32Err := decodeSliceTemplate(vVal, vType.Elem(), int32(0))
			if int32Err != nil {
				err = int32Err
				return
			}
			if vType.IsPtrType() {
				ret = &int32Slice
			} else {
				ret = int32Slice
			}
			return
		},
		model.TypeBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			int64Slice, int64Err := decodeSliceTemplate(vVal, vType.Elem(), int64(0))
			if int64Err != nil {
				err = int64Err
				return
			}
			if vType.IsPtrType() {
				ret = &int64Slice
			} else {
				ret = int64Slice
			}
			return
		},
		model.TypeIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			intSlice, intErr := decodeSliceTemplate(vVal, vType.Elem(), int(0))
			if intErr != nil {
				err = intErr
				return
			}
			if vType.IsPtrType() {
				ret = &intSlice
			} else {
				ret = intSlice
			}
			return
		},
		model.TypePositiveBitValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint8Slice, uint8Err := decodeSliceTemplate(vVal, vType.Elem(), uint8(0))
			if uint8Err != nil {
				err = uint8Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint8Slice
			} else {
				ret = uint8Slice
			}
			return
		},
		model.TypePositiveSmallIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint16Slice, uint16Err := decodeSliceTemplate(vVal, vType.Elem(), uint16(0))
			if uint16Err != nil {
				err = uint16Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint16Slice
			} else {
				ret = uint16Slice
			}
			return
		},
		model.TypePositiveInteger32Value: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint32Slice, uint32Err := decodeSliceTemplate(vVal, vType.Elem(), uint32(0))
			if uint32Err != nil {
				err = uint32Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint32Slice
			} else {
				ret = uint32Slice
			}
			return
		},
		model.TypePositiveBigIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uint64Slice, uint64Err := decodeSliceTemplate(vVal, vType.Elem(), uint64(0))
			if uint64Err != nil {
				err = uint64Err
				return
			}
			if vType.IsPtrType() {
				ret = &uint64Slice
			} else {
				ret = uint64Slice
			}
			return
		},
		model.TypePositiveIntegerValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			uintSlice, uintErr := decodeSliceTemplate(vVal, vType.Elem(), uint(0))
			if uintErr != nil {
				err = uintErr
				return
			}
			if vType.IsPtrType() {
				ret = &uintSlice
			} else {
				ret = uintSlice
			}
			return
		},
		model.TypeFloatValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float32Slice, float32Err := decodeSliceTemplate(vVal, vType.Elem(), float32(0))
			if float32Err != nil {
				err = float32Err
				return
			}
			if vType.IsPtrType() {
				ret = &float32Slice
			} else {
				ret = float32Slice
			}
			return
		},
		model.TypeDoubleValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			float64Slice, float64Err := decodeSliceTemplate(vVal, vType.Elem(), float64(0))
			if float64Err != nil {
				err = float64Err
				return
			}
			if vType.IsPtrType() {
				ret = &float64Slice
			} else {
				ret = float64Slice
			}
			return
		},
		model.TypeStringValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			strSlice, strErr := decodeSliceTemplate(vVal, vType.Elem(), "")
			if strErr != nil {
				err = strErr
				return
			}
			if vType.IsPtrType() {
				ret = &strSlice
			} else {
				ret = strSlice
			}
			return
		},
		model.TypeDateTimeValue: func(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
			dateTimeSlice, dateTimeErr := decodeSliceTemplate(vVal, vType.Elem(), "")
			if dateTimeErr != nil {
				err = dateTimeErr
				return
			}
			if vType.IsPtrType() {
				ret = &dateTimeSlice
			} else {
				ret = dateTimeSlice
			}
			return
		},
	}
}

func EncodeValue(vVal any, vType model.Type) (ret any, err *cd.Result) {
	if !model.IsBasic(vType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("EncodeValue failed, illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	rVal := reflect.ValueOf(vVal)
	ret, err = encodeValue(rVal, vType)
	return
}

func encodeValue(rVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
	if !rVal.IsValid() {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal value, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	funcVal, funcOK := encodeValueConvertMap[vType.GetValue()]
	if !funcOK {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	ret, err = funcVal(rVal, vType)
	return
}

func encodeSliceValue(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
	if !model.IsBasic(vType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	eType := vType.Elem()
	funcVal, funcOK := encodeValueConvertSliceMap[eType.GetValue()]
	if !funcOK {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	ret, err = funcVal(vVal, vType)
	return
}

func encodeSliceTemplate[T any](vVal reflect.Value, vType model.Type, _ T) (ret []T, err *cd.Result) {
	rSliceValList, rSliceValErr := utils.ElemDependValue(vVal)
	if rSliceValErr != nil {
		err = rSliceValErr
		log.Errorf("encodeSliceTemplate failed, valErr:%v", rSliceValErr.Error())
		return
	}

	ret = []T{}
	for _, val := range rSliceValList {
		// 这里为了避免在使用[]any{},存储数据，通过这种方式取出实际的数据值
		val := reflect.ValueOf(val.Interface())
		encodeVal, encodeErr := encodeValue(val, vType)
		if encodeErr != nil {
			err = encodeErr
			log.Errorf("encodeSliceTemplate failed, encodeErr:%v", encodeErr.Error())
			return
		}

		tVal, tOk := encodeVal.(T)
		if !tOk {
			err = cd.NewResult(cd.UnExpected, "illegal type")
			log.Errorf("encodeSliceTemplate failed, illegal type")
			return
		}

		ret = append(ret, tVal)
	}

	return
}

func DecodeValue(vVal any, vType model.Type) (ret any, err *cd.Result) {
	if !model.IsBasic(vType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	rVal := reflect.ValueOf(vVal)
	ret, err = decodeValue(rVal, vType)
	return
}

func decodeValue(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
	if !vVal.IsValid() {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal value, type pkgKey:%s", vType.GetPkgKey()))
		return
	}
	funcPtr, funcOK := decodeValueConvertMap[vType.GetValue()]
	if !funcOK {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	ret, err = funcPtr(vVal, vType)
	return
}

func decodeSliceValue(vVal reflect.Value, vType model.Type) (ret any, err *cd.Result) {
	if !model.IsBasic(vType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	funcPtr, funcOK := decodeConvertSliceMap[vType.Elem().GetValue()]
	if !funcOK {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal type, type pkgKey:%s", vType.GetPkgKey()))
		return
	}

	ret, err = funcPtr(vVal, vType)
	return
}

func decodeSliceTemplate[T any](rSliceVal reflect.Value, vType model.Type, _ T) (ret []T, err *cd.Result) {
	ret = []T{}
	for idx := 0; idx < rSliceVal.Len(); idx++ {
		iVal := rSliceVal.Index(idx)
		decodeVal, decodeErr := DecodeValue(iVal.Interface(), vType.Elem())
		if decodeErr != nil {
			err = decodeErr
			log.Errorf("decodeSliceTemplate failed, decodeErr:%v", decodeErr.Error())
			return
		}

		tVal, tOk := decodeVal.(T)
		if !tOk {
			err = cd.NewResult(cd.UnExpected, "illegal type")
			log.Errorf("decodeSliceTemplate failed, illegal type, vType:%v, decodeVal:%+v", vType.GetPkgKey(), decodeVal)
			return
		}

		ret = append(ret, tVal)
	}

	return
}
