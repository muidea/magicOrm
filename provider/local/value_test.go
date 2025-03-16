package local

import (
	"reflect"
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
)

func TestArray(t *testing.T) {
	type Demo struct {
		IntArray  []int
		StrArray  []string
		StrPtr    *string
		NotStrPtr *string
	}

	strVal := "abc"
	demo := &Demo{IntArray: []int{}, NotStrPtr: &strVal}
	if demo.IntArray == nil {
		t.Errorf("IntArray is nil")
	}
	if demo.StrArray != nil {
		t.Errorf("StrArray is not nil")
	}

	if demo.StrPtr != nil {
		t.Errorf("StrPtr is not nil")
	}

	rDemoVal := reflect.Indirect(reflect.ValueOf(demo))
	rIntArray := rDemoVal.FieldByName("IntArray")
	rStrArray := rDemoVal.FieldByName("StrArray")
	rStrPtr := rDemoVal.FieldByName("StrPtr")
	rNotStrPtr := rDemoVal.FieldByName("NotStrPtr")
	log.Infof("rIntArray isValid:%v", rIntArray.IsValid())
	log.Infof("rIntArray isNil:%v", rIntArray.IsNil())
	log.Infof("rIntArray isZero:%v", rIntArray.IsZero())
	log.Infof("rIntArray raw---------------")
	rIntArrayPtr := NewValue(rIntArray)
	rIntArrayPtr.reset(true)
	log.Infof("rIntArray isValid:%v", rIntArray.IsValid())
	log.Infof("rIntArray isNil:%v", rIntArray.IsNil())
	log.Infof("rIntArray isZero:%v", rIntArray.IsZero())
	log.Infof("rIntArray reset(true)---------------")
	rIntArrayPtr.reset(false)
	log.Infof("rIntArray isValid:%v", rIntArray.IsValid())
	log.Infof("rIntArray isNil:%v", rIntArray.IsNil())
	log.Infof("rIntArray isZero:%v", rIntArray.IsZero())
	log.Infof("rIntArray reset(false)---------------")
	log.Infof("################################################")

	log.Infof("rStrArray isValid:%v", rStrArray.IsValid())
	log.Infof("rStrArray isNil:%v", rStrArray.IsNil())
	log.Infof("rStrArray isZero:%v", rStrArray.IsZero())
	log.Infof("rStrArray raw---------------")
	rStrArrayPtr := NewValue(rStrArray)
	rStrArrayPtr.reset(true)
	log.Infof("rStrArray isValid:%v", rStrArray.IsValid())
	log.Infof("rStrArray isNil:%v", rStrArray.IsNil())
	log.Infof("rStrArray isZero:%v", rStrArray.IsZero())
	log.Infof("rStrArray reset(true)---------------")

	log.Infof("################################################")

	log.Infof("rStrPtr isValid:%v", rStrPtr.IsValid())
	log.Infof("rStrPtr isNil:%v", rStrPtr.IsNil())
	log.Infof("rStrPtr isZero:%v", rStrPtr.IsZero())
	log.Infof("rStrPtr raw---------------")
	rStrPtrPtr := NewValue(rStrPtr)
	rStrPtrPtr.reset(true)
	log.Infof("rStrPtr isValid:%v", rStrPtr.IsValid())
	log.Infof("rStrPtr isNil:%v", rStrPtr.IsNil())
	log.Infof("rStrPtr isZero:%v", rStrPtr.IsZero())
	log.Infof("rStrPtr reset(true)---------------")

	log.Infof("################################################")

	log.Infof("rNotStrPtr isValid:%v", rNotStrPtr.IsValid())
	log.Infof("rNotStrPtr isNil:%v", rNotStrPtr.IsNil())
	log.Infof("rNotStrPtr isZero:%v", rNotStrPtr.IsZero())
	log.Infof("rNotStrPtr raw---------------")
	rNotStrPtrPtr := NewValue(rNotStrPtr)
	rNotStrPtrPtr.reset(true)
	log.Infof("rNotStrPtr isValid:%v", rNotStrPtr.IsValid())
	log.Infof("rNotStrPtr isNil:%v", rNotStrPtr.IsNil())
	log.Infof("rNotStrPtr isZero:%v", rNotStrPtr.IsZero())
	log.Infof("rNotStrPtr reset(true)---------------")
	rNotStrPtrPtr.reset(false)
	log.Infof("rNotStrPtr isValid:%v", rNotStrPtr.IsValid())
	log.Infof("rNotStrPtr isNil:%v", rNotStrPtr.IsNil())
	log.Infof("rNotStrPtr isZero:%v", rNotStrPtr.IsZero())
	log.Infof("rNotStrPtr reset(false)---------------")

	log.Infof("################################################")
}

