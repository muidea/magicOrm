package util

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	fu "github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
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

type Pagination struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
}

func (s *Pagination) Limit() int64 {
	if s.PageNum < 1 {
		s.PageNum = 1
	}

	if s.PageSize < 1 {
		s.PageSize = 10
	}

	return int64(s.PageNum) * int64(s.PageSize)
}

func (s *Pagination) Offset() int64 {
	if s.PageNum < 1 {
		s.PageNum = 1
	}

	if s.PageSize < 1 {
		s.PageSize = 10
	}

	return int64(s.PageNum-1) * int64(s.PageSize)
}

type SortFilter struct {
	// true:升序,false:降序
	AscFlag bool `json:"ascFlag"`
	// 排序字段
	FieldName string `json:"fieldName"`
}

func (s *SortFilter) Name() string {
	return s.FieldName
}

func (s *SortFilter) AscSort() bool {
	return s.AscFlag
}

func GetCurrentDateTime() (ret time.Time) {
	ret = time.Now().UTC()
	return
}

func GetCurrentDateTimeStr() (ret string) {
	ret = time.Now().UTC().Format(fu.CSTLayout)
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

func IsPtr(tType reflect.Type) bool {
	return tType.Kind() == reflect.Ptr
}

func GetTypeEnum(val reflect.Type) (ret model.TypeDeclare, err *cd.Result) {
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
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported type:%v", val.String()))
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

	val = reflect.Indirect(val)
	if IsNil(val) {
		ret = true
		return
	}

	if val.Kind() == reflect.Slice {
		ret = val.Len() == 0
		return
	}

	ret = val.IsZero()
	return
}

// isSameStruct check if same
func isSameStruct(firstVal, secondVal reflect.Value) (ret bool, err *cd.Result) {
	firstNum := firstVal.NumField()
	secondNum := secondVal.NumField()
	if firstNum != secondNum {
		ret = false
		return
	}

	for idx := 0; idx < firstNum; idx++ {
		firstField := firstVal.Field(idx)
		secondField := secondVal.Field(idx)
		ret, err = isSameVal(firstField, secondField)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	ret = true
	return
}

// isSameVal is same value
func isSameVal(firstVal, secondVal reflect.Value) (ret bool, err *cd.Result) {
	if firstVal.Kind() == reflect.Interface {
		firstVal = firstVal.Elem()
	}
	if secondVal.Kind() == reflect.Interface {
		secondVal = secondVal.Elem()
	}
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
		log.Errorf("GetTypeEnum failed, error:%s", typeErr.Error())
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
			ret = firstVal.Float() == secondVal.Float()
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
		ret, err = isSameVal(firstItem, secondItem)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	return
}

func IsSameValue(firstVal, secondVal any) (ret bool) {
	rFirstVal := reflect.ValueOf(firstVal)
	rSecondVal := reflect.ValueOf(secondVal)
	sameOK, sameErr := isSameVal(rFirstVal, rSecondVal)
	ret = sameOK && sameErr == nil

	return
}

func GetBool(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawBool(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetInt(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawInt(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetInt8(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawInt8(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetInt16(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawInt16(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetInt32(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawInt32(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetInt64(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawInt64(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetUint(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawUint(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetUint8(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawUint8(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetUint16(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawUint16(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetUint32(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawUint32(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetUint64(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawUint64(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetFloat32(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawFloat32(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetFloat64(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawFloat64(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetString(val any) (ret model.RawVal, err *cd.Result) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawString(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetDateTime(val any) (ret model.RawVal, err *cd.Result) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal dateTime value, val:%v", val))
		}
	}()

	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := GetRawDateTime(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = model.NewRawVal(rawVal)
	return
}

func GetRawBool(rVal reflect.Value) (ret bool, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		ret = rVal.Bool()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = rVal.Int() > 0
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = rVal.Uint() > 0
	case reflect.Float32, reflect.Float64:
		ret = rVal.Float() > 0
	case reflect.String:
		ret = len(rVal.String()) > 0 && rVal.String() == "1"
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal bool value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawInt(rVal reflect.Value) (ret int, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = int(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = int(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = int(rVal.Float())
	case reflect.String:
		i64, iErr := strconv.ParseInt(rVal.String(), 0, 64)
		if iErr != nil {
			err = cd.NewResult(cd.UnExpected, iErr.Error())
			return
		}

		ret = int(i64)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal int value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawInt8(rVal reflect.Value) (ret int8, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = int8(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = int8(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = int8(rVal.Float())
	case reflect.String:
		i64, iErr := strconv.ParseInt(rVal.String(), 0, 64)
		if iErr != nil {
			err = cd.NewResult(cd.UnExpected, iErr.Error())
			return
		}

		ret = int8(i64)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal int8 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawInt16(rVal reflect.Value) (ret int16, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = int16(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = int16(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = int16(rVal.Float())
	case reflect.String:
		i64, iErr := strconv.ParseInt(rVal.String(), 0, 64)
		if iErr != nil {
			err = cd.NewResult(cd.UnExpected, iErr.Error())
			return
		}

		ret = int16(i64)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal int16 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawInt32(rVal reflect.Value) (ret int32, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = int32(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = int32(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = int32(rVal.Float())
	case reflect.String:
		i64, iErr := strconv.ParseInt(rVal.String(), 0, 64)
		if iErr != nil {
			err = cd.NewResult(cd.UnExpected, iErr.Error())
			return
		}

		ret = int32(i64)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal int32 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawInt64(rVal reflect.Value) (ret int64, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = int64(1)
		} else {
			ret = int64(0)
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = rVal.Int()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = int64(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = int64(rVal.Float())
	case reflect.String:
		i64, iErr := strconv.ParseInt(rVal.String(), 0, 64)
		if iErr != nil {
			err = cd.NewResult(cd.UnExpected, iErr.Error())
			return
		}

		ret = i64
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal int64 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawUint(rVal reflect.Value) (ret uint, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = uint(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = uint(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = uint(rVal.Float())
	case reflect.String:
		ui64, uiErr := strconv.ParseUint(rVal.String(), 0, 64)
		if uiErr != nil {
			err = cd.NewResult(cd.UnExpected, uiErr.Error())
			return
		}
		ret = uint(ui64)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal uint value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawUint8(rVal reflect.Value) (ret uint8, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = uint8(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = uint8(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = uint8(rVal.Float())
	case reflect.String:
		uiVal, uiErr := strconv.ParseUint(rVal.String(), 0, 64)
		if uiErr != nil {
			err = cd.NewResult(cd.UnExpected, uiErr.Error())
			return
		}
		ret = uint8(uiVal)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal uint8 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawUint16(rVal reflect.Value) (ret uint16, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = uint16(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = uint16(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = uint16(rVal.Float())
	case reflect.String:
		uiVal, uiErr := strconv.ParseUint(rVal.String(), 0, 64)
		if uiErr != nil {
			err = cd.NewResult(cd.UnExpected, uiErr.Error())
			return
		}
		ret = uint16(uiVal)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal uint16 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawUint32(rVal reflect.Value) (ret uint32, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = uint32(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = uint32(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = uint32(rVal.Float())
	case reflect.String:
		uiVal, uiErr := strconv.ParseUint(rVal.String(), 0, 64)
		if uiErr != nil {
			err = cd.NewResult(cd.UnExpected, uiErr.Error())
			return
		}
		ret = uint32(uiVal)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal uint32 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawUint64(rVal reflect.Value) (ret uint64, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = uint64(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = rVal.Uint()
	case reflect.Float32, reflect.Float64:
		ret = uint64(rVal.Float())
	case reflect.String:
		uiVal, uiErr := strconv.ParseUint(rVal.String(), 0, 64)
		if uiErr != nil {
			err = cd.NewResult(cd.UnExpected, uiErr.Error())
			return
		}
		ret = uiVal
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal uint64 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawFloat32(rVal reflect.Value) (ret float32, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = float32(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = float32(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = float32(rVal.Float())
	case reflect.String:
		fVal, fErr := strconv.ParseFloat(rVal.String(), 32)
		if fErr != nil {
			err = cd.NewResult(cd.UnExpected, fErr.Error())
			return
		}

		ret = float32(fVal)
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal float32 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawFloat64(rVal reflect.Value) (ret float64, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = 1
		} else {
			ret = 0
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = float64(rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = float64(rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = rVal.Float()
	case reflect.String:
		fVal, fErr := strconv.ParseFloat(rVal.String(), 64)
		if fErr != nil {
			err = cd.NewResult(cd.UnExpected, fErr.Error())
			return
		}

		ret = fVal
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal float64 value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawString(rVal reflect.Value) (ret string, err *cd.Result) {
	switch rVal.Kind() {
	case reflect.Bool:
		if rVal.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = fmt.Sprintf("%d", rVal.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = fmt.Sprintf("%d", rVal.Uint())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", rVal.Float())
	case reflect.String:
		ret = rVal.String()
	case reflect.Struct:
		switch rVal.Type().String() {
		case "time.Time":
			ret = rVal.Interface().(time.Time).Format(fu.CSTLayout)
		default:
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal string value, val type:%v", rVal.Type().String()))
		}
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal string value, val type:%v", rVal.Type().String()))
	}

	return
}

func GetRawDateTime(rVal reflect.Value) (ret time.Time, err *cd.Result) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal dateTime value, val:%v", rVal.Interface()))
		}
	}()

	switch rVal.Kind() {
	case reflect.String:
		if rVal.String() == "" {
			ret = time.Time{}
			return
		}
		
		tVal, tErr := time.Parse(fu.CSTLayout, rVal.String())
		if tErr != nil {
			err = cd.NewResult(cd.UnExpected, tErr.Error())
			return
		}
		ret = tVal
	case reflect.Struct:
		switch rVal.Type().String() {
		case "time.Time":
			ret = rVal.Interface().(time.Time)
		default:
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal dateTime value, val type:%v", rVal.Type().String()))
		}
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal dateTime value, val type:%v", rVal.Type().String()))
	}

	return
}
