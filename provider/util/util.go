package util

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
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
	DateTime  = "datetime"
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

	return int64(s.PageSize)
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

var (
	// 类型映射表
	integerKindMap = map[reflect.Kind]bool{
		reflect.Int8: true, reflect.Int16: true, reflect.Int32: true, reflect.Int: true, reflect.Int64: true,
	}

	uintegerKindMap = map[reflect.Kind]bool{
		reflect.Uint8: true, reflect.Uint16: true, reflect.Uint32: true, reflect.Uint: true, reflect.Uint64: true,
	}

	floatKindMap = map[reflect.Kind]bool{
		reflect.Float32: true, reflect.Float64: true,
	}

	// numberKindMap包含所有数字类型
	numberKindMap = map[reflect.Kind]bool{
		reflect.Int8: true, reflect.Int16: true, reflect.Int32: true, reflect.Int: true, reflect.Int64: true,
		reflect.Uint8: true, reflect.Uint16: true, reflect.Uint32: true, reflect.Uint: true, reflect.Uint64: true,
		reflect.Float32: true, reflect.Float64: true,
	}
)

func IsInteger(tType reflect.Type) bool {
	return integerKindMap[tType.Kind()]
}

func IsUInteger(tType reflect.Type) bool {
	return uintegerKindMap[tType.Kind()]
}

func IsFloat(tType reflect.Type) bool {
	return floatKindMap[tType.Kind()]
}

func IsNumber(tType reflect.Type) bool {
	return numberKindMap[tType.Kind()]
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

// typeEnumMap 用于快速查找基础类型的枚举值
var typeEnumMap = map[reflect.Kind]model.TypeDeclare{
	reflect.Int8:    model.TypeBitValue,
	reflect.Uint8:   model.TypePositiveBitValue,
	reflect.Int16:   model.TypeSmallIntegerValue,
	reflect.Uint16:  model.TypePositiveSmallIntegerValue,
	reflect.Int32:   model.TypeInteger32Value,
	reflect.Uint32:  model.TypePositiveInteger32Value,
	reflect.Int64:   model.TypeBigIntegerValue,
	reflect.Uint64:  model.TypePositiveBigIntegerValue,
	reflect.Int:     model.TypeIntegerValue,
	reflect.Uint:    model.TypePositiveIntegerValue,
	reflect.Float32: model.TypeFloatValue,
	reflect.Float64: model.TypeDoubleValue,
	reflect.Bool:    model.TypeBooleanValue,
	reflect.String:  model.TypeStringValue,
	reflect.Map:     model.TypeMapValue,
}

func GetTypeEnum(val reflect.Type) (ret model.TypeDeclare, err *cd.Result) {
	log.Infof("GetTypeEnum, val:%v", val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		return GetTypeEnum(val.Elem())
	}

	// 从映射表中查找基础类型
	if enumVal, exists := typeEnumMap[val.Kind()]; exists {
		ret = enumVal
		return
	}

	// 处理特殊类型
	switch val.Kind() {
	case reflect.Struct:
		if val.String() == "time.Time" {
			ret = model.TypeDateTimeValue
		} else {
			ret = model.TypeStructValue
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				log.Infof("3....isSameVal, type name:%v, field name:%v, type:%v", val, field.Name, field.Type)
				if !isValidFieldType(field.Type) {
					err = cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported field type in struct: %v", field.Type))
					return
				}
			}
		}
	case reflect.Slice:
		eType := val.Elem()
		if eType.Kind() == reflect.Ptr {
			eType = eType.Elem()
		}
		if _, err = GetTypeEnum(eType); err != nil {
			return
		}
		ret = model.TypeSliceValue
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported type: %v", val.String()))
	}

	return
}

func isValidFieldType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr:
		return isValidFieldType(t.Elem())
	case reflect.Struct, reflect.Slice:
		return true
	default:
		_, err := GetTypeEnum(t)
		return err == nil
	}
}

// recoverToTrue 是一个通用的 recover 辅助函数
// 用于在恢复 panic 后返回 true
func recoverToTrue(ret *bool) {
	if err := recover(); err != nil {
		log.Errorf("Check failed: %v", err)
		*ret = true
	}
}

