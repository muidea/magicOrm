package remote

import (
	"fmt"
	"math"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

type ValueImpl struct {
	value any
}

var NilValue = ValueImpl{}

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
		[]bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		*ObjectValue, *SliceObjectValue:
		valPtr.value = val
	default:
		err := fmt.Errorf("illegal value, val:%v", val)
		panic(err.Error())
	}

	ret = valPtr
	return
}

func (s *ValueImpl) IsNil() (ret bool) {
	ret = s.value == nil
	return
}

func (s *ValueImpl) IsZero() (ret bool) {
	if s.IsNil() {
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
		return rVal.String() == ""
	case []bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]any:
		return rVal.Len() == 0
	case *ObjectValue:
		return !s.value.(*ObjectValue).IsAssigned()
	case *SliceObjectValue:
		return !s.value.(*SliceObjectValue).IsAssigned()
	default:
		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	}

	return true
}

func (s *ValueImpl) Set(val any) (err error) {
	if val == nil {
		s.value = nil
		return
	}

	switch val.(type) {
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
		[]any,
		*ObjectValue, *SliceObjectValue:
		s.value = val
	default:
		err := fmt.Errorf("illegal value, val:%v", val)
		panic(err.Error())
	}

	return
}

func (s *ValueImpl) Get() any {
	return s.value
}

func (s *ValueImpl) Addr() model.Value {
	impl := &ValueImpl{value: s.value}
	return impl
}

func (s *ValueImpl) Interface() any {
	return s.value
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
		[]string:
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
		[]string:
		ret.value = s.value
	case *ObjectValue:
	case *SliceObjectValue:
	default:
		err := fmt.Errorf("illegal value, val:%v", s.value)
		panic(err.Error())
	}

	return
}
