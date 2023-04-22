package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

var _declareObjectSliceValue SliceObjectValue
var _declareObjectValue ObjectValue

func getBoolSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
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
	return
}

func getStringSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
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
	return
}

func getNumberSlice(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
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
	return
}

func getSliceType(tType model.Type) (ret reflect.Type, err error) {
	eType := tType.Elem()
	switch eType.GetValue() {
	case util.TypeBooleanField:
		ret, err = getBoolSlice(tType)
	case util.TypeStringField,
		util.TypeDateTimeField:
		ret, err = getStringSlice(tType)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField,
		util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField,
		util.TypeFloatField, util.TypeDoubleField:
		ret, err = getNumberSlice(tType)
	case util.TypeStructField:
		ret = reflect.TypeOf(&_declareObjectSliceValue)
	default:
		err = fmt.Errorf("unexpect slice item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}

func getType(tType model.Type) (ret reflect.Type, err error) {
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
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField,
		util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField,
		util.TypeFloatField, util.TypeDoubleField:
		if tType.IsPtrType() {
			var val *float64
			ret = reflect.TypeOf(val)
			return
		}
		var val float64
		ret = reflect.TypeOf(val)
	case util.TypeStructField:
		ret = reflect.TypeOf(&_declareObjectValue)
	case util.TypeSliceField:
		ret, err = getSliceType(tType)
	default:
		err = fmt.Errorf("unexpect item type, name:%s, type:%d", tType.GetName(), tType.GetValue())
	}

	return
}

func getInitializeValue(tType model.Type) (ret reflect.Value) {
	cType, _ := getType(tType)
	if tType.IsPtrType() || !tType.IsBasic() {
		cType = cType.Elem()
	}

	cValue := reflect.New(cType).Elem()
	if !tType.IsBasic() {
		cValue.FieldByName("Name").SetString(tType.GetName())
		cValue.FieldByName("PkgPath").SetString(tType.GetPkgPath())
		cValue.FieldByName("IsPtr").SetBool(tType.IsPtrType())
		if util.IsSliceType(tType.GetValue()) {
			cValue.FieldByName("IsElemPtr").SetBool(tType.Elem().IsPtrType())
		}
	}

	if tType.IsPtrType() || !tType.IsBasic() {
		cValue = cValue.Addr()
	}

	ret = cValue
	return
}
