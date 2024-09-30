package remote

import (
	"fmt"
	"math"
	"reflect"
	"time"

	fu "github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/model"
)

type ValueImpl struct {
	value any
}

var NilValue = ValueImpl{}

var zeroString = ""
var zeroDateTime = time.Time{}.Format(fu.CSTLayout)

func NewValue(val any) (ret *ValueImpl) {
	if val == nil {
		return &NilValue
	}

	valPtr := &ValueImpl{}
	switch val.(type) {
	case bool,
		int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64,
		string,
		time.Time,
		[]bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]time.Time,
		*int8, *int16, *int32, *int, *int64,
		*uint8, *uint16, *uint32, *uint, *uint64,
		*float32, *float64,
		*string,
		*time.Time,
		*[]bool,
		*[]int8, *[]int16, *[]int32, *[]int, *[]int64,
		*[]uint8, *[]uint16, *[]uint32, *[]uint, *[]uint64,
		*[]float32, *[]float64,
		*[]string,
		*[]time.Time,
		//[]any,
		//*[]any,
		*ObjectValue, *SliceObjectValue:
		valPtr.value = val
	case ObjectValue, SliceObjectValue:
		valPtr.value = &val
	default:
		err := fmt.Errorf("illegal value, val:%v", val)
		panic(err.Error())
	}

	ret = valPtr
	return
}

func (s *ValueImpl) IsValid() (ret bool) {
	ret = s.value != nil
	return
}

func (s *ValueImpl) IsZero() (ret bool) {
	if !s.IsValid() {
		return true
	}

	rVal := reflect.ValueOf(s.value)
	switch s.value.(type) {
	case bool:
		return rVal.Bool() == false
	case int8, int16, int32, int, int64:
		return rVal.Int() == 0
	case uint8, uint16, uint32, uint, uint64:
		return rVal.Uint() == 0
	case float32, float64:
		return math.Float64bits(rVal.Float()) == 0
	case string:
		return rVal.String() == zeroString || rVal.String() == zeroDateTime
	case []bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]any:
		return rVal.Len() == 0
	case *ObjectValue:
		valuePtr, valueOK := s.value.(*ObjectValue)
		if valueOK {
			if valuePtr != nil {
				return !valuePtr.IsAssigned()
			}
			return true
		}

		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	case *SliceObjectValue:
		valuePtr, valueOK := s.value.(*SliceObjectValue)
		if valueOK {
			if valuePtr != nil {
				return !valuePtr.IsAssigned()
			}
			return true
		}
		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	default:
		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	}

	return true
}

func (s *ValueImpl) Set(val any) {
	s.value = val
	return
}

func (s *ValueImpl) Get() any {
	return s.value
}

func (s *ValueImpl) Addr() model.Value {
	impl := &ValueImpl{value: s.value}
	return impl
}

func (s *ValueImpl) Interface() model.RawVal {
	return model.NewRawVal(s.value)
}

func (s *ValueImpl) IsBasic() bool {
	if s.value == nil {
		return false
	}

	switch s.value.(type) {
	case bool,
		int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64,
		string,
		[]bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]any:
		return true
	case *ObjectValue, *SliceObjectValue:
		return false
	default:
		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	}

	return false
}

func (s *ValueImpl) Copy() (ret *ValueImpl) {
	if s.value == nil {
		ret = &ValueImpl{}
		return
	}

	ret = &ValueImpl{}
	switch s.value.(type) {
	case bool,
		int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64,
		string:
		ret.value = s.value
	case []bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]any,
		*ObjectValue,
		*SliceObjectValue:
		ret.value = s.value
	default:
		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	}

	return
}
