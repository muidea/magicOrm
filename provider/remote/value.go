package remote

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
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
		time.Time,
		[]bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]time.Time,
		[]any,
		*ObjectValue, *SliceObjectValue:
		valPtr.value = val
	case ObjectValue, SliceObjectValue:
		valPtr.value = &val
	default:
		rVal := reflect.ValueOf(val)
		err := fmt.Errorf("illegal value, val:%v, val type:%s", val, rVal.Type().String())
		panic(err.Error())
	}

	ret = valPtr
	return
}

func (s *ValueImpl) IsValid() (ret bool) {
	ret = utils.IsReallyValid(s.value)
	return
}

// IsZero checks if the value is zero.
// 如果对应的值是ObjectValue,SliceObjectValue或者对应的指针值，还需要继续判断是否包含Fields，Fields的包含的items为0也认为是0
func (s *ValueImpl) IsZero() bool {
	if s.value == nil {
		return true
	}

	switch v := s.value.(type) {
	case *ObjectValue:
		return v == nil || len(v.Fields) == 0
	case *SliceObjectValue:
		return v == nil || len(v.Values) == 0
	case ObjectValue:
		return len(v.Fields) == 0
	case SliceObjectValue:
		return len(v.Values) == 0
	default:
		return utils.IsReallyZero(s.value)
	}
}

func (s *ValueImpl) Set(val any) {
	s.value = val
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
}

func (s *ValueImpl) Copy() (ret *ValueImpl, err error) {
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
		err = fmt.Errorf("illegal value type, val:%+v", s.value)
	}

	return
}
