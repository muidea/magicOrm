package util

import (
	"log"
	"reflect"
	"testing"
)

type TestVal struct {
	BVal     bool    `json:"bVal"`
	IVal     int     `json:"iVal"`
	I16Val   int16   `json:"i16Val"`
	FVal     float32 `json:"fVal"`
	F64Val   float64 `json:"f64Val"`
	SVal     string  `json:"sVal"`
	ArrayVal []int64 `json:"arrayVal"`
}

func TestNilValue(t *testing.T) {
	var val reflect.Value
	log.Printf("IsNil(val), val:%v", val)
	// nil
	if !IsNil(val) {
		t.Errorf("Check val is nil failed")
		return
	}

	var iVal int
	log.Printf("IsNil(reflect.ValueOf(iVal)), val:%v", iVal)
	// not nil
	if IsNil(reflect.ValueOf(iVal)) {
		t.Errorf("Check int is nil failed")
		return
	}
	log.Printf("IsNil(reflect.ValueOf(&iVal)), val:%v", &iVal)
	// not nil
	if IsNil(reflect.ValueOf(&iVal)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	var iValPtr *int
	log.Printf("!IsNil(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	// nil
	if !IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	iValPtr = &iVal
	log.Printf("IsNil(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	// not nil
	if IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&iValPtr)), val:%v", &iValPtr)
	// not nil
	if IsNil(reflect.ValueOf(&iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	var interfaceVal interface{}
	log.Printf("!IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	// nil
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	interfaceVal = iVal
	log.Printf("IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	// not nil
	if IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	interfaceVal = nil
	log.Printf("IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	// nil
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	var arrayIntVal []int
	log.Printf("IsNil(reflect.ValueOf(arrayIntVal)), val:%v", arrayIntVal)
	// nil
	if !IsNil(reflect.ValueOf(arrayIntVal)) {
		t.Errorf("Check arrayIntVal is nil failed")
		return
	}

	var arrayIntInterfaceVal interface{}
	arrayIntInterfaceVal = arrayIntVal
	log.Printf("IsNil(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	// nil
	if !IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal interface is nil failed")
		return
	}

	arrayIntInterfaceVal = &arrayIntVal
	log.Printf("IsNil(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	// not nil
	if IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil failed")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&arrayIntInterfaceVal)), val:%v", &arrayIntInterfaceVal)
	// not nil
	if IsNil(reflect.ValueOf(&arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil failed")
		return
	}

	var mapVal map[string]string
	log.Printf("IsNil(reflect.ValueOf(mapVal)), val:%v", mapVal)
	// nil
	if !IsNil(reflect.ValueOf(mapVal)) {
		t.Errorf("Check mapVal is nil failed")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&mapVal)), val:%v", &mapVal)
	// not nil
	if IsNil(reflect.ValueOf(&mapVal)) {
		t.Errorf("Check mapVal ptr is nil failed")
		return
	}

	intSlice := []int64{}
	log.Printf("IsNil(reflect.ValueOf(intSlice)), val:%v", &intSlice)
	// not nil
	if IsNil(reflect.ValueOf(intSlice)) {
		t.Errorf("Check intSlice ptr is nil failed")
		return
	}
}

func TestStructNilValue(t *testing.T) {
	type Demo struct {
		IntVal       int
		PtrVal       *int
		InterfaceVal interface{}
		ArrayVal     []int
	}

	demo1 := Demo{}
	dv := reflect.ValueOf(demo1)
	intVal := dv.FieldByName("IntVal")
	log.Printf("IsNil(intVal), val:%v", intVal.Interface())
	// not nil
	if IsNil(intVal) {
		t.Errorf("Check intVal is nil failed")
		return
	}

	ptrVal := dv.FieldByName("PtrVal")
	log.Printf("!IsNil(ptrVal), val:%v", ptrVal.Interface())
	// nil
	if !IsNil(ptrVal) {
		t.Errorf("Check ptrVal is nil failed")
		return
	}

	interfaceVal := dv.FieldByName("InterfaceVal")
	log.Printf("!IsNil(interfaceVal), val:%v", interfaceVal.Interface())
	// nil
	if !IsNil(interfaceVal) {
		t.Errorf("Check interfaceVal is nil failed")
		return
	}

	arrayVal := dv.FieldByName("ArrayVal")
	log.Printf("IsNil(arrayVal), val:%v", arrayVal.Interface())
	// nil
	if !IsNil(arrayVal) {
		t.Errorf("Check arrayVal is nil failed")
		return
	}

	ii := 10
	demo2 := Demo{PtrVal: &ii}
	dv2 := reflect.ValueOf(demo2)
	intVal2 := dv2.FieldByName("IntVal")
	log.Printf("IsNil(intVal2), val:%v", intVal2.Interface())
	// not nil
	if IsNil(intVal2) {
		t.Errorf("Check intVal2 is nil failed")
		return
	}

	ptrVal2 := dv2.FieldByName("PtrVal")
	log.Printf("IsNil(ptrVal2), val:%v", ptrVal2.Interface())
	// not nil
	if IsNil(ptrVal2) {
		t.Errorf("Check ptrVal2 is nil failed")
		return
	}

	interfaceVal2 := dv2.FieldByName("InterfaceVal")
	log.Printf("!IsNil(interfaceVal2), val:%v", interfaceVal2.Interface())
	// nil
	if !IsNil(interfaceVal2) {
		t.Errorf("Check interfaceVal2 is nil failed")
		return
	}

	arrayVal2 := dv2.FieldByName("ArrayVal")
	log.Printf("IsNil(arrayVal2), val:%v", arrayVal2.Interface())
	// nil
	if !IsNil(arrayVal2) {
		t.Errorf("Check arrayVal2 is nil failed")
		return
	}
}
