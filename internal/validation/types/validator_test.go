package types

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type customStringHandler struct{}

func (h *customStringHandler) Validate(value any) error {
	if _, ok := value.(string); !ok {
		return errors.New("not a string")
	}
	return nil
}
func (h *customStringHandler) Convert(value any) (any, error) {
	if v, ok := value.(string); ok {
		return "custom:" + v, nil
	}
	return nil, errors.New("cannot convert")
}
func (h *customStringHandler) GetZeroValue() any     { return "custom-zero" }
func (h *customStringHandler) GetType() reflect.Type { return reflect.TypeOf("") }

func TestTypeValidatorValidateAndConvert(t *testing.T) {
	validator := NewTypeValidator()

	if err := validator.ValidateType("value", reflect.TypeOf("")); err != nil {
		t.Fatalf("expected direct string validation success, got %v", err)
	}
	if err := validator.ValidateType(nil, reflect.TypeOf((*string)(nil))); err != nil {
		t.Fatalf("expected nil pointer validation success, got %v", err)
	}
	if err := validator.ValidateType(struct{}{}, reflect.TypeOf("")); err == nil {
		t.Fatal("expected incompatible validation failure")
	}
	if err := validator.ValidateType("ignored", nil); err != nil {
		t.Fatalf("expected nil field type to be ignored, got %v", err)
	}

	converted, err := validator.Convert(int32(7), reflect.TypeOf(int64(0)))
	if err != nil || converted.(int64) != 7 {
		t.Fatalf("unexpected numeric conversion result: %v %v", converted, err)
	}

	ptrValue, err := validator.Convert("name", reflect.TypeOf((*string)(nil)))
	if err != nil || *(ptrValue.(*string)) != "name" {
		t.Fatalf("unexpected pointer conversion result: %v %v", ptrValue, err)
	}

	value := "alias"
	unwrapped, err := validator.Convert(&value, reflect.TypeOf(""))
	if err != nil || unwrapped.(string) != "alias" {
		t.Fatalf("unexpected pointer unwrap result: %v %v", unwrapped, err)
	}

	if _, err := validator.Convert("fail", nil); err == nil {
		t.Fatal("expected nil target type conversion failure")
	}
}

func TestTypeValidatorHandlers(t *testing.T) {
	validator := NewTypeValidator()
	handler := &customStringHandler{}

	if err := validator.RegisterTypeHandler("custom-string", handler); err != nil {
		t.Fatalf("register type handler failed: %v", err)
	}

	converted, err := validator.Convert("value", handler.GetType())
	if err != nil || converted.(string) != "value" {
		t.Fatalf("expected direct assignment to bypass handler conversion, got %v %v", converted, err)
	}

	supportedTypes := validator.GetSupportedTypes()
	if len(supportedTypes) == 0 {
		t.Fatal("expected built-in supported types")
	}
	if zero := validator.GetZeroValue(handler.GetType()); zero != "custom-zero" {
		t.Fatalf("unexpected custom zero value: %v", zero)
	}
	if zero := validator.GetZeroValue(nil); zero != nil {
		t.Fatalf("expected nil zero value for nil type, got %v", zero)
	}
}

func TestBuiltinHandlers(t *testing.T) {
	basicHandler := &basicTypeHandler{typ: reflect.TypeOf(int64(0))}
	if err := basicHandler.Validate(int32(1)); err != nil {
		t.Fatalf("expected basic handler validation success, got %v", err)
	}
	if _, err := basicHandler.Convert(int32(2)); err != nil {
		t.Fatalf("expected basic handler conversion success, got %v", err)
	}
	if basicHandler.GetZeroValue().(int64) != 0 || basicHandler.GetType() != reflect.TypeOf(int64(0)) {
		t.Fatal("unexpected basic handler metadata")
	}

	timeHandler := &timeTypeHandler{typ: reflect.TypeOf(time.Time{})}
	if err := timeHandler.Validate(time.Now()); err != nil {
		t.Fatalf("expected time validation success, got %v", err)
	}
	if err := timeHandler.Validate("2024-01-02"); err != nil {
		t.Fatalf("expected date string validation success, got %v", err)
	}
	if err := timeHandler.Validate("bad"); err == nil {
		t.Fatal("expected invalid time string failure")
	}
	if converted, err := timeHandler.Convert("2024-01-02 15:04:05"); err != nil || converted.(time.Time).Year() != 2024 {
		t.Fatalf("unexpected time conversion result: %v %v", converted, err)
	}
	if _, err := timeHandler.Convert(1); err == nil {
		t.Fatal("expected invalid time conversion failure")
	}
	if timeHandler.GetZeroValue().(time.Time).IsZero() == false || timeHandler.GetType() != reflect.TypeOf(time.Time{}) {
		t.Fatal("unexpected time handler metadata")
	}

	sliceHandler := &sliceTypeHandler{typ: reflect.TypeOf([]int{})}
	if err := sliceHandler.Validate([]int32{1, 2}); err != nil {
		t.Fatalf("expected slice validation success, got %v", err)
	}
	if err := sliceHandler.Validate("bad"); err == nil {
		t.Fatal("expected non-slice validation failure")
	}
	if converted, err := sliceHandler.Convert([]int32{1, 2}); err != nil || !reflect.DeepEqual(converted, []int{1, 2}) {
		t.Fatalf("unexpected slice conversion result: %v %v", converted, err)
	}
	if sliceHandler.GetType() != reflect.TypeOf([]int{}) {
		t.Fatal("unexpected slice handler type")
	}
	if !reflect.DeepEqual(sliceHandler.GetZeroValue(), []int(nil)) {
		t.Fatalf("unexpected slice zero value: %#v", sliceHandler.GetZeroValue())
	}

	pointerHandler := &pointerTypeHandler{typ: reflect.TypeOf((*int)(nil))}
	if err := pointerHandler.Validate(1); err != nil {
		t.Fatalf("expected pointer validation success, got %v", err)
	}
	if err := pointerHandler.Validate(struct{}{}); err == nil {
		t.Fatal("expected pointer validation failure")
	}
	if converted, err := pointerHandler.Convert(1); err != nil || *(converted.(*int)) != 1 {
		t.Fatalf("unexpected pointer conversion result: %v %v", converted, err)
	}
	if converted, err := pointerHandler.Convert(nil); err != nil || converted != nil {
		t.Fatalf("unexpected nil pointer conversion result: %v %v", converted, err)
	}
	if pointerHandler.GetZeroValue() != nil || pointerHandler.GetType() != reflect.TypeOf((*int)(nil)) {
		t.Fatal("unexpected pointer handler metadata")
	}
}
