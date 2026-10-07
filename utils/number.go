package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"

	cd "github.com/muidea/magicCommon/def"
)

func integerBits(kind reflect.Kind) int {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return 8
	case reflect.Int16, reflect.Uint16:
		return 16
	case reflect.Int32, reflect.Uint32:
		return 32
	case reflect.Int64, reflect.Uint64:
		return 64
	case reflect.Int, reflect.Uint:
		return strconv.IntSize
	}
	return 0
}

func numberRangeError(kind reflect.Kind, value any) *cd.Error {
	return cd.NewError(cd.IllegalParam, fmt.Sprintf("numeric value %v is outside %v range", value, kind))
}

func checkSignedNumberRange(kind reflect.Kind, value int64) *cd.Error {
	bits := integerBits(kind)
	if integerKindMap[kind] && bits < 64 && (value < -(int64(1)<<(bits-1)) || value >= int64(1)<<(bits-1)) {
		return numberRangeError(kind, value)
	}
	if uintegerKindMap[kind] && (value < 0 || (bits < 64 && uint64(value) >= uint64(1)<<bits)) {
		return numberRangeError(kind, value)
	}
	return nil
}

func checkUnsignedNumberRange(kind reflect.Kind, value uint64) *cd.Error {
	bits := integerBits(kind)
	if integerKindMap[kind] && value >= uint64(1)<<(bits-1) {
		return numberRangeError(kind, value)
	}
	if uintegerKindMap[kind] && bits < 64 && value >= uint64(1)<<bits {
		return numberRangeError(kind, value)
	}
	return nil
}

func checkFloatNumberRange(kind reflect.Kind, value float64) *cd.Error {
	if math.IsNaN(value) || math.IsInf(value, 0) || (kind == reflect.Float32 && math.Abs(value) > math.MaxFloat32) {
		return numberRangeError(kind, value)
	}
	// Preserve the existing truncation of fractions, but never wrap overflow.
	truncated := math.Trunc(value)
	bits := integerBits(kind)
	if integerKindMap[kind] && (truncated < -math.Ldexp(1, bits-1) || truncated >= math.Ldexp(1, bits-1)) {
		return numberRangeError(kind, value)
	}
	if uintegerKindMap[kind] && (truncated < 0 || truncated >= math.Ldexp(1, bits)) {
		return numberRangeError(kind, value)
	}
	return nil
}

func convertJSONNumber(kind reflect.Kind, number json.Number) (any, *cd.Error) {
	if floatKindMap[kind] {
		return convertStringToNumber(kind, number.String())
	}
	if integerKindMap[kind] {
		if value, err := number.Int64(); err == nil {
			return convertIntToNumber(kind, value)
		}
	} else if uintegerKindMap[kind] {
		if value, err := strconv.ParseUint(number.String(), 10, 64); err == nil {
			return convertUintToNumber(kind, value)
		}
	}
	// Decimal and exponent tokens use exact arithmetic before truncation; a
	// float64 intermediate would change large integers even within int64 range.
	approx, err := number.Float64()
	if err != nil || math.IsNaN(approx) || math.IsInf(approx, 0) {
		return nil, numberRangeError(kind, number)
	}
	if approx == 0 {
		return convertIntToNumber(kind, 0)
	}
	if err := checkFloatNumberRange(kind, approx); err != nil {
		// The rounded upper boundary may still represent a valid exact integer.
		if math.Abs(approx) > math.Ldexp(1, 64) {
			return nil, err
		}
	}
	exact, ok := new(big.Rat).SetString(number.String())
	if !ok {
		return nil, cd.NewError(cd.IllegalParam, "invalid JSON numeric value")
	}
	integer := new(big.Int).Quo(exact.Num(), exact.Denom())
	if integerKindMap[kind] && integer.IsInt64() {
		return convertIntToNumber(kind, integer.Int64())
	}
	if uintegerKindMap[kind] && integer.IsUint64() {
		return convertUintToNumber(kind, integer.Uint64())
	}
	return nil, numberRangeError(kind, number)
}
