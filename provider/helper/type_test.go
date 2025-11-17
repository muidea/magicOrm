package helper

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/remote"
)

// TestStruct is a simple test struct for type tests
type TestStruct struct {
	ID   int    `orm:"id,key"`
	Name string `orm:"name"`
	Age  int    `orm:"age"`
}

func TestTypeInterface(t *testing.T) {
	tests := []struct {
		name         string
		typeImpl     *remote.TypeImpl
		initVal      interface{}
		expectedType interface{}
		expectError  bool
	}{
		{
			name:         "Boolean Interface",
			typeImpl:     &remote.TypeImpl{Name: "bool", Value: models.TypeBooleanValue},
			initVal:      true,
			expectedType: true,
			expectError:  false,
		},
		{
			name:         "Integer Interface",
			typeImpl:     &remote.TypeImpl{Name: "int", Value: models.TypeIntegerValue},
			initVal:      123,
			expectedType: int(123),
			expectError:  false,
		},
		{
			name:         "Float Interface",
			typeImpl:     &remote.TypeImpl{Name: "float64", Value: models.TypeFloatValue},
			initVal:      float32(123.45),
			expectedType: float32(123.45),
			expectError:  false,
		},
		{
			name:         "String Interface",
			typeImpl:     &remote.TypeImpl{Name: "string", Value: models.TypeStringValue},
			initVal:      "test",
			expectedType: "test",
			expectError:  false,
		},
		{
			name:         "DateTime Interface",
			typeImpl:     &remote.TypeImpl{Name: "time.Time", Value: models.TypeDateTimeValue},
			initVal:      "2024-01-01T00:00:00Z",
			expectedType: "2024-01-01T00:00:00Z",
			expectError:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test Interface method
			result, err := test.typeImpl.Interface(test.initVal)

			// Check error expectation
			if test.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Skip further checks if error was expected
			if test.expectError {
				return
			}

			// For non-nil results, we need to check the underlying value
			if result == nil {
				t.Errorf("Expected non-nil result for non-nil input")
				return
			}

			// Get the actual value from the result
			actualValue := result.Get()

			// Special handling for time.Time
			if _, ok := test.expectedType.(time.Time); ok {
				if _, isTimeValue := actualValue.(time.Time); !isTimeValue {
					t.Errorf("Expected time.Time type, got %T", actualValue)
				}
				return
			}

			// For basic types, compare the values
			switch expected := test.expectedType.(type) {
			case bool:
				// Compare boolean values
				if actual, ok := actualValue.(bool); !ok || actual != expected {
					t.Errorf("Boolean value mismatch: expected %v (%T), got %v (%T)", expected, expected, actualValue, actualValue)
				}
			case int:
				// Compare integer values
				// The actual value might be int64 due to reflection
				if actual, ok := actualValue.(int); !ok {
					// Try int64 if not exact int
					if actual64, ok64 := actualValue.(int64); !ok64 || int(actual64) != expected {
						t.Errorf("Integer value mismatch: expected %v (%T), got %v (%T)", expected, expected, actualValue, actualValue)
					}
				} else if actual != expected {
					t.Errorf("Integer value mismatch: expected %v, got %v", expected, actual)
				}
			case float32:
				// Compare float values
				if actual, ok := actualValue.(float32); !ok || actual != expected {
					t.Errorf("Float value mismatch: expected %v (%T), got %v (%T)", expected, expected, actualValue, actualValue)
				}
			case string:
				// Compare string values
				if actual, ok := actualValue.(string); !ok || actual != expected {
					t.Errorf("String value mismatch: expected %v (%T), got %v (%T)", expected, expected, actualValue, actualValue)
				}
			default:
				// This should not happen with our test cases
				t.Errorf("Unexpected type in test: %T", expected)
			}
		})
	}
}

