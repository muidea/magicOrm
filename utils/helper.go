package utils

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	fu "github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/models"
)

const ()

var snowFlakeNodePtr *fu.SnowflakeNode
var snowFlakeOnce sync.Once

func init() {
	snowFlakeOnce.Do(func() {
		strNodeID := os.Getenv("NODE_ID")
		if strNodeID == "" {
			strNodeID = "1"
		}
		nodeID, nodeErr := strconv.ParseInt(strNodeID, 10, 64)
		if nodeErr != nil {
			nodeID = 1
		}

		snowFlakeNodePtr, _ = fu.NewSnowflakeNode(nodeID)
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
	ret = time.Now().UTC().Format(fu.CSTLayoutWithMillisecond)
	return
}

func GetNewUUID() (ret string) {
	ret = fu.NewUUID()
	return
}

func GetNewSnowflakeID() (ret int64) {
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
	switch tType.String() {
	case models.TypeStructTimeName:
		return true
	default:
		return false
	}
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
var typeEnumMap = map[reflect.Kind]models.TypeDeclare{
	reflect.Int8:    models.TypeByteValue,
	reflect.Uint8:   models.TypePositiveByteValue,
	reflect.Int16:   models.TypeSmallIntegerValue,
	reflect.Uint16:  models.TypePositiveSmallIntegerValue,
	reflect.Int32:   models.TypeInteger32Value,
	reflect.Uint32:  models.TypePositiveInteger32Value,
	reflect.Int64:   models.TypeBigIntegerValue,
	reflect.Uint64:  models.TypePositiveBigIntegerValue,
	reflect.Int:     models.TypeIntegerValue,
	reflect.Uint:    models.TypePositiveIntegerValue,
	reflect.Float32: models.TypeFloatValue,
	reflect.Float64: models.TypeDoubleValue,
	reflect.Bool:    models.TypeBooleanValue,
	reflect.String:  models.TypeStringValue,
}

func GetTypeEnum(val reflect.Type) (ret models.TypeDeclare, err *cd.Error) {
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
		if val.String() == models.TypeStructTimeName {
			ret = models.TypeDateTimeValue
		} else {
			ret = models.TypeStructValue
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				if !isValidFieldType(field.Type) {
					err = cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported field type in struct: %v", field.Type))
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
		ret = models.TypeSliceValue
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported type: %v", val.String()))
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

// ConvertRawToBool convert raw bool
// 如果val为指针值，尝试将其转换成*bool，否则转换成bool
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToBool(val any) (ret bool, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToBool(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToInt convert raw int
// 如果val为指针值，尝试将其转换成*int，否则转换成int
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt(val any) (ret int, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToInt(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToInt8 convert raw int8
// 如果val为指针值，尝试将其转换成*int8，否则转换成int8
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt8(val any) (ret int8, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToInt8(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToInt16 convert raw int16
// 如果val为指针值，尝试将其转换成*int16，否则转换成int16
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt16(val any) (ret int16, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToInt16(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToInt32 convert raw int32
// 如果val为指针值，尝试将其转换成*int32，否则转换成int32
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt32(val any) (ret int32, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToInt32(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToInt64 convert raw int64
// 如果val为指针值，尝试将其转换成*int64，否则转换成int64
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt64(val any) (ret int64, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToInt64(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToUint convert raw uint
// 如果val为指针值，尝试将其转换成*uint，否则转换成uint
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint(val any) (ret uint, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToUint(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToUint8 convert raw uint8
// 如果val为指针值，尝试将其转换成*uint8，否则转换成uint8
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint8(val any) (ret uint8, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToUint8(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToUint16 convert raw uint16
// 如果val为指针值，尝试将其转换成*uint16，否则转换成uint16
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint16(val any) (ret uint16, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToUint16(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToUint32 convert raw uint32
// 如果val为指针值，尝试将其转换成*uint32，否则转换成uint32
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint32(val any) (ret uint32, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToUint32(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToUint64 convert raw uint64
// 如果val为指针值，尝试将其转换成*uint64，否则转换成uint64
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint64(val any) (ret uint64, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToUint64(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToFloat32 convert raw float32
// 如果val为指针值，尝试将其转换成*float32，否则转换成float32
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToFloat32(val any) (ret float32, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToFloat32(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToFloat64 convert raw float64
// 如果val为指针值，尝试将其转换成*float64，否则转换成float64
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToFloat64(val any) (ret float64, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToFloat64(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToString convert raw string
// 如果val为指针值，尝试将其转换成*string，否则转换成string
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToString(val any) (ret string, err *cd.Error) {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToString(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertRawToDateTime convert raw datetime
// 如果val为指针值，尝试将其转换成*time.Time，否则转换成time.Time
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToDateTime(val any) (ret time.Time, err *cd.Error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal datetime value, val:%v", val))
		}
	}()

	rVal := reflect.Indirect(reflect.ValueOf(val))
	rawVal, rawErr := ConvertToDateTime(rVal)
	if rawErr != nil {
		err = rawErr
		return
	}
	ret = rawVal
	return
}

// ConvertToBool convert bool
// 将各基础数据类型的值转换为布尔值
// rVal如果是Bool类型，则返回其值
// rVal如果是数值类型，则大于0为true,否则为false
// rVal如果是字符串类型，则尝试将其解析成bool，接受 "true"、"yes"、"1" 等常见 true 值（不区分大小写）
// rVal其他类型返回错误
func ConvertToBool(rVal reflect.Value) (ret bool, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
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
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal bool value, val type:%v", rVal.Type().String()))
	}

	return
}

// ConvertToInt convert int
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int
// rVal如果是字符串类型，则尝试将其解析成Int
// rVal其他类型返回错误
func ConvertToInt(rVal reflect.Value) (ret int, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Int, rVal)
	if err != nil {
		return
	}
	ret = result.(int)
	return
}

// ConvertToInt8 convert int8
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int8
// rVal如果是字符串类型，则尝试将其解析成Int8
// rVal其他类型返回错误
func ConvertToInt8(rVal reflect.Value) (ret int8, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Int8, rVal)
	if err != nil {
		return
	}
	ret = result.(int8)
	return
}

// ConvertToInt16 convert int16
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int16
// rVal如果是字符串类型，则尝试将其解析成Int16
// rVal其他类型返回错误
func ConvertToInt16(rVal reflect.Value) (ret int16, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Int16, rVal)
	if err != nil {
		return
	}
	ret = result.(int16)
	return
}

// ConvertToInt32 convert int32
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int32
// rVal如果是字符串类型，则尝试将其解析成Int32
// rVal其他类型返回错误
func ConvertToInt32(rVal reflect.Value) (ret int32, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Int32, rVal)
	if err != nil {
		return
	}
	ret = result.(int32)
	return
}

// ConvertToInt64 convert int64
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Int64
// rVal如果是字符串类型，则尝试将其解析成Int64
// rVal其他类型返回错误
func ConvertToInt64(rVal reflect.Value) (ret int64, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Int64, rVal)
	if err != nil {
		return
	}
	ret = result.(int64)
	return
}

// ConvertToUint convert uint
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint
// rVal如果是字符串类型，则尝试将其解析成Uint
// rVal其他类型返回错误
func ConvertToUint(rVal reflect.Value) (ret uint, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Uint, rVal)
	if err != nil {
		return
	}
	ret = result.(uint)
	return
}

// ConvertToUint8 convert uint8
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint8
// rVal如果是字符串类型，则尝试将其解析成Uint8
// rVal其他类型返回错误
func ConvertToUint8(rVal reflect.Value) (ret uint8, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Uint8, rVal)
	if err != nil {
		return
	}
	ret = result.(uint8)
	return
}

// ConvertToUint16 convert uint16
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint16
// rVal如果是字符串类型，则尝试将其解析成Uint16
// rVal其他类型返回错误
func ConvertToUint16(rVal reflect.Value) (ret uint16, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Uint16, rVal)
	if err != nil {
		return
	}
	ret = result.(uint16)
	return
}

// ConvertToUint32 convert uint32
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint32
// rVal如果是字符串类型，则尝试将其解析成Uint32
// rVal其他类型返回错误
func ConvertToUint32(rVal reflect.Value) (ret uint32, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Uint32, rVal)
	if err != nil {
		return
	}
	ret = result.(uint32)
	return
}

// ConvertToUint64 convert uint64
// 将各基础数据类型的值转换为整数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Uint64
// rVal如果是字符串类型，则尝试将其解析成Uint64
// rVal其他类型返回错误
func ConvertToUint64(rVal reflect.Value) (ret uint64, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Uint64, rVal)
	if err != nil {
		return
	}
	ret = result.(uint64)
	return
}

// ConvertToFloat32 convert float32
// 将各基础数据类型的值转换为浮点数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Float32
// rVal如果是字符串类型，则尝试将其解析成Float32
// rVal其他类型返回错误
func ConvertToFloat32(rVal reflect.Value) (ret float32, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Float32, rVal)
	if err != nil {
		return
	}
	ret = result.(float32)
	return
}

// ConvertToFloat64 convert float64
// 将各基础数据类型的值转换为浮点数
// rVal如果是Bool类型，则将其格式化成"0"或"1"
// rVal如果是数值类型，则转换成对应的Float64
// rVal如果是字符串类型，则尝试将其解析成Float64
// rVal其他类型返回错误
func ConvertToFloat64(rVal reflect.Value) (ret float64, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
	result, err := convertNumberVal(reflect.Float64, rVal)
	if err != nil {
		return
	}
	ret = result.(float64)
	return
}

// ConvertToString convert string
// 将各基础数据类型的值转换为字符串
// rVal的类型如果是基础数据类型，则将其格式化成对应的字符串
// rVal的类型如果是Bool,则将其格式化成"0"或"1"
// rVal的类型如果是Struct,则要求值的类型是time.Time,将其以CSTLayout格式化("2006-01-02 15:04:05"),其他类型的Struct不支持
// rVal如果是其他类型，则返回nil,并设置错误
func ConvertToString(rVal reflect.Value) (ret string, err *cd.Error) {
	rVal = reflect.Indirect(rVal)
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
		case models.TypeStructTimeName:
			ret = rVal.Interface().(time.Time).Format(fu.CSTLayoutWithMillisecond)
		default:
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal string value, val type:%v", rVal.Type().String()))
		}
	case reflect.Array, reflect.Slice:
		ret = fmt.Sprintf("%s", rVal.Interface())
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal string value, val type:%v", rVal.Type().String()))
	}

	return
}

// ConvertToDateTime convert datetime
// rVal 对应的类型如果是String，则要求值的格式必须是符合CSTLayout的时间格式("2006-01-02 15:04:05")
// rVal 对应的类型如果是Struct，则要求值是time.Time类型
func ConvertToDateTime(rVal reflect.Value) (ret time.Time, err *cd.Error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal datetime value, val:%v", rVal.Interface()))
		}
	}()

	rVal = reflect.Indirect(rVal)
	switch rVal.Kind() {
	case reflect.String:
		if rVal.String() == "" {
			ret = time.Time{}
			return
		}

		tVal, tErr := time.Parse(fu.CSTLayoutWithMillisecond, rVal.String())
		if tErr != nil {
			err = cd.NewError(cd.Unexpected, tErr.Error())
			return
		}
		ret = tVal
	case reflect.Struct:
		switch rVal.Type().String() {
		case models.TypeStructTimeName:
			ret = rVal.Interface().(time.Time)
		default:
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal datetime value, val type:%v", rVal.Type().String()))
		}
	case reflect.Array, reflect.Slice:
		tVal, tErr := time.Parse(fu.CSTLayoutWithMillisecond, fmt.Sprintf("%s", rVal.Interface()))
		if tErr != nil {
			err = cd.NewError(cd.Unexpected, tErr.Error())
			return
		}
		ret = tVal
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal datetime value, val type:%v", rVal.Type().String()))
	}

	return
}

// convertNumberVal 是一个通用的数值转换函数，用于各种整数和浮点数转换
// kind 指定要转换的类型，例如 reflect.Int64
// rVal 是要转换的 reflect.Value
// 返回一个 interface{} 和一个错误
// 要求返回值严格符合 kind 的类型
func convertNumberVal(kind reflect.Kind, rVal reflect.Value) (result interface{}, err *cd.Error) {
	if !numberKindMap[kind] {
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported target kind: %v", kind))
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
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("illegal %v value, val type:%v", kind, rVal.Type().String()))
	}
}

func convertBoolToNumber(kind reflect.Kind, val bool) (interface{}, *cd.Error) {
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
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported conversion from bool to %v", kind))
	}
}

func convertIntToNumber(kind reflect.Kind, val int64) (interface{}, *cd.Error) {
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
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported conversion from int64 to %v", kind))
	}
}

func convertUintToNumber(kind reflect.Kind, val uint64) (interface{}, *cd.Error) {
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
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported conversion from uint64 to %v", kind))
	}
}

func convertFloatToNumber(kind reflect.Kind, val float64) (interface{}, *cd.Error) {
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
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported conversion from float64 to %v", kind))
	}
}

func convertStringToNumber(kind reflect.Kind, val string) (interface{}, *cd.Error) {
	switch {
	case integerKindMap[kind]:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("parse int value failed, error:%s", err.Error()))
		}
		return convertIntToNumber(kind, i)
	case uintegerKindMap[kind]:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("parse uint value failed, error:%s", err.Error()))
		}
		return convertUintToNumber(kind, u)
	case floatKindMap[kind]:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("parse float value failed, error:%s", err.Error()))
		}
		return convertFloatToNumber(kind, f)
	default:
		return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("unsupported conversion from string to %v", kind))
	}
}

func ElemDependValue(vVal reflect.Value) (ret []reflect.Value, err *cd.Error) {
	if vVal.Kind() == reflect.Interface {
		vVal = vVal.Elem()
	}
	rVal := reflect.Indirect(vVal)
	if rVal.Kind() == reflect.Struct {
		ret = append(ret, vVal)
		return
	}

	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "illegal slice value")
		return
	}

	for idx := 0; idx < rVal.Len(); idx++ {
		val := rVal.Index(idx)
		ret = append(ret, val)
	}
	return
}
