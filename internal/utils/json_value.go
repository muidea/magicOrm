package utils

import "reflect"

// BasicValueForJSON projects a model's uint8 collection as numbers. This is
// only for declared model values; byte streams elsewhere keep normal JSON
// encoding. Preserve nil, empty, pointer and in-process typed-value semantics.
func BasicValueForJSON(value any) any {
	switch v := value.(type) {
	case nil, bool, string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		[]string, []int, []int64, []uint16, []float64, []bool, []any:
		return value
	case []byte:
		if v == nil {
			return nil
		}
		numbers := make([]uint16, len(v))
		for i, item := range v {
			numbers[i] = uint16(item)
		}
		return numbers
	}

	v := reflect.ValueOf(value)
	for v.IsValid() && v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Slice || v.Type().Elem().Kind() != reflect.Uint8 {
		return value
	}
	if v.IsNil() {
		return nil
	}
	numbers := make([]uint16, v.Len())
	for i := range numbers {
		numbers[i] = uint16(v.Index(i).Uint())
	}
	return numbers
}