func TestTypeConversion(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedType models.TypeDeclare
		expectError  bool
	}{
		{
			name:         "Boolean conversion",
			value:        true,
			expectedType: models.TypeBooleanValue,
			expectError:  false,
		},
		{
			name:         "Int conversion",
			value:        123,
			expectedType: models.TypeIntegerValue,
			expectError:  false,
		},
		{
			name:         "Int8 conversion",
			value:        int8(123),
			expectedType: models.TypeByteValue,
			expectError:  false,
		},
		{
			name:         "Int16 conversion",
			value:        int16(123),
			expectedType: models.TypeSmallIntegerValue,
			expectError:  false,
		},
		{
			name:         "Int32 conversion",
			value:        int32(123),
			expectedType: models.TypeInteger32Value,
			expectError:  false,
		},
		{
			name:         "Int64 conversion",
			value:        int64(123),
			expectedType: models.TypeBigIntegerValue,
			expectError:  false,
		},
		{
			name:         "Uint conversion",
			value:        uint(123),
			expectedType: models.TypePositiveIntegerValue,
			expectError:  false,
		},
		{
			name:         "Uint8 conversion",
			value:        uint8(123),
			expectedType: models.TypePositiveByteValue,
			expectError:  false,
		},
		{
			name:         "Uint16 conversion",
			value:        uint16(123),
			expectedType: models.TypePositiveSmallIntegerValue,
			expectError:  false,
		},
		{
			name:         "Uint32 conversion",
			value:        uint32(123),
			expectedType: models.TypePositiveInteger32Value,
			expectError:  false,
		},
		{
			name:         "Uint64 conversion",
			value:        uint64(123),
			expectedType: models.TypePositiveBigIntegerValue,
			expectError:  false,
		},
		{
			name:         "Float32 conversion",
			value:        float32(123.45),
			expectedType: models.TypeFloatValue,
			expectError:  false,
		},
		{
			name:         "Float64 conversion",
			value:        float64(123.45),
			expectedType: models.TypeDoubleValue,
			expectError:  false,
		},
		{
			name:         "String conversion",
			value:        "test",
			expectedType: models.TypeStringValue,
			expectError:  false,
		},
		{
			name:         "Time conversion",
			value:        time.Now(),
			expectedType: models.TypeDateTimeValue,
			expectError:  false,
		},
		{
			name:         "int slice conversion",
			value:        []int{10, 20},
			expectedType: models.TypeSliceValue,
			expectError:  false,
		},
		{
			name:         "Time slice conversion",
			value:        []time.Time{time.Now(), time.Now()},
			expectedType: models.TypeSliceValue,
			expectError:  false,
		},
		{
			name:         "Struct conversion",
			value:        TestStruct{},
			expectedType: models.TypeStructValue,
			expectError:  false,
		},
		{
			name:        "Nil conversion",
			value:       nil,
			expectError: true,
		},
		{
			name:        "Unsupported type conversion",
			value:       make(chan int),
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typeInfo, err := getEntityType(test.value)

			// Check error expectation
			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for %v, but got none", test.value)
				}
				return
			}

			// If not expecting error, verify we got the right type
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if typeInfo == nil {
				t.Errorf("getEntityType returned nil for %v", test.value)
				return
			}

			if typeInfo.GetValue() != test.expectedType {
				t.Errorf("Type value mismatch: expected %v, got %v", test.expectedType, typeInfo.GetValue())
			}
		})
	}
}

func TestCopyTypeInfo(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "String type",
			value: "test",
		},
		{
			name:  "Integer type",
			value: 123,
		},
		{
			name:  "Float type",
			value: 123.45,
		},
		{
			name:  "Boolean type",
			value: true,
		},
		{
			name:  "DateTime type",
			value: time.Now(),
		},
		{
			name:  "Struct type",
			value: TestStruct{},
		},
		{
			name:  "Slice type",
			value: []string{},
		},
		{
			name:  "Struct pointer type",
			value: &TestStruct{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Get type info for value
			typeInfo, err := getEntityType(test.value)
			if err != nil {
				t.Errorf("getEntityType failed with error: %v", err)
				return
			}

			// Create a copy
			copyInfo := typeInfo.Copy()
			if copyInfo == nil {
				t.Errorf("copy() returned nil")
				return
			}

			// Check if the copy has the same basic properties
			if copyInfo.GetName() != typeInfo.GetName() ||
				copyInfo.GetValue() != typeInfo.GetValue() ||
				copyInfo.GetPkgPath() != typeInfo.GetPkgPath() {
				t.Errorf("Copy failed, basic properties don't match: %s vs %s, %v vs %v, %s vs %s",
					copyInfo.GetName(), typeInfo.GetName(),
					copyInfo.GetValue(), typeInfo.GetValue(),
					copyInfo.GetPkgPath(), typeInfo.GetPkgPath())
			}
		})
	}
}

func TestEntityTypeDetection(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectedErr bool
	}{
		{
			name:        "String type",
			value:       "test",
			expectedErr: false,
		},
		{
			name:        "Integer type",
			value:       123,
			expectedErr: false,
		},
		{
			name:        "Float type",
			value:       123.45,
			expectedErr: false,
		},
		{
			name:        "Boolean type",
			value:       true,
			expectedErr: false,
		},
		{
			name:        "DateTime type",
			value:       time.Now(),
			expectedErr: false,
		},
		{
			name:        "Struct type",
			value:       TestStruct{},
			expectedErr: false,
		},
		{
			name:        "Slice type",
			value:       []string{},
			expectedErr: false,
		},
		{
			name:        "Nil value",
			value:       nil,
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Use getEntityType function
			typeInfo, err := getEntityType(test.value)

			if test.expectedErr {
				if err == nil {
					t.Errorf("Expected error for name:%s,  %v, but got none", test.name, test.value)
				}
			} else {
				if err != nil {
					t.Errorf("Got unexpected error for name:%s,  %v: %v", test.name, test.value, err)
					return
				}

				if typeInfo == nil {
					t.Errorf("getEntityType returned nil for name:%s,  %v", test.name, test.value)
					return
				}

				// Just check that we got a valid type for non-error cases
				if typeInfo.GetValue() == 0 {
					t.Errorf("Expected valid type value, got 0 for name:%s,  %v", test.name, test.value)
				}
			}
		})
	}
}
