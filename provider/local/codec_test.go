package local

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/utils"
	"github.com/stretchr/testify/assert"
)

func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"encode bool", true, int8(1)},
		{"encode int8", int8(8), int8(8)},
		{"encode int16", int16(16), int16(16)},
		{"encode int32", int32(32), int32(32)},
		{"encode int64", int64(64), int64(64)},
		{"encode int", 42, int(42)},
		{"encode uint8", uint8(8), uint8(8)},
		{"encode uint16", uint16(16), uint16(16)},
		{"encode uint32", uint32(32), uint32(32)},
		{"encode uint64", uint64(64), uint64(64)},
		{"encode uint", uint(42), uint(42)},
		{"encode float32", float32(3.14), float32(3.14)},
		{"encode float64", 3.14, float64(3.14)},
		{"encode []bool", []bool{true, false, true}, []int8{1, 0, 1}},
		{"encode []int", []int{1, 2, 3}, []int{1, 2, 3}},
		{"encode []float64", []float64{1.23, 2.34, 3.45}, []float64{1.23, 2.34, 3.45}},
		{"encode []string", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vType, err := NewType(reflect.TypeOf(tt.input))
			assert.Nil(t, err)

			result, err := EncodeValue(tt.input, vType)
			assert.Nil(t, err)
			if !utils.IsSameValue(tt.expected, result) {
				t.Errorf("name: %s, expected: %v, got: %v", tt.name, tt.expected, result)
			}
		})
	}
}
