package model

import "fmt"

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

func (s ValueDeclare) IsDaeTime() bool {
	return s == DateTime
}

func IsCustomer(val ValueDeclare) bool {
	return val == Customer
}

func IsAutoIncrement(val ValueDeclare) bool {
	return val == AutoIncrement
}

func IsUUID(val ValueDeclare) bool {
	return val == UUID
}

func IsSnowFlake(val ValueDeclare) bool {
	return val == SnowFlake
}

func IsDateTime(val ValueDeclare) bool {
	return val == DateTime
}

type TypeDeclare int

// Define the Type enum
const (
	// TypeBooleanValue bool
	TypeBooleanValue = iota + 100
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
	// TypeMapValue map
	TypeMapValue = 500
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
	case TypeMapValue:
		return "map"
	default:
		return fmt.Sprintf("illegal type %d", s)
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

func (s TypeDeclare) IsMapType() bool {
	return s == TypeMapValue
}

func IsBasicType(typeValue TypeDeclare) bool {
	return typeValue < TypeStructValue
}

func IsDateTimeType(typeValue TypeDeclare) bool {
	return typeValue == TypeDateTimeValue
}

func IsStructType(typeValue TypeDeclare) bool {
	return typeValue == TypeStructValue
}

func IsSliceType(typeValue TypeDeclare) bool {
	return typeValue == TypeSliceValue
}

func IsMapType(typeVal TypeDeclare) bool {
	return typeVal == TypeMapValue
}

func IsBasicSlice(tType Type) bool {
	return tType.IsBasic() && IsSliceType(tType.GetValue())
}

func IsStructSlice(tType Type) bool {
	return !tType.IsBasic() && IsSliceType(tType.GetValue())
}

type ViewDeclare int

const (
	OriginView = 0
	FullView   = 1
	LiteView   = 2
)

func (s ViewDeclare) String() string {
	switch s {
	case OriginView:
		return "origin"
	case FullView:
		return "detail"
	case LiteView:
		return "lite"
	default:
		return "unknown"
	}
}
