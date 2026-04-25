package remote

/*
Value 实现models.Value接口

value值对应的数值类型：
1. 基本数值类型：boolean, int8, int16, int32, int, int64, uint8, uint16, uint32, uint, uint64, float32, float64, string
2. 基本数值类型对应的指针类型：*boolean, *int8, *int16, *int32, *int, *int64, *uint8, *uint16, *uint32, *uint, *uint64, *float32, *float64, *string
3. 基本数值类型的slice: []boolean, []int8, []int16, []int32, []int, []int64, []uint8, []uint16, []uint32, []uint, []uint64, []float32, []float64, []string
4. 基本数值类型指针的slice: []*boolean, []*int8, []*int16, []*int32, []*int, []*int64, []*uint8, []*uint16, []*uint32, []*uint, []*uint64, []*float32, []*float64, []*string
4. ObjectValue: 对象类型，对应一个Object
5. *ObjectValue: 对象类型，对应一个Object
5. SliceObjectValue: 对象类型的slice，对应一个SliceObject
6. *SliceObjectValue: 对象类型的slice，对应一个SliceObject
*/

import (
	"fmt"
	"reflect"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type ValueImpl struct {
	value    any
	assigned bool
}

// NewValue 根据val创建Value
// 如果传入的val不合法，则直接panic
func NewValue(val any) (ret *ValueImpl) {
	valPtr := &ValueImpl{}
	if val == nil {
		ret = valPtr
		return
	}

	switch val.(type) {
	case *ObjectValue, *SliceObjectValue:
		valPtr.value = val
	case ObjectValue:
		objectVal := val.(ObjectValue)
		valPtr.value = &objectVal
	case SliceObjectValue:
		sliceObjectVal := val.(SliceObjectValue)
		valPtr.value = &sliceObjectVal
	default:
		if !isSupportedBasicValue(val) {
			panic(fmt.Sprintf("illegal value:%+v", val))
		}

		valPtr.value = val
	}
	valPtr.assigned = !isZero(valPtr.value)

	ret = valPtr
	return
}

// IsValid checks if the value is valid.
// 对于 ObjectValue / SliceObjectValue，非 nil 包装值本身即视为 valid；
// “未赋值 / 清空”的语义由 IsZero 进一步区分。
func (s *ValueImpl) IsValid() (ret bool) {
	return isValid(s.value)
}

func (s *ValueImpl) IsAssigned() bool {
	return s.assigned
}

// IsZero checks if the value is zero.
// 对于 SliceObjectValue，需要区分 nil（未赋值）和 []（显式赋值为空）。
func (s *ValueImpl) IsZero() bool {
	return isZero(s.value)
}

// Get 获取值
func (s *ValueImpl) Get() any {
	return s.value
}

// Set 设置值
// 如果传入的val不合法，则直接panic
func (s *ValueImpl) Set(val any) (err *cd.Error) {
	s.assigned = true
	if val == nil {
		s.value = nil
		return
	}

	switch val.(type) {
	case *ObjectValue:
		if s.value == nil {
			s.value = val
			return
		}
		err = rewriteObjectValue(s.value.(*ObjectValue), val.(*ObjectValue))
	case *SliceObjectValue:
		if s.value == nil {
			s.value = val
			return
		}
		err = rewriteSliceObjectValue(s.value.(*SliceObjectValue), val.(*SliceObjectValue))
	case ObjectValue:
		if s.value == nil {
			s.value = &val
			return
		}
		objectVal := val.(ObjectValue)
		err = rewriteObjectValue(s.value.(*ObjectValue), &objectVal)
	case SliceObjectValue:
		if s.value == nil {
			s.value = &val
			return
		}
		sliceObjectVal := val.(SliceObjectValue)
		err = rewriteSliceObjectValue(s.value.(*SliceObjectValue), &sliceObjectVal)
	default:
		if !isSupportedBasicValue(val) {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal value:%+v", val))
			return
		}

		s.value = val
	}
	return
}

// UnpackValue expands the contained value into individual elements.
// For slices, it returns the slice directly.
// For non-slice values, returns a single-element slice of the appropriate type.
func (s *ValueImpl) UnpackValue() (ret []models.Value) {
	if s.value == nil {
		return
	}

	rVal := reflect.ValueOf(s.value)
	if rVal.Kind() == reflect.Slice {
		for idx := 0; idx < rVal.Len(); idx++ {
			ret = append(ret, NewValue(rVal.Index(idx).Interface()))
		}
	} else {
		// For special types like ObjectValue or SliceObjectValue
		switch val := s.value.(type) {
		case *ObjectValue:
			ret = append(ret, NewValue(val))
		case ObjectValue:
			ret = append(ret, NewValue(&val))
		case *SliceObjectValue:
			for _, sv := range val.Values {
				ret = append(ret, NewValue(sv))
			}
		case SliceObjectValue:
			for _, sv := range val.Values {
				ret = append(ret, NewValue(sv))
			}
		default:
			ret = append(ret, NewValue(s.value))
		}
	}

	return
}

// Append appends the given value to the slice value.
// Append appends the given value to the slice value.
func (s *ValueImpl) Append(val any) (err *cd.Error) {
	if s.value == nil {
		err = cd.NewError(cd.Unexpected, "value is nil")
		slog.Warn("ValueImpl.Append: value is nil")
		return
	}

	switch s.value.(type) {
	case *SliceObjectValue:
		switch val.(type) {
		case *ObjectValue:
			if s.value.(*SliceObjectValue).GetPkgPath() != val.(*ObjectValue).GetPkgPath() {
				err = cd.NewError(cd.Unexpected, "pkgPath is not match")
				slog.Warn("ValueImpl.Append: pkgPath mismatch", "slice", s.value.(*SliceObjectValue).GetPkgPath(), "val", val.(*ObjectValue).GetPkgPath())
				return
			}
			s.value.(*SliceObjectValue).Values = append(s.value.(*SliceObjectValue).Values, val.(*ObjectValue))
		default:
			err = cd.NewError(cd.Unexpected, "value is not ObjectValue")
		}
	default:
		s.value, err = appendBasicValue(s.value, val)
	}

	return
}

// Copy 复制一个新值
// 要求进行深度copy
func (s *ValueImpl) copy() (ret *ValueImpl, err error) {
	if s.value == nil {
		ret = &ValueImpl{}
		return
	}

	ret = &ValueImpl{assigned: s.assigned}
	switch s.value.(type) {
	case *ObjectValue:
		ret.value = s.value.(*ObjectValue).Copy()
	case *SliceObjectValue:
		ret.value = s.value.(*SliceObjectValue).Copy()
	default:
		copiedVal, copiedErr := utils.DeepCopy(s.value)
		if copiedErr != nil {
			err = copiedErr
			slog.Error("ValueImpl.copy DeepCopy failed", "error", copiedErr.Error())
			return
		}
		ret.value = copiedVal
	}

	return
}