// IsNil checks if a value is nil.
// It returns true if the value is nil, false otherwise.
// For pointer types, it returns true if the pointer is uninitialized.
// If val is not a valid Value (i.e., val.IsValid() returns false), it returns true.
func IsNil(val reflect.Value) (ret bool) {
	defer recoverToTrue(&ret)

	if !val.IsValid() {
		return true
	}

	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}

// IsZero checks if a value is zero (the initial value for its type).
// It returns true if the value is zero, false otherwise.
func IsZero(val reflect.Value) (ret bool) {
	defer recoverToTrue(&ret)

	val = reflect.Indirect(val)
	if !val.IsValid() {
		return true
	}

	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	case reflect.Array:
		return val.Len() == 0 || IsZero(val.Index(0))
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			if !IsZero(val.Field(i)) {
				return false
			}
		}
		return true
	default:
		return val.IsZero()
	}
}

// IsSameValue 判断两个值是否相同
func IsSameValue(firstVal, secondVal any) (ret bool) {
	// 如果两个值都是 nil，则认为它们相同
	if firstVal == nil && secondVal == nil {
		return true
	}
	
	// 如果只有一个值是 nil，则认为它们不同
	if firstVal == nil || secondVal == nil {
		return false
	}
	
	rFirstVal := reflect.ValueOf(firstVal)
	rSecondVal := reflect.ValueOf(secondVal)
	
	// 如果类型不同，则认为它们不同
	if rFirstVal.Type() != rSecondVal.Type() {
		return false
	}
	
	// 根据类型进行比较
	switch rFirstVal.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		// 基本类型可以直接比较
		return rFirstVal.Interface() == rSecondVal.Interface()
		
	case reflect.Slice, reflect.Array:
		// 比较切片或数组的每个元素
		length := rFirstVal.Len()
		if length != rSecondVal.Len() {
			return false
		}
		
		for i := 0; i < length; i++ {
			if !IsSameValue(rFirstVal.Index(i).Interface(), rSecondVal.Index(i).Interface()) {
				return false
			}
		}
		return true
		
	case reflect.Map:
		// 比较映射的键值对
		keys := rFirstVal.MapKeys()
		if len(keys) != rSecondVal.Len() {
			return false
		}
		
		for _, key := range keys {
			val1 := rFirstVal.MapIndex(key)
			val2 := rSecondVal.MapIndex(key)
			if !val2.IsValid() || !IsSameValue(val1.Interface(), val2.Interface()) {
				return false
			}
		}
		return true
		
	case reflect.Struct:
		// 比较结构体的每个字段
		numField := rFirstVal.NumField()
		for i := 0; i < numField; i++ {
			field1 := rFirstVal.Field(i)
			field2 := rSecondVal.Field(i)
			
			// 跳过不可比较的字段
			if !field1.CanInterface() || !field2.CanInterface() {
				continue
			}
			
			if !IsSameValue(field1.Interface(), field2.Interface()) {
				return false
			}
		}
		return true
		
	case reflect.Ptr, reflect.Interface:
		// 如果是指针或接口，则比较它们指向的值
		if rFirstVal.IsNil() && rSecondVal.IsNil() {
			return true
		}
		if rFirstVal.IsNil() || rSecondVal.IsNil() {
			return false
		}
		return IsSameValue(rFirstVal.Elem().Interface(), rSecondVal.Elem().Interface())
		
	default:
		// 对于其他类型，尝试直接比较
		// 注意：这可能不适用于所有类型
		return rFirstVal.Interface() == rSecondVal.Interface()
	}
}

