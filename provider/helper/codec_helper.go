// Package helper provides shared utilities for provider implementations
package helper

import (
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
	"log/slog"
)

// EncodeSliceTemplate is a generic helper function for encoding slice values
// This eliminates code duplication between local and remote providers
func EncodeSliceTemplate[T any](
	encodeValueFunc func(reflect.Value, models.Type) (any, *cd.Error),
	vVal reflect.Value,
	vType models.Type,
	_ T,
) (ret []T, err *cd.Error) {
	rSliceValList, rSliceValErr := utils.ElemDependValue(vVal)
	if rSliceValErr != nil {
		err = rSliceValErr
		slog.Error("encodeSliceTemplate failed", "error", err.Error())
		return
	}

	ret = make([]T, 0, len(rSliceValList))
	for _, val := range rSliceValList {
		encodeVal, encodeErr := encodeValueFunc(val, vType)
		if encodeErr != nil {
			err = encodeErr
			slog.Error("encodeSliceTemplate failed", "error", err.Error())
			return
		}

		tVal, tOk := encodeVal.(T)
		if !tOk {
			err = cd.NewError(cd.Unexpected, "encode value type mismatch")
			slog.Error("EncodeSliceTemplate type assertion failed", "error", err.Error())
			return
		}

		ret = append(ret, tVal)
	}

	return
}

// DecodeSliceValue is a generic helper function for decoding slice values
func DecodeSliceValue(
	decodeValueFunc func(any, models.Type) (any, *cd.Error),
	vVal reflect.Value,
	vType models.Type,
) (ret any, err *cd.Error) {
	if vVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "value is not slice")
		slog.Error("decodeSliceValue failed", "error", err.Error())
		return
	}

	sliceLen := vVal.Len()
	sliceVal := make([]any, sliceLen)
	for i := 0; i < sliceLen; i++ {
		val := vVal.Index(i).Interface()
		decodeVal, decodeErr := decodeValueFunc(val, vType)
		if decodeErr != nil {
			err = decodeErr
			slog.Error("decodeSliceValue failed", "error", err.Error())
			return
		}

		sliceVal[i] = decodeVal
	}

	ret = sliceVal
	return
}
