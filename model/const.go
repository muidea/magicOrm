package model

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

func IsBasicType(typeValue int) bool {
	return typeValue < TypeStructValue
}

func IsStructType(typeValue int) bool {
	return typeValue == TypeStructValue
}

func IsSliceType(typeValue int) bool {
	return typeValue == TypeSliceValue
}

func IsMapType(typeVal int) bool {
	return typeVal == TypeMapValue
}