// GetBool get bool
// 如果val为指针值，尝试将其转换成*bool，否则转换成bool
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetInt get int
// 如果val为指针值，尝试将其转换成*int，否则转换成int
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetInt8 get int8
// 如果val为指针值，尝试将其转换成*int8，否则转换成int8
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetInt16 get int16
// 如果val为指针值，尝试将其转换成*int16，否则转换成int16
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetInt32 get int32
// 如果val为指针值，尝试将其转换成*int32，否则转换成int32
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetInt64 get int64
// 如果val为指针值，尝试将其转换成*int64，否则转换成int64
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetUint get uint
// 如果val为指针值，尝试将其转换成*uint，否则转换成uint
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetUint8 get uint8
// 如果val为指针值，尝试将其转换成*uint8，否则转换成uint8
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetUint16 get uint16
// 如果val为指针值，尝试将其转换成*uint16，否则转换成uint16
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetUint32 get uint32
// 如果val为指针值，尝试将其转换成*uint32，否则转换成uint32
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetUint64 get uint64
// 如果val为指针值，尝试将其转换成*uint64，否则转换成uint64
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetFloat32 get float32
// 如果val为指针值，尝试将其转换成*float32，否则转换成float32
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetFloat64 get float64
// 如果val为指针值，尝试将其转换成*float64，否则转换成float64
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetString get string
// 如果val为指针值，尝试将其转换成*string，否则转换成string
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetDateTime get dateTime
// 如果val为指针值，尝试将其转换成*time.Time，否则转换成time.Time
// 将转换后的结果以model.RawVal形式返回
// 转换出错返回*cd.Result
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

// GetRawBool get bool
// 将各基础数据类型的值转换为布尔值
// rVal如果是Bool类型，则返回其值
// rVal如果是数值类型，则大于0为true,否则为false
// rVal如果是字符串类型，则尝试将其解析成bool，接受 "true"、"yes"、"1" 等常见 true 值（不区分大小写）
// rVal其他类型返回错误
func GetRawBool(rVal reflect.Value) (ret bool, err *cd.Result) {
	trueSynonyms := map[string]bool{
		"true":  true,
		"yes":   true,
		"1":     true,
		"t":     true,
		"y":     true,
		"on":    true,
		"ok":    true,
		"true;": true,
	}

	switch rVal.Kind() {
	case reflect.Bool:
		ret = rVal.Bool()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = rVal.Int() != 0
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = rVal.Uint() != 0
	case reflect.Float32, reflect.Float64:
		ret = rVal.Float() != 0
	case reflect.String:
		strVal := strings.ToLower(strings.TrimSpace(rVal.String()))
		ret = trueSynonyms[strVal]
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal bool value, val type:%v", rVal.Type().String()))
	}

	return
}

// GetRawInt get int
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int
// rVal如果是字符串类型，则尝试将其解析成Int
// rVal其他类型返回错误
func GetRawInt(rVal reflect.Value) (ret int, err *cd.Result) {
	result, err := convertNumberVal(reflect.Int, rVal)
	if err != nil {
		return
	}
	ret = result.(int)
	return
}

// GetRawInt8 get int8
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int8
// rVal如果是字符串类型，则尝试将其解析成Int8
// rVal其他类型返回错误
func GetRawInt8(rVal reflect.Value) (ret int8, err *cd.Result) {
	result, err := convertNumberVal(reflect.Int8, rVal)
	if err != nil {
		return
	}
	ret = result.(int8)
	return
}

// GetRawInt16 get int16
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int16
// rVal如果是字符串类型，则尝试将其解析成Int16
// rVal其他类型返回错误
func GetRawInt16(rVal reflect.Value) (ret int16, err *cd.Result) {
	result, err := convertNumberVal(reflect.Int16, rVal)
	if err != nil {
		return
	}
	ret = result.(int16)
	return
}

// GetRawInt32 get int32
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int32
// rVal如果是字符串类型，则尝试将其解析成Int32
// rVal其他类型返回错误
func GetRawInt32(rVal reflect.Value) (ret int32, err *cd.Result) {
	result, err := convertNumberVal(reflect.Int32, rVal)
	if err != nil {
		return
	}
	ret = result.(int32)
	return
}

// GetRawInt64 get int64
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int64
// rVal如果是字符串类型，则尝试将其解析成Int64
// rVal其他类型返回错误
func GetRawInt64(rVal reflect.Value) (ret int64, err *cd.Result) {
	result, err := convertNumberVal(reflect.Int64, rVal)
	if err != nil {
		return
	}
	ret = result.(int64)
	return
}

// GetRawUint get uint
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint
// rVal如果是字符串类型，则尝试将其解析成Uint
// rVal其他类型返回错误
func GetRawUint(rVal reflect.Value) (ret uint, err *cd.Result) {
	result, err := convertNumberVal(reflect.Uint, rVal)
	if err != nil {
		return
	}
	ret = result.(uint)
	return
}

