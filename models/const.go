package models

import "fmt"

// 基础信息定义

// 基本数值类型
// 基本数值类型包括:
// bool、int8、int16、int32、int、int64、uint8、uint16、uint32、uint、uint64、float32、float64、string、dateTime
// []bool、[]int8、[]int16、[]int32、[]int、[]int64、[]uint8、[]uint16、[]uint32、[]uint、[]uint64、[]float32、[]float64、[]string、[]datetime
// 以及以上类型对应的指针
// *bool, *int8, *int16, *int32, *int, *int64, *uint8, *uint16, *uint32, *uint, *uint64, *float32, *float64, *string, *datetime
// []*bool、[]*int8、[]*int16、[]*int32、[]*int、[]*int64、[]*uint8、[]*uint16、[]*uint32、[]*uint、[]*uint64、[]*float32、[]*float64、[]*string、[]*datetime
// *[]bool, *[]int8, *[]int16, *[]int32, *[]int, *[]int64, *[]uint8, *[]uint16, *[]uint32, *[]uint, *[]uint64, *[]float32, *[]float64, *[]string, *[]datetime
// *[]*bool、*[]*int8、*[]*int16、*[]*int32、*[]*int、*[]*int64、*[]*uint8、*[]*uint16、*[]*uint32、*[]*uint、*[]*uint64、*[]*float32、*[]*float64、*[]*string、*[]*datetime

// 复合数值类型
// 复合数值类型包括:
// struct, []struct
// 以及对应的指针
// *struct, []*struct, *[]struct，*[]*struct

// 基本数据值申明
// 基本数据值申明包括:
// autoIncrement, uuid, snowflake, datetime
// autoIncrement: 自增长，数值类型是int64，系统自动赋值
// uuid: 唯一标识，数值类型是string，系统自动赋值
// snowflake: 雪花算法，数值类型是int64，系统自动赋值

// 基本数据视图申明
// 基本数据视图申明包括:
// origin, detail, lite
// 基本数据视图定义
// 基本数据视图定义包括:
// origin: 原始数据，包括所有字段
// detail: 详细数据，在类型定义时需要主动定义
// lite: 精简数据，在类型定义时需要主动定义

// 在初始化一个新模型对象，需要主动定义视图，否则默认为原始数据
// 初始化模型对象，可以通过使用以下方法
// 1. 通过provider.GetEntityModel(entity)，entity为实体，该方法默认产生origin模型对象
// 2. 通过provider.GetTypeModel(Type, ViewDeclare)方法, 新建指定类型的模型对象，Type为类型定义，ViewDeclare为视图定义
// 3. 通过Model.Copy(reset, viewSpec)方法，复制模型对象，reset为是否重置模型对象，viewSpec为视图定义

type TypeDeclare int

const (
	TypeBooleanName              = "bool"
	TypeByteName                 = "int8"
	TypeSmallIntegerName         = "int16"
	TypeInteger32Name            = "int32"
	TypeIntegerName              = "int"
	TypeBigIntegerName           = "int64"
	TypePositiveByteName         = "uint8"
	TypePositiveSmallIntegerName = "uint16"
	TypePositiveInteger32Name    = "uint32"
	TypePositiveIntegerName      = "uint"
	TypePositiveBigIntegerName   = "uint64"
	TypeFloatName                = "float32"
	TypeDoubleName               = "float64"
	TypeStringName               = "string"
	TypeDateTimeName             = "datetime"
	TypeStructTimeName           = "time.Time"
	TypeStructName               = "struct"
	TypeSliceName                = "array"
)

// Define the Type enum
const (
	// TypeBooleanValue bool
	TypeBooleanValue TypeDeclare = iota + 100
	// TypeByteValue int8
	TypeByteValue
	// TypeSmallIntegerValue int16
	TypeSmallIntegerValue
	// TypeInteger32Value int32
	TypeInteger32Value
	// TypeIntegerValue int
	TypeIntegerValue
	// TypeBigIntegerValue int64
	TypeBigIntegerValue
	// TypePositiveByteValue uint8
	TypePositiveByteValue
	// TypePositiveSmallIntegerValue uint16
	TypePositiveSmallIntegerValue
	// TypePositiveInteger32Value uint32
	TypePositiveInteger32Value
	// TypePositiveIntegerValue uint
	TypePositiveIntegerValue
	// TypePositiveBigIntegerValue uint64
	TypePositiveBigIntegerValue
	// TypeFloatValue float32
	TypeFloatValue
	// TypeDoubleValue float64
	TypeDoubleValue
	// TypeStringValue string
	TypeStringValue
	// TypeDateTimeValue time.Time
	TypeDateTimeValue
	// TypeStructValue struct
	TypeStructValue
	// TypeSliceValue slice
	TypeSliceValue
)