func TestValue(t *testing.T) {
	intVal := 10

	unsetReflect := reflect.ValueOf(intVal)
	value := NewValue(unsetReflect)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Set failed, panicked: %v", r)
		}
	}()
	err := value.Set(0)
	if err == nil {
		t.Errorf("Set failed, error is nil")
	}

	if intVal == 0 {
		t.Errorf("Set failed, value is not zero")
	}

	intReflect := reflect.ValueOf(&intVal)
	value = NewValue(intReflect)
	if !value.IsValid() {
		t.Errorf("NewValue failed, is not nil")
	}

	if value.IsZero() {
		t.Errorf("NewValue failed, is not zero")
	}

	err = value.Set(0)
	if err == nil {
		t.Errorf("Set failed, error is not nil")
	}

	if intVal == 0 {
		t.Errorf("Set failed, value is not zero")
	}
	if value.IsZero() {
		t.Errorf("IsZero() for non-zero value should be false")
	}

	rawVal := value.Get()
	switch rawVal.(type) {
	case *int:
		if *(rawVal.(*int)) != intVal {
			t.Errorf("Get failed, expected *int, got %T", rawVal)
		}
	default:
		t.Errorf("Get failed, expected *int, got %T", rawVal)
	}

	var nilValue int
	nilReflect := reflect.ValueOf(&nilValue).Elem()
	value = NewValue(nilReflect)
	if !value.IsValid() {
		t.Errorf("NewValue failed, IsValid false")
	}

	if !value.IsZero() {
		t.Errorf("NewValue failed, IsZero true")
	}

	rawVal = value.Get()
	switch rawVal.(type) {
	case int:
		if rawVal.(int) != 0 {
			t.Errorf("Get failed, expected int, got %T", rawVal)
		}
	default:
		t.Errorf("Get failed, expected int, got %T", rawVal)
	}

	value.Set(10)
	if value.IsZero() {
		t.Errorf("Set failed, IsZero true")
	}
	rawVal = value.Get()
	switch rawVal.(type) {
	case int:
		if rawVal.(int) != 10 {
			t.Errorf("Get failed, expected int, got %T", rawVal)
		}
	default:
		t.Errorf("Get failed, expected int, got %T", rawVal)
	}
}

// TestValueInterface tests the Interface method
func TestValueInterface(t *testing.T) {
	// Test with non-nil value
	iVal := 10
	iReflect := reflect.ValueOf(&iVal).Elem()
	valuePtr := NewValue(iReflect)
	if valuePtr.Get().(int) != iVal {
		t.Errorf("Interface() returned wrong value, expected: %v, got: %v", iVal, valuePtr.Get())
	}
}

// TestValueWithNilPointer tests value methods with nil pointers
func TestValueWithNilPointer(t *testing.T) {
	// Create nil pointer of type *int
	var ptr *int
	valuePtr := NewValue(reflect.ValueOf(ptr))

	// Test IsValid
	if valuePtr.IsValid() {
		t.Errorf("IsValid() for nil pointer should be false (value is invalid if nil)")
	}

	// Test IsZero
	if !valuePtr.IsZero() {
		t.Errorf("IsZero() for nil pointer should be true")
	}

	iVal := 10
	ptr = &iVal
	valuePtr = NewValue(reflect.ValueOf(ptr))
	if !valuePtr.IsValid() {
		t.Errorf("IsValid() for non-nil pointer should be true")
	}
	if valuePtr.IsZero() {
		t.Errorf("IsZero() for non-nil pointer should be false")
	}
}

func TestUnpackValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected []model.Value
	}{
		{
			name:  "slice value",
			value: []int{1, 2, 3},
			expected: []model.Value{
				NewValue(reflect.ValueOf(1)),
				NewValue(reflect.ValueOf(2)),
				NewValue(reflect.ValueOf(3)),
			},
		},
		{
			name:     "non-slice value",
			value:    42,
			expected: []model.Value{NewValue(reflect.ValueOf(42))},
		},
		{
			name:     "empty slice value",
			value:    []int{},
			expected: []model.Value{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueImpl := &ValueImpl{value: reflect.ValueOf(tt.value)}
			actual := valueImpl.UnpackValue()
			if len(actual) != len(tt.expected) {
				t.Errorf("UnpackValue() name:%s = %v, want %v", tt.name, actual, tt.expected)
				return
			}

			for i := range actual {
				if actual[i].Get() != tt.expected[i].Get() {
					t.Errorf("UnpackValue() name:%s = %v, want %v", tt.name, actual, tt.expected)
				}
			}
		})
	}
}

func TestValueImpl_Append(t *testing.T) {
	tests := []struct {
		name    string
		value   reflect.Value
		val     reflect.Value
		wantErr bool
	}{
		{
			name:    "append to slice",
			value:   reflect.Indirect(reflect.ValueOf(&[]int{1, 2})),
			val:     reflect.ValueOf(3),
			wantErr: false,
		},
		{
			name:    "append to non-slice",
			value:   reflect.ValueOf(1),
			val:     reflect.ValueOf(2),
			wantErr: true,
		},
		{
			name:    "type mismatch",
			value:   reflect.ValueOf([]int{1, 2}),
			val:     reflect.ValueOf("3"),
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   reflect.ValueOf([]int{1, 2}),
			val:     reflect.ValueOf(nil),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &ValueImpl{value: tt.value}
			err := v.Append(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValueImpl.Append() name:%s, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if v.value.Len() != 3 {
					t.Errorf("ValueImpl.Append() name:%s, len = %d, want 3", tt.name, v.value.Len())
				}
			}
		})
	}
}
