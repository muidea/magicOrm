// Package utils provides model value conversion and filter data contracts.
package utils

import (
	"reflect"
	"time"

	"github.com/google/go-cmp/cmp"
	cd "github.com/muidea/magicCommon/def"
	impl "github.com/muidea/magicOrm/internal/utils"
	"github.com/muidea/magicOrm/models"
)

// CompareWithNumericConversion 支持安全类型转换的比较函数
// 返回值：是否相等，差异描述
func CompareWithNumericConversion(a, b any, opts ...cmp.Option) (bool, string) {
	return impl.CompareWithNumericConversion(a, b, opts...)
}

// parseConstraints 将 "req,min=1:100,ro" 解析为结构化 Map
func ParseConstraints(tagStr string) models.Constraints { return impl.ParseConstraints(tagStr) }

// ParseConstraintsChecked validates the built-in portable constraint language.
// ParseConstraints remains available to consumers registering custom directives.
func ParseConstraintsChecked(value string) (models.Constraints, error) {
	return impl.ParseConstraintsChecked(value)
}

// EncodeUTCDateTime produces the existing timezone-free database representation.
// Such text always denotes UTC, never the process or database session timezone.
func EncodeUTCDateTime(value reflect.Value) (string, *cd.Error) { return impl.EncodeUTCDateTime(value) }

func GetCurrentDateTime() (ret time.Time) { return impl.GetCurrentDateTime() }

func GetCurrentDateTimeStr() (ret string) { return impl.GetCurrentDateTimeStr() }

func GetNewUUID() (ret string) { return impl.GetNewUUID() }

func GetNewSnowflakeID() (ret int64) { return impl.GetNewSnowflakeID() }

func GetTypeEnum(val reflect.Type) (ret models.TypeDeclare, err *cd.Error) {
	return impl.GetTypeEnum(val)
}

// ConvertRawToBool convert raw bool
// 如果val为指针值，尝试将其转换成*boolean，否则转换成boolean
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToBool(val any) (ret bool, err *cd.Error) { return impl.ConvertRawToBool(val) }

// ConvertRawToInt convert raw int
// 如果val为指针值，尝试将其转换成*int，否则转换成int
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt(val any) (ret int, err *cd.Error) { return impl.ConvertRawToInt(val) }

// ConvertRawToInt8 convert raw int8
// 如果val为指针值，尝试将其转换成*int8，否则转换成int8
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt8(val any) (ret int8, err *cd.Error) { return impl.ConvertRawToInt8(val) }

// ConvertRawToInt16 convert raw int16
// 如果val为指针值，尝试将其转换成*int16，否则转换成int16
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt16(val any) (ret int16, err *cd.Error) { return impl.ConvertRawToInt16(val) }

// ConvertRawToInt32 convert raw int32
// 如果val为指针值，尝试将其转换成*int32，否则转换成int32
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt32(val any) (ret int32, err *cd.Error) { return impl.ConvertRawToInt32(val) }

// ConvertRawToInt64 convert raw int64
// 如果val为指针值，尝试将其转换成*int64，否则转换成int64
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToInt64(val any) (ret int64, err *cd.Error) { return impl.ConvertRawToInt64(val) }

// ConvertRawToUint convert raw uint
// 如果val为指针值，尝试将其转换成*uint，否则转换成uint
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint(val any) (ret uint, err *cd.Error) { return impl.ConvertRawToUint(val) }

// ConvertRawToUint8 convert raw uint8
// 如果val为指针值，尝试将其转换成*uint8，否则转换成uint8
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint8(val any) (ret uint8, err *cd.Error) { return impl.ConvertRawToUint8(val) }

// ConvertRawToUint16 convert raw uint16
// 如果val为指针值，尝试将其转换成*uint16，否则转换成uint16
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint16(val any) (ret uint16, err *cd.Error) { return impl.ConvertRawToUint16(val) }

// ConvertRawToUint32 convert raw uint32
// 如果val为指针值，尝试将其转换成*uint32，否则转换成uint32
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint32(val any) (ret uint32, err *cd.Error) { return impl.ConvertRawToUint32(val) }

// ConvertRawToUint64 convert raw uint64
// 如果val为指针值，尝试将其转换成*uint64，否则转换成uint64
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToUint64(val any) (ret uint64, err *cd.Error) { return impl.ConvertRawToUint64(val) }

// ConvertRawToFloat32 convert raw float32
// 如果val为指针值，尝试将其转换成*float32，否则转换成float32
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToFloat32(val any) (ret float32, err *cd.Error) { return impl.ConvertRawToFloat32(val) }

// ConvertRawToFloat64 convert raw float64
// 如果val为指针值，尝试将其转换成*float64，否则转换成float64
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToFloat64(val any) (ret float64, err *cd.Error) { return impl.ConvertRawToFloat64(val) }

// ConvertRawToString convert raw string
// 如果val为指针值，尝试将其转换成*string，否则转换成string
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToString(val any) (ret string, err *cd.Error) { return impl.ConvertRawToString(val) }

// ConvertRawToDateTime convert raw datetime
// 如果val为指针值，尝试将其转换成*time.Time，否则转换成time.Time
// 将转换后的结果以models.RawVal形式返回
// 转换出错返回*cd.Error
func ConvertRawToDateTime(val any) (ret time.Time, err *cd.Error) {
	return impl.ConvertRawToDateTime(val)
}

// ConvertToDateTime convert datetime
// 字符串支持带偏移的 RFC3339、UTC 毫秒文本及 UTC 日期；不使用宿主机本地时区。
// time.Time 和带偏移的输入均规范为 UTC，持久化精度仍为毫秒。
func ConvertToDateTime(rVal reflect.Value) (ret time.Time, err *cd.Error) {
	return impl.ConvertToDateTime(rVal)
}

// IsReallyZeroValue 判断一个基本数值类型是否为零值
// 1. 必须是合法的基本数值类型
// 2. 如果是指针类型，则判断该指针指向的实际值是否为零值
func IsReallyZeroValue(val any) bool { return impl.IsReallyZeroValue(val) }

// ConstraintErrorDetails preserves custom validator reasons as well as the
// directive identity supplied by the built-in validator. It never formats val.
func ConstraintErrorDetails(err error) (key models.Key, message string) {
	return impl.ConstraintErrorDetails(err)
}

func FieldValidationError(field string, err error) *cd.Error {
	return impl.FieldValidationError(field, err)
}

func NewValueValidator() models.ValueValidator { return impl.NewValueValidator() }
