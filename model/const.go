package model

import "fmt"

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

type ValueDeclare int

const (
	Customer ValueDeclare = iota
	AutoIncrement
	UUID
	SnowFlake
	DateTime
)

func (s ValueDeclare) String() string {
	return fmt.Sprintf("%d", s)
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

func IsBasicType(typeValue TypeDeclare) bool {
	return typeValue < TypeStructValue
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