// GetRawUint8 get uint8
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint8
// rVal如果是字符串类型，则尝试将其解析成Uint8
// rVal其他类型返回错误
func GetRawUint8(rVal reflect.Value) (ret uint8, err *cd.Result) {
	result, err := convertNumberVal(reflect.Uint8, rVal)
	if err != nil {
		return
	}
	ret = result.(uint8)
	return
}

// GetRawUint16 get uint16
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint16
// rVal如果是字符串类型，则尝试将其解析成Uint16
// rVal其他类型返回错误
func GetRawUint16(rVal reflect.Value) (ret uint16, err *cd.Result) {
	result, err := convertNumberVal(reflect.Uint16, rVal)
	if err != nil {
		return
	}
	ret = result.(uint16)
	return
}

// GetRawUint32 get uint32
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint32
// rVal如果是字符串类型，则尝试将其解析成Uint32
// rVal其他类型返回错误
func GetRawUint32(rVal reflect.Value) (ret uint32, err *cd.Result) {
	result, err := convertNumberVal(reflect.Uint32, rVal)
	if err != nil {
		return
	}
	ret = result.(uint32)
	return
}

// GetRawUint64 get uint64
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint64
// rVal如果是字符串类型，则尝试将其解析成Uint64
// rVal其他类型返回错误
func GetRawUint64(rVal reflect.Value) (ret uint64, err *cd.Result) {
	result, err := convertNumberVal(reflect.Uint64, rVal)
	if err != nil {
		return
	}
	ret = result.(uint64)
	return
}

// GetRawFloat32 get float32
// 将各基础数据类型的值转换为浮点数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Float32
// rVal如果是字符串类型，则尝试将其解析成Float32
// rVal其他类型返回错误
func GetRawFloat32(rVal reflect.Value) (ret float32, err *cd.Result) {
	result, err := convertNumberVal(reflect.Float32, rVal)
	if err != nil {
		return
	}
	ret = result.(float32)
	return
}

// GetRawFloat64 get float64
// 将各基础数据类型的值转换为浮点数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Float64
// rVal如果是字符串类型，则尝试将其解析成Float64
// rVal其他类型返回错误
func GetRawFloat64(rVal reflect.Value) (ret float64, err *cd.Result) {
	result, err := convertNumberVal(reflect.Float64, rVal)
	if err != nil {
		return
	}
	ret = result.(float64)
	return
}

// GetRawString get string
// 将各基础数据类型的值转换为字符串
// rVal的类型如果是基础数据类型，则将其格式化成对应的字符串
// rVal的类型如果是Bool,则将其格式化成"0"或"1"
// rVal的类型如果是Struct,则要求值的类型是time.Time,将其以CSTLayout格式化("2006-01-02 15:04:05"),其他类型的Struct不支持
// rVal如果是其他类型，则返回nil,并设置错误
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

// GetRawDateTime get dateTime
// rVal 对应的类型如果是String，则要求值的格式必须是符合CSTLayout的时间格式("2006-01-02 15:04:05")
// rVal 对应的类型如果是Struct，则要求值是time.Time类型
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

