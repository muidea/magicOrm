package remote

/*
Value 实现model.Value接口

value值对应的数值类型：
1. 基本数值类型：bool, int8, int16, int32, int, int64, uint8, uint16, uint32, uint, uint64, float32, float64, string
2. 基本数值类型对应的指针类型：*bool, *int8, *int16, *int32, *int, *int64, *uint8, *uint16, *uint32, *uint, *uint64, *float32, *float64, *string
3. 基本数值类型的slice: []bool, []int8, []int16, []int32, []int, []int64, []uint8, []uint16, []uint32, []uint, []uint64, []float32, []float64, []string
4. 基本数值类型指针的slice: []*bool, []*int8, []*int16, []*int32, []*int, []*int64, []*uint8, []*uint16, []*uint32, []*uint, []*uint64, []*float32, []*float64, []*string
4. ObjectValue: 对象类型，对应一个Object
5. *ObjectValue: 对象类型，对应一个Object
5. SliceObjectValue: 对象类型的slice，对应一个SliceObject
6. *SliceObjectValue: 对象类型的slice，对应一个SliceObject
*/

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
)

type ValueImpl struct {
	value any
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
	case ObjectValue, SliceObjectValue:
		valPtr.value = &val
	default:
		if !utils.IsReallyValidValue(val) {
			panic(fmt.Sprintf("illegal value:%+v", val))
		}

		valPtr.value = val
	}

	ret = valPtr
	return
}

// IsValid checks if the value is valid
// 如果对应的值是ObjectValue,SliceObjectValue或者对应的指针值，还需要继续判断是否包含Fields，Fields的包含的items为0也认为是invalid
func (s *ValueImpl) IsValid() (ret bool) {
	if s.value == nil {
		return false
	}

	// 不用继续检查，在赋值时已经做过校验
	return true
}

// IsZero checks if the value is zero.
// 如果对应的值是ObjectValue,SliceObjectValue或者对应的指针值，还需要继续判断是否包含Fields，Fields的包含的items为0也认为是0
func (s *ValueImpl) IsZero() bool {
	if s.value == nil {
		return true
	}

	switch v := s.value.(type) {
	case *ObjectValue:
		return v == nil || len(v.Fields) == 0 || !v.IsAssigned()
	case *SliceObjectValue:
		return v == nil || len(v.Values) == 0
	case ObjectValue:
		return len(v.Fields) == 0 || !v.IsAssigned()
	case SliceObjectValue:
		return len(v.Values) == 0
	default:
		return utils.IsReallyZeroValue(s.value)
	}
}

// Get 获取值
func (s *ValueImpl) Get() any {
	return s.value
}

// Set 设置值
// 如果传入的val不合法，则直接panic
func (s *ValueImpl) Set(val any) (err *cd.Error) {
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
		if !utils.IsReallyValidValue(val) {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal value:%+v", val))
			return
		}

		s.value = val
	}
	return
}

// UnpackValue expands the contained value into individual elements.
// For slices, it returns the slice directly.
// For non-slice values, returns a single-element slice of the appropriate type.
func (s *ValueImpl) UnpackValue() (ret []model.Value) {
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
		err = cd.NewError(cd.UnExpected, "value is nil")
		log.Warnf("Append failed, value is nil")
		return
	}

	switch s.value.(type) {
	case *SliceObjectValue:
		switch val.(type) {
		case *ObjectValue:
			if s.value.(*SliceObjectValue).GetPkgPath() != val.(*ObjectValue).GetPkgPath() {
				err = cd.NewError(cd.UnExpected, "pkgPath is not match")
				log.Warnf("Append failed, pkgPath is not match")
				return
			}
			s.value.(*SliceObjectValue).Values = append(s.value.(*SliceObjectValue).Values, val.(*ObjectValue))
		default:
			err = cd.NewError(cd.UnExpected, "value is not ObjectValue")
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

	ret = &ValueImpl{}
	switch s.value.(type) {
	case *ObjectValue:
		ret.value = s.value.(*ObjectValue).Copy()
	case *SliceObjectValue:
		ret.value = s.value.(*SliceObjectValue).Copy()
	default:
		copiedVal, copiedErr := utils.DeepCopy(s.value)
		if copiedErr != nil {
			err = copiedErr
			log.Errorf("copy failed, copiedErr:%v", copiedErr.Error())
			return
		}
		ret.value = copiedVal
	}

	return
}
