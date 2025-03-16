package model

import "fmt"

// 基础信息定义

// 基本数值类型
// 基本数值类型包括:
// bool、int8、int16、int32、int、int64、uint8、uint16、uint32、uint、uint64、float32、float64、string、dateTime
// []bool、[]int8、[]int16、[]int32、[]int、[]int64、[]uint8、[]uint16、[]uint32、[]uint、[]uint64、[]float32、[]float64、[]string、[]dateTime
// 以及以上类型对应的指针
// *bool, *int8, *int16, *int32, *int, *int64, *uint8, *uint16, *uint32, *uint, *uint64, *float32, *float64, *string, *dateTime
// []*bool、[]*int8、[]*int16、[]*int32、[]*int、[]*int64、[]*uint8、[]*uint16、[]*uint32、[]*uint、[]*uint64、[]*float32、[]*float64、[]*string、[]*dateTime
// *[]bool, *[]int8, *[]int16, *[]int32, *[]int, *[]int64, *[]uint8, *[]uint16, *[]uint32, *[]uint, *[]uint64, *[]float32, *[]float64, *[]string, *[]dateTime
// *[]*bool、*[]*int8、*[]*int16、*[]*int32、*[]*int、*[]*int64、*[]*uint8、*[]*uint16、*[]*uint32、*[]*uint、*[]*uint64、*[]*float32、*[]*float64、*[]*string、*[]*dateTime

// 复合数值类型
// 复合数值类型包括:
// struct, []struct
// 以及对应的指针
// *struct, []*struct, *[]struct，*[]*struct

// 基本数据值申明
// 基本数据值申明包括:
// autoIncrement, uuid, snowFlake, dateTime
// autoIncrement: 自增长，数值类型是int64，系统自动赋值
// uuid: 唯一标识，数值类型是string，系统自动赋值
// snowFlake: 雪花算法，数值类型是int64，系统自动赋值

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

// Define the Type enum
const (
	// TypeBooleanValue bool
	TypeBooleanValue TypeDeclare = iota + 100
	// TypeBitValue int8
	TypeBitValue
	// TypeSmallIntegerValue int16
	TypeSmallIntegerValue
	// TypeInteger32Value int32
	TypeInteger32Value
	// TypeIntegerValue int
	TypeIntegerValue
	// TypeBigIntegerValue int64
	TypeBigIntegerValue
	// TypePositiveBitValue uint8
	TypePositiveBitValue
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

func (s TypeDeclare) String() string {
	switch s {
	case TypeBooleanValue:
		return "bool"
	case TypeBitValue:
		return "int8"
	case TypeSmallIntegerValue:
		return "int16"
	case TypeInteger32Value:
		return "int32"
	case TypeIntegerValue:
		return "int"
	case TypeBigIntegerValue:
		return "int64"
	case TypePositiveBitValue:
		return "uint8"
	case TypePositiveSmallIntegerValue:
		return "uint16"
	case TypePositiveInteger32Value:
		return "uint32"
	case TypePositiveIntegerValue:
		return "uint"
	case TypePositiveBigIntegerValue:
		return "uint64"
	case TypeFloatValue:
		return "float32"
	case TypeDoubleValue:
		return "float64"
	case TypeStringValue:
		return "string"
	case TypeDateTimeValue:
		return "time"
	case TypeStructValue:
		return "struct"
	case TypeSliceValue:
		return "array"
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

type ValueDeclare int

const (
	Customer ValueDeclare = iota
	AutoIncrement
	UUID
	SnowFlake
	DateTime
)

func (s ValueDeclare) String() string {
	switch s {
	case AutoIncrement:
		return "autoIncrement"
	case UUID:
		return "uuid"
	case SnowFlake:
		return "snowFlake"
	case DateTime:
		return "dateTime"
	default:
		return "customer"
	}
}

func (s ValueDeclare) IsCustomer() bool {
	return s == Customer
}

func (s ValueDeclare) IsAutoIncrement() bool {
	return s == AutoIncrement
}

func (s ValueDeclare) IsUUID() bool {
	return s == UUID
}

func (s ValueDeclare) IsSnowFlake() bool {
	return s == SnowFlake
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
	// LiteView 简单数据，在类型定义时需要主动定义
	LiteView = "lite"
)
