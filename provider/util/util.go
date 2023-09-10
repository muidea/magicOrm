package util

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	log "github.com/cihub/seelog"

	fu "github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
)

const (
	NameTag    = "Name"
	ValueTag   = "Value"
	PkgPathTag = "PkgPath"
	FieldsTag  = "Fields"
	ValuesTag  = "Values"
)

const (
	Key       = "key"
	Auto      = "auto"
	UUID      = "uuid"
	SnowFlake = "snowflake"
	DateTime  = "dateTime"
)

var snowFlakeNodePtr *fu.SnowFlakeNode
var snowFlakeOnce sync.Once

func init() {
	snowFlakeOnce.Do(func() {
		strNodeID := os.Getenv("node_id")
		if strNodeID == "" {
			strNodeID = "1"
		}
		nodeID, nodeErr := strconv.ParseInt(strNodeID, 10, 64)
		if nodeErr != nil {
			nodeID = 1
		}

		snowFlakeNodePtr, _ = fu.NewSnowFlakeNode(nodeID)
	})
}

func GetCurrentDateTime() (ret time.Time) {
	ret = time.Now().UTC()
	return
}

func GetNewUUID() (ret string) {
	ret = fu.NewUUID()
	return
}

func GetNewSnowFlakeID() (ret int64) {
	ret = snowFlakeNodePtr.Generate().Int64()
	return
}

func IsInteger(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		return true
	}

	return false
}

func IsUInteger(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		return true
	}

	return false
}

func IsFloat(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	}

	return false
}

func IsNumber(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}

	return false
}

func IsBool(tType reflect.Type) bool {
	return tType.Kind() == reflect.Bool
}

func IsString(tType reflect.Type) bool {
	return tType.Kind() == reflect.String
}

func IsDateTime(tType reflect.Type) bool {
	return tType.String() == "time.Time"
}

func IsSlice(tType reflect.Type) bool {
	return tType.Kind() == reflect.Slice
}

func IsStruct(tType reflect.Type) bool {
	if IsDateTime(tType) {
		return false
	}

	return tType.Kind() == reflect.Struct
}

func IsMap(tType reflect.Type) bool {
	return tType.Kind() == reflect.Map
}

// GetTypeEnum return field type as type constant from reflect.Value
func GetTypeEnum(val reflect.Type) (ret model.TypeDeclare, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ret = model.TypeBitValue
	case reflect.Uint8:
		ret = model.TypePositiveBitValue
	case reflect.Int16:
		ret = model.TypeSmallIntegerValue
	case reflect.Uint16:
		ret = model.TypePositiveSmallIntegerValue
	case reflect.Int32:
		ret = model.TypeInteger32Value
	case reflect.Uint32:
		ret = model.TypePositiveInteger32Value
	case reflect.Int64:
		ret = model.TypeBigIntegerValue
	case reflect.Uint64:
		ret = model.TypePositiveBigIntegerValue
	case reflect.Int:
		ret = model.TypeIntegerValue
	case reflect.Uint:
		ret = model.TypePositiveIntegerValue
	case reflect.Float32:
		ret = model.TypeFloatValue
	case reflect.Float64:
		ret = model.TypeDoubleValue
	case reflect.Bool:
		ret = model.TypeBooleanValue
	case reflect.String:
		ret = model.TypeStringValue
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ret = model.TypeDateTimeValue
		default:
			ret = model.TypeStructValue
		}
	case reflect.Slice:
		eType := val.Elem()
		if eType.Kind() == reflect.Ptr {
			eType = eType.Elem()
		}
		_, err = GetTypeEnum(eType)
		if err != nil {
			return
		}

		ret = model.TypeSliceValue
	default:
		err = fmt.Errorf("unsupported type:%v", val.String())
	}

	return
}

// IsNil check value if nil
func IsNil(val reflect.Value) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("check isNil failed, err:%v", err)
			ret = true
		}
	}()

	if !val.IsValid() {
		ret = true
		return
	}

	switch val.Kind() {
	case reflect.Invalid:
		ret = true
	case reflect.Interface, reflect.Slice, reflect.Map, reflect.Pointer:
		ret = val.IsNil()
	default:
		ret = false
	}

	return
}

func IsZero(val reflect.Value) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("check isZero failed, err:%v", err)
			ret = true
		}
	}()

	if IsNil(val) {
		ret = true
		return
	}

	ret = val.IsZero()
	return
}

// isSameStruct check if same
func isSameStruct(firstVal, secondVal reflect.Value) (ret bool, err error) {
	firstNum := firstVal.NumField()
	secondNum := secondVal.NumField()
	if firstNum != secondNum {
		ret = false
		return
	}

	for idx := 0; idx < firstNum; idx++ {
		firstField := firstVal.Field(idx)
		secondField := secondVal.Field(idx)
		ret, err = IsSameVal(firstField, secondField)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	ret = true
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
	typeVal, typeErr := GetTypeEnum(firstVal.Type())
	if typeErr != nil {
		err = typeErr
		ret = false
		return
	}

	if model.IsStructType(typeVal) {
		ret, err = isSameStruct(firstVal, secondVal)
		return
	}

	if model.IsBasicType(typeVal) {
		switch typeVal {
		case model.TypeBooleanValue:
			ret = firstVal.Bool() == secondVal.Bool()
		case model.TypeStringValue:
			ret = firstVal.String() == secondVal.String()
		case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
			ret = firstVal.Int() == secondVal.Int()
		case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
			ret = firstVal.Uint() == secondVal.Uint()
		case model.TypeFloatValue, model.TypeDoubleValue:
			ret = math.Abs(firstVal.Float()-secondVal.Float()) <= 0.0001
		case model.TypeDateTimeValue:
			ret = firstVal.Interface().(time.Time).Sub(secondVal.Interface().(time.Time)) == 0
		default:
			ret = false
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