// convertNumberVal 是一个通用的数值转换函数，用于各种整数和浮点数转换
// kind 指定要转换的类型，例如 reflect.Int64
// rVal 是要转换的 reflect.Value
// 返回一个 interface{} 和一个错误
// 要求返回值严格符合 kind 的类型
func convertNumberVal(kind reflect.Kind, rVal reflect.Value) (result interface{}, err *cd.Result) {
	if !numberKindMap[kind] {
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported target kind: %v", kind))
	}

	switch rVal.Kind() {
	case reflect.Bool:
		return convertBoolToNumber(kind, rVal.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertIntToNumber(kind, rVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertUintToNumber(kind, rVal.Uint())
	case reflect.Float32, reflect.Float64:
		return convertFloatToNumber(kind, rVal.Float())
	case reflect.String:
		return convertStringToNumber(kind, rVal.String())
	default:
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal %v value, val type:%v", kind, rVal.Type().String()))
	}
}

func convertBoolToNumber(kind reflect.Kind, val bool) (interface{}, *cd.Result) {
	var result interface{}
	if val {
		result = 1
	} else {
		result = 0
	}

	switch kind {
	case reflect.Int:
		return int(result.(int)), nil
	case reflect.Int8:
		return int8(result.(int)), nil
	case reflect.Int16:
		return int16(result.(int)), nil
	case reflect.Int32:
		return int32(result.(int)), nil
	case reflect.Int64:
		return int64(result.(int)), nil
	case reflect.Uint:
		return uint(result.(int)), nil
	case reflect.Uint8:
		return uint8(result.(int)), nil
	case reflect.Uint16:
		return uint16(result.(int)), nil
	case reflect.Uint32:
		return uint32(result.(int)), nil
	case reflect.Uint64:
		return uint64(result.(int)), nil
	case reflect.Float32:
		return float32(result.(int)), nil
	case reflect.Float64:
		return float64(result.(int)), nil
	default:
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported conversion from bool to %v", kind))
	}
}

func convertIntToNumber(kind reflect.Kind, val int64) (interface{}, *cd.Result) {
	switch kind {
	case reflect.Int:
		return int(val), nil
	case reflect.Int8:
		return int8(val), nil
	case reflect.Int16:
		return int16(val), nil
	case reflect.Int32:
		return int32(val), nil
	case reflect.Int64:
		return val, nil
	case reflect.Uint:
		return uint(val), nil
	case reflect.Uint8:
		return uint8(val), nil
	case reflect.Uint16:
		return uint16(val), nil
	case reflect.Uint32:
		return uint32(val), nil
	case reflect.Uint64:
		return uint64(val), nil
	case reflect.Float32:
		return float32(val), nil
	case reflect.Float64:
		return float64(val), nil
	default:
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported conversion from int64 to %v", kind))
	}
}

func convertUintToNumber(kind reflect.Kind, val uint64) (interface{}, *cd.Result) {
	switch kind {
	case reflect.Int:
		return int(val), nil
	case reflect.Int8:
		return int8(val), nil
	case reflect.Int16:
		return int16(val), nil
	case reflect.Int32:
		return int32(val), nil
	case reflect.Int64:
		return int64(val), nil
	case reflect.Uint:
		return uint(val), nil
	case reflect.Uint8:
		return uint8(val), nil
	case reflect.Uint16:
		return uint16(val), nil
	case reflect.Uint32:
		return uint32(val), nil
	case reflect.Uint64:
		return val, nil
	case reflect.Float32:
		return float32(val), nil
	case reflect.Float64:
		return float64(val), nil
	default:
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported conversion from uint64 to %v", kind))
	}
}

func convertFloatToNumber(kind reflect.Kind, val float64) (interface{}, *cd.Result) {
	switch kind {
	case reflect.Int:
		return int(val), nil
	case reflect.Int8:
		return int8(val), nil
	case reflect.Int16:
		return int16(val), nil
	case reflect.Int32:
		return int32(val), nil
	case reflect.Int64:
		return int64(val), nil
	case reflect.Uint:
		return uint(val), nil
	case reflect.Uint8:
		return uint8(val), nil
	case reflect.Uint16:
		return uint16(val), nil
	case reflect.Uint32:
		return uint32(val), nil
	case reflect.Uint64:
		return uint64(val), nil
	case reflect.Float32:
		return float32(val), nil
	case reflect.Float64:
		return val, nil
	default:
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported conversion from float64 to %v", kind))
	}
}

func convertStringToNumber(kind reflect.Kind, val string) (interface{}, *cd.Result) {
	switch {
	case integerKindMap[kind]:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("parse int value failed, error:%s", err.Error()))
		}
		return convertIntToNumber(kind, i)
	case uintegerKindMap[kind]:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("parse uint value failed, error:%s", err.Error()))
		}
		return convertUintToNumber(kind, u)
	case floatKindMap[kind]:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("parse float value failed, error:%s", err.Error()))
		}
		return convertFloatToNumber(kind, f)
	default:
		return nil, cd.NewResult(cd.UnExpected, fmt.Sprintf("unsupported conversion from string to %v", kind))
	}
}
