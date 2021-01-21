package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func getSliceType(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	switch eType.GetValue() {
	case util.TypeBooleanField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*bool
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]bool
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*bool
			ret = reflect.TypeOf(val)
			return
		}
		var val []bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField,
		util.TypeDateTimeField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*string
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]string
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*string
			ret = reflect.TypeOf(val)
			return
		}
		var val []string
		ret = reflect.TypeOf(val)
	case util.TypeBitField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*int8
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]int8
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*int8
			ret = reflect.TypeOf(val)
			return
		}
		var val []int8
		ret = reflect.TypeOf(val)
	case util.TypeSmallIntegerField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*int16
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]int16
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*int16
			ret = reflect.TypeOf(val)
			return
		}
		var val []int16
		ret = reflect.TypeOf(val)
	case util.TypeInteger32Field:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*int32
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]int32
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*int32
			ret = reflect.TypeOf(val)
			return
		}
		var val []int32
		ret = reflect.TypeOf(val)
	case util.TypeIntegerField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*int
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]int
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*int
			ret = reflect.TypeOf(val)
			return
		}
		var val []int
		ret = reflect.TypeOf(val)
	case util.TypeBigIntegerField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*int64
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]int64
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*int64
			ret = reflect.TypeOf(val)
			return
		}
		var val []int64
		ret = reflect.TypeOf(val)
	case util.TypePositiveBitField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*uint8
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]uint8
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*uint8
			ret = reflect.TypeOf(val)
			return
		}
		var val []uint8
		ret = reflect.TypeOf(val)
	case util.TypePositiveSmallIntegerField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*uint16
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]uint16
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*uint16
			ret = reflect.TypeOf(val)
			return
		}
		var val []uint16
		ret = reflect.TypeOf(val)
	case util.TypePositiveInteger32Field:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*uint32
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]uint32
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*uint32
			ret = reflect.TypeOf(val)
			return
		}
		var val []uint32
		ret = reflect.TypeOf(val)
	case util.TypePositiveIntegerField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*uint
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]uint
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*uint
			ret = reflect.TypeOf(val)
			return
		}
		var val []uint
		ret = reflect.TypeOf(val)
	case util.TypePositiveBigIntegerField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*uint64
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]uint64
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*uint64
			ret = reflect.TypeOf(val)
			return
		}
		var val []uint64
		ret = reflect.TypeOf(val)
	case util.TypeFloatField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*float32
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]float32
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*float32
			ret = reflect.TypeOf(val)
			return
		}
		var val []float32
		ret = reflect.TypeOf(val)
	case util.TypeDoubleField:
		if tType.IsPtrType() {
			if eType.IsPtrType() {
				var val *[]*float64
				ret = reflect.TypeOf(val)
				return
			}
			var val *[]float64
			ret = reflect.TypeOf(val)
			return
		}
		if eType.IsPtrType() {
			var val []*float64
			ret = reflect.TypeOf(val)
			return
		}
		var val []float64
		ret = reflect.TypeOf(val)
	default:
		err = fmt.Errorf("unexpect slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}

func GetType(tType model.Type) (ret reflect.Type, err error) {
	switch tType.GetValue() {
	case util.TypeBooleanField:
		if tType.IsPtrType() {
			var val *bool
			ret = reflect.TypeOf(val)
			return
		}
		var val bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField,
		util.TypeDateTimeField:
		if tType.IsPtrType() {
			var val *string
			ret = reflect.TypeOf(val)
			return
		}
		var val string
		ret = reflect.TypeOf(val)
	case util.TypeBitField:
		if tType.IsPtrType() {
			var val *int8
			ret = reflect.TypeOf(val)
			return
		}
		var val int8
		ret = reflect.TypeOf(val)
	case util.TypeSmallIntegerField:
		if tType.IsPtrType() {
			var val *int16
			ret = reflect.TypeOf(val)
			return
		}
		var val int16
		ret = reflect.TypeOf(val)
	case util.TypeInteger32Field:
		if tType.IsPtrType() {
			var val *int32
			ret = reflect.TypeOf(val)
			return
		}
		var val int32
		ret = reflect.TypeOf(val)
	case util.TypeIntegerField:
		if tType.IsPtrType() {
			var val *int
			ret = reflect.TypeOf(val)
			return
		}
		var val int
		ret = reflect.TypeOf(val)
	case util.TypeBigIntegerField:
		if tType.IsPtrType() {
			var val *int64
			ret = reflect.TypeOf(val)
			return
		}
		var val int64
		ret = reflect.TypeOf(val)
	case util.TypePositiveBitField:
		if tType.IsPtrType() {
			var val *uint8
			ret = reflect.TypeOf(val)
			return
		}
		var val uint8
		ret = reflect.TypeOf(val)
	case util.TypePositiveSmallIntegerField:
		if tType.IsPtrType() {
			var val *uint16
			ret = reflect.TypeOf(val)
			return
		}
		var val uint16
		ret = reflect.TypeOf(val)
	case util.TypePositiveInteger32Field:
		if tType.IsPtrType() {
			var val *uint32
			ret = reflect.TypeOf(val)
			return
		}
		var val uint32
		ret = reflect.TypeOf(val)
	case util.TypePositiveIntegerField:
		if tType.IsPtrType() {
			var val *uint
			ret = reflect.TypeOf(val)
			return
		}
		var val uint
		ret = reflect.TypeOf(val)
	case util.TypePositiveBigIntegerField:
		if tType.IsPtrType() {
			var val *uint64
			ret = reflect.TypeOf(val)
			return
		}
		var val uint64
		ret = reflect.TypeOf(val)
	case util.TypeFloatField:
		if tType.IsPtrType() {
			var val *float32
			ret = reflect.TypeOf(val)
			return
		}
		var val float32
		ret = reflect.TypeOf(val)
	case util.TypeDoubleField:
		if tType.IsPtrType() {
			var val *float64
			ret = reflect.TypeOf(val)
		}
		var val float64
		ret = reflect.TypeOf(val)
	case util.TypeSliceField:
		ret, err = getSliceType(tType)
	default:
		err = fmt.Errorf("unexpect item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}
