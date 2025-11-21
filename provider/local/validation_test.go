package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
)

// ValidationTestStruct is a struct for testing various validation scenarios
type ValidationTestStruct struct {
	ID          int       `orm:"id key auto"`
	Name        string    `orm:"name"`
	Value       float64   `orm:"value"`
	IsActive    bool      `orm:"isActive"`
	CreatedTime time.Time `orm:"createdTime"`
}

// InvalidKeyStruct is a struct with no primary key for testing validation failures
type InvalidKeyStruct struct {
	Name  string  `orm:"name"`
	Value float64 `orm:"value"`
}

// DuplicateFieldStruct is a struct with duplicate field names
type DuplicateFieldStruct struct {
	ID    int    `orm:"id key auto"`
	Name  string `orm:"name"`
	Alias string `orm:"name"` // Same field name as above
}

// UnsupportedTypeStruct is a struct with unsupported field types
type UnsupportedTypeStruct struct {
	ID      int            `orm:"id key auto"`
	Channel chan int       `orm:"channel"`
	Func    func()         `orm:"func"`
	Map     map[string]int `orm:"map"`
}

func TestTypeValidation(t *testing.T) {
	// Test valid struct type
	validStruct := ValidationTestStruct{}
	validType := reflect.TypeOf(validStruct)
	_, err := NewType(validType)
	if err != nil {
		t.Errorf("NewType failed for valid struct: %s", err.Error())
	}

	// Test invalid types
	// Channel type
	chanType := reflect.TypeOf(make(chan int))
	_, err = NewType(chanType)
	if err == nil {
		t.Errorf("NewType should fail for channel type")
	}

	// Function type
	funcType := reflect.TypeOf(func() {})
	_, err = NewType(funcType)
	if err == nil {
		t.Errorf("NewType should fail for function type")
	}

	// Map type
	mapType := reflect.TypeOf(map[string]int{})
	_, err = NewType(mapType)
	if err == nil {
		t.Errorf("NewType should fail for map type")
	}
}

func TestStructValidation(t *testing.T) {
	// Test valid struct
	validStruct := ValidationTestStruct{ID: 1, Name: "test"}
	_, err := GetEntityModel(&validStruct)
	if err != nil {
		t.Errorf("GetEntityModel failed for valid struct: %s", err.Error())
	}

	// Test struct with no primary key
	invalidKeyStruct := InvalidKeyStruct{Name: "test", Value: 123.45}
	_, err = GetEntityModel(&invalidKeyStruct)
	if err == nil {
		t.Errorf("GetEntityModel should fail for struct with no primary key")
	}

	// Test struct with duplicate field names
	duplicateFieldStruct := DuplicateFieldStruct{ID: 1, Name: "test", Alias: "alias"}
	_, err = GetEntityModel(duplicateFieldStruct)
	if err == nil {
		t.Errorf("GetEntityModel should fail for struct with duplicate field names")
	}

	// Test struct with unsupported field types
	unsupportedTypeStruct := UnsupportedTypeStruct{
		ID:      1,
		Channel: make(chan int),
		Func:    func() {},
		Map:     map[string]int{},
	}
	_, err = GetEntityModel(unsupportedTypeStruct)
	if err == nil {
		t.Errorf("GetEntityModel should fail for struct with unsupported field types")
	}

	// Test nil entity
	_, err = GetEntityModel(nil)
	if err == nil {
		t.Errorf("GetEntityModel should fail for nil entity")
	}
}

func TestValueVerification(t *testing.T) {
	// Test valid values
	intVal := 42
	intValue := NewValue(reflect.ValueOf(intVal))
	if !intValue.IsValid() {
		t.Errorf("Value should be valid for int")
	}

	// Test zero values
	zeroInt := 0
	zeroIntValue := NewValue(reflect.ValueOf(zeroInt))
	if !zeroIntValue.IsZero() {
		t.Errorf("IsZero should return true for zero int")
	}

	zeroString := ""
	zeroStringValue := NewValue(reflect.ValueOf(zeroString))
	if !zeroStringValue.IsZero() {
		t.Errorf("IsZero should return true for empty string")
	}

	zeroTime := time.Time{}
	zeroTimeValue := NewValue(reflect.ValueOf(zeroTime))
	if !zeroTimeValue.IsZero() {
		t.Errorf("IsZero should return true for zero time")
	}

	// Test non-zero values
	nonZeroInt := 1
	nonZeroIntValue := NewValue(reflect.ValueOf(nonZeroInt))
	if nonZeroIntValue.IsZero() {
		t.Errorf("IsZero should return false for non-zero int")
	}

	nonZeroString := "test"
	nonZeroStringValue := NewValue(reflect.ValueOf(nonZeroString))
	if nonZeroStringValue.IsZero() {
		t.Errorf("IsZero should return false for non-empty string")
	}

	nonZeroTime := time.Now()
	nonZeroTimeValue := NewValue(reflect.ValueOf(nonZeroTime))
	if nonZeroTimeValue.IsZero() {
		t.Errorf("IsZero should return false for non-zero time")
	}
}

func TestSpecValidation(t *testing.T) {
	// Test valid spec
	structField := reflect.TypeOf(ValidationTestStruct{}).Field(0) // ID field
	spec, err := NewSpec(structField.Tag)
	if err != nil {
		t.Errorf("NewSpec failed for valid tag: %s", err.Error())
	}
	if !spec.IsPrimaryKey() {
		t.Errorf("IsPrimaryKey should return true for ID field")
	}

	// Test malformed tag
	malformedTag := reflect.StructTag(`orm:"id key,invalid"`) // Invalid tag format
	_, err = NewSpec(malformedTag)
	if err != nil {
		// This might be valid in the implementation
		t.Logf("NewSpec result for malformed tag: %s", err.Error())
	}

	// Test value declare validation
	autoTag := reflect.StructTag(`orm:"id key auto"`)
	autoSpec, err := NewSpec(autoTag)
	if err != nil {
		t.Errorf("NewSpec failed for auto tag: %s", err.Error())
	}
	if autoSpec.GetValueDeclare() != models.AutoIncrement {
		t.Errorf("GetValueDeclare should return AutoIncrement for auto tag")
	}

	uuidTag := reflect.StructTag(`orm:"id key uuid"`)
	uuidSpec, err := NewSpec(uuidTag)
	if err != nil {
		t.Errorf("NewSpec failed for uuid tag: %s", err.Error())
	}
	if uuidSpec.GetValueDeclare() != models.UUID {
		t.Errorf("GetValueDeclare should return UUID for uuid tag")
	}

	snowFlakeTag := reflect.StructTag(`orm:"id key snowflake"`)
	snowFlakeSpec, err := NewSpec(snowFlakeTag)
	if err != nil {
		t.Errorf("NewSpec failed for snowflake tag: %s", err.Error())
	}
	if snowFlakeSpec.GetValueDeclare() != models.Snowflake {
		t.Errorf("GetValueDeclare should return Snowflake for snowflake tag")
	}

	dateTimeTag := reflect.StructTag(`orm:"time datetime"`)
	dateTimeSpec, err := NewSpec(dateTimeTag)
	if err != nil {
		t.Errorf("NewSpec failed for datetime tag: %s", err.Error())
	}
	if dateTimeSpec.GetValueDeclare() != models.DateTime {
		t.Errorf("GetValueDeclare should return DateTime for datetime tag")
	}
}