var typeName2ValueMap = map[string]TypeDeclare{
	TypeBooleanName:              TypeBooleanValue,
	TypeByteName:                 TypeByteValue,
	TypeSmallIntegerName:         TypeSmallIntegerValue,
	TypeInteger32Name:            TypeInteger32Value,
	TypeIntegerName:              TypeIntegerValue,
	TypeBigIntegerName:           TypeBigIntegerValue,
	TypePositiveByteName:         TypePositiveByteValue,
	TypePositiveSmallIntegerName: TypePositiveSmallIntegerValue,
	TypePositiveInteger32Name:    TypePositiveInteger32Value,
	TypePositiveIntegerName:      TypePositiveIntegerValue,
	TypePositiveBigIntegerName:   TypePositiveBigIntegerValue,
	TypeFloatName:                TypeFloatValue,
	TypeDoubleName:               TypeDoubleValue,
	TypeStringName:               TypeStringValue,
	TypeDateTimeName:             TypeDateTimeValue,
	TypeStructTimeName:           TypeDateTimeValue,
	TypeSliceName:                TypeSliceValue,
}

func (s TypeDeclare) String() string {
	switch s {
	case TypeBooleanValue:
		return TypeBooleanName
	case TypeByteValue:
		return TypeByteName
	case TypeSmallIntegerValue:
		return TypeSmallIntegerName
	case TypeInteger32Value:
		return TypeInteger32Name
	case TypeIntegerValue:
		return TypeIntegerName
	case TypeBigIntegerValue:
		return TypeBigIntegerName
	case TypePositiveByteValue:
		return TypePositiveByteName
	case TypePositiveSmallIntegerValue:
		return TypePositiveSmallIntegerName
	case TypePositiveInteger32Value:
		return TypePositiveInteger32Name
	case TypePositiveIntegerValue:
		return TypePositiveIntegerName
	case TypePositiveBigIntegerValue:
		return TypePositiveBigIntegerName
	case TypeFloatValue:
		return TypeFloatName
	case TypeDoubleValue:
		return TypeDoubleName
	case TypeStringValue:
		return TypeStringName
	case TypeDateTimeValue:
		return TypeDateTimeName
	case TypeStructValue:
		return TypeStructName
	case TypeSliceValue:
		return TypeSliceName
	default:
		return fmt.Sprintf("illegal type decare value %d", s)
	}
}

func (s TypeDeclare) IsBasicType() bool {
	return s < TypeStructValue
}

func (s TypeDeclare) IsDateTimeType() bool {
	return s == TypeDateTimeValue
}

func (s TypeDeclare) IsStructType() bool {
	return s == TypeStructValue
}

func (s TypeDeclare) IsSliceType() bool {
	return s == TypeSliceValue
}

func (s TypeDeclare) IsStringValueType() bool {
	return s == TypeStringValue || s == TypeDateTimeValue
}

func (s TypeDeclare) IsNumberValueType() bool {
	return s > TypeBooleanValue && s <= TypePositiveBigIntegerValue
}

type ValueDeclare string

const (
	Customer      = ""
	AutoIncrement = "auto"
	UUID          = "uuid"
	Snowflake     = "snowflake"
	DateTime      = "datetime"
)

func (s ValueDeclare) IsCustomer() bool {
	return s == Customer
}

func (s ValueDeclare) IsAutoIncrement() bool {
	return s == AutoIncrement
}

func (s ValueDeclare) IsUUID() bool {
	return s == UUID
}

func (s ValueDeclare) IsSnowflake() bool {
	return s == Snowflake
}

func (s ValueDeclare) IsDateTime() bool {
	return s == DateTime
}

type ViewDeclare string

const (
	// OriginView 当前数据, 根据MaskValue定义的字段
	// 如果MaskValue为空，则默认返回OriginView
	OriginView = "origin"
	// MetaView 原始数据，包括所有字段元数据，字段值为初始化值
	MetaView = "meta"
	// DetailView 详细数据，在类型定义时需要主动定义
	DetailView = "detail"
	// ListView 列表数据，在类型定义时需要主动定义
	BasicView = "basic"
	// LiteView 简单数据，在类型定义时需要主动定义
	LiteView = "lite"
)

const (
	Key = "key"
)
