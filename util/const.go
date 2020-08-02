package util

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// Define the Type enum
const (
	// bool
	TypeBooleanField = 1 << iota
	// string
	TypeStringField
	// time.Time
	TypeDateTimeField
	// int8
	TypeBitField
	// int16
	TypeSmallIntegerField
	// int32
	TypeInteger32Field
	// int
	TypeIntegerField
	// int64
	TypeBigIntegerField
	// uint8
	TypePositiveBitField
	// uint16
	TypePositiveSmallIntegerField
	// uint32
	TypePositiveInteger32Field
	// uint
	TypePositiveIntegerField
	// uint64
	TypePositiveBigIntegerField
	// float32
	TypeFloatField
	// float64
	TypeDoubleField
	// struct
	TypeStructField
	// slice
	TypeSliceField
)

// IsBasicType IsBasicType
func IsBasicType(typeValue int) bool {
	return typeValue < TypeStructField
}

// IsStructType IsStructType
func IsStructType(typeValue int) bool {
	return typeValue == TypeStructField
}

// IsSliceType IsSliceType
func IsSliceType(typeValue int) bool {
	return typeValue == TypeSliceField
}

// GetBasicTypeInitValue GetBasicTypeInitValue
func GetBasicTypeInitValue(typeValue int) (ret interface{}, err error) {
	switch typeValue {
	case TypeBooleanField,
		TypeBitField, TypeSmallIntegerField, TypeIntegerField, TypeInteger32Field, TypeBigIntegerField:
		val := int64(0)
		ret = &val
		break
	case TypePositiveBitField, TypePositiveSmallIntegerField, TypePositiveIntegerField, TypePositiveInteger32Field, TypePositiveBigIntegerField:
		val := uint64(0)
		ret = &val
		break
	case TypeStringField, TypeDateTimeField:
		val := ""
		ret = &val
		break
	case TypeFloatField, TypeDoubleField:
		val := 0.00
		ret = &val
		break
	case TypeStructField:
		val := 0
		ret = &val
	case TypeSliceField:
		val := ""
		ret = &val
	default:
		err = fmt.Errorf("no support fileType, %d", typeValue)
	}

	return
}

// GetTypeValueEnum return field type as type constant from reflect.Value
func GetTypeValueEnum(val reflect.Type) (ret int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ret = TypeBitField
	case reflect.Uint8:
		ret = TypePositiveBitField
	case reflect.Int16:
		ret = TypeSmallIntegerField
	case reflect.Uint16:
		ret = TypePositiveSmallIntegerField
	case reflect.Int32:
		ret = TypeInteger32Field
	case reflect.Uint32:
		ret = TypePositiveInteger32Field
	case reflect.Int64:
		ret = TypeBigIntegerField
	case reflect.Uint64:
		ret = TypePositiveBigIntegerField
	case reflect.Int:
		ret = TypeIntegerField
	case reflect.Uint:
		ret = TypePositiveIntegerField
	case reflect.Float32:
		ret = TypeFloatField
	case reflect.Float64:
		ret = TypeDoubleField
	case reflect.Bool:
		ret = TypeBooleanField
	case reflect.String:
		ret = TypeStringField
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ret = TypeDateTimeField
		default:
			ret = TypeStructField
		}
	case reflect.Slice:
		ret = TypeSliceField
	default:
		err = fmt.Errorf("unsupport field type:%v", val.String())
	}

	return
}

// GetSliceRawTypeEnum get slice rawType
func GetSliceRawTypeEnum(sliceType reflect.Type) (ret int, err error) {
	if sliceType.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal type, not slice. typeVal:%s", sliceType.Kind().String())
		return
	}

	rawType := sliceType.Elem()
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	ret, err = GetTypeValueEnum(rawType)
	if err != nil {
		return
	}

	return
}

// IsNil check value if nil
func IsNil(val reflect.Value) (ret bool) {
	if val.Kind() == reflect.Ptr {
		ret = val.IsNil()
		return
	}

	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Interface:
		ret = val.IsNil()
	case reflect.Slice, reflect.Map:
		ret = false
	case reflect.Invalid:
		ret = true
	default:
		ret = false
	}

	return
}

//IsSame check if same
func IsSame(firstVal, secondVal reflect.Value) (ret bool, err error) {
	ret = firstVal.Type().String() == secondVal.Type().String()
	if !ret {
		return
	}

	firstIsNil := IsNil(firstVal)
	secondIsNul := IsNil(secondVal)
	if firstIsNil != secondIsNul {
		ret = false
		return
	}
	if firstIsNil {
		ret = true
		return
	}

	firstVal = reflect.Indirect(firstVal)
	secondVal = reflect.Indirect(secondVal)
	typeVal, typeErr := GetTypeValueEnum(firstVal.Type())
	if typeErr != nil {
		ret = false
		err = typeErr
		return
	}
	if typeVal != TypeStructField {
		ret = false
		err = fmt.Errorf("illegal value, type:%s", firstVal.Type().String())
		return
	}

	numField := firstVal.NumField()
	for idx := 0; idx < numField; idx++ {
		firstField := firstVal.Field(idx)
		secondField := secondVal.Field(idx)
		ret, err = IsSameVal(firstField, secondField)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	return
}

// IsSameVal is same value
func IsSameVal(firstVal, secondVal reflect.Value) (ret bool, err error) {
	ret = firstVal.Type().String() == secondVal.Type().String()
	if !ret {
		return
	}

	firstIsNil := IsNil(firstVal)
	secondIsNil := IsNil(secondVal)
	if firstIsNil != secondIsNil {
		ret = false
		return
	}
	if firstIsNil {
		ret = true
		return
	}
	firstVal = reflect.Indirect(firstVal)
	secondVal = reflect.Indirect(secondVal)
	typeVal, typeErr := GetTypeValueEnum(firstVal.Type())
	if typeErr != nil {
		err = typeErr
		ret = false
		return
	}

	if IsStructType(typeVal) {
		ret, err = IsSame(firstVal, secondVal)
		return
	}

	firstVal = reflect.Indirect(firstVal)
	secondVal = reflect.Indirect(secondVal)
	if IsBasicType(typeVal) {
		switch typeVal {
		case TypeBooleanField:
			ret = firstVal.Bool() == secondVal.Bool()
		case TypeStringField:
			ret = firstVal.String() == secondVal.String()
		case TypeBitField, TypeSmallIntegerField, TypeInteger32Field, TypeIntegerField, TypeBigIntegerField:
			ret = firstVal.Int() == secondVal.Int()
		case TypePositiveBitField, TypePositiveSmallIntegerField, TypePositiveInteger32Field, TypePositiveIntegerField, TypePositiveBigIntegerField:
			ret = firstVal.Uint() == secondVal.Uint()
		case TypeFloatField, TypeDoubleField:
			ret = math.Abs(firstVal.Float()-secondVal.Float()) <= 0.0001
		case TypeDateTimeField:
			ret = firstVal.Interface().(time.Time).Sub(secondVal.Interface().(time.Time)) == 0
		default:
			ret = false
			err = fmt.Errorf("illegal value, is a struct value")
		}

		return
	}

	ret = firstVal.Len() == secondVal.Len()
	if !ret {
		return
	}

	for idx := 0; idx < firstVal.Len(); idx++ {
		firstItem := firstVal.Index(idx)
		secondItem := secondVal.Index(idx)
		ret, err = IsSameVal(firstItem, secondItem)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	return
}
