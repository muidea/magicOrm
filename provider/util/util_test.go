package util

import (
	"log"
	"reflect"
	"testing"
)

func TestNilValue(t *testing.T) {
	var val reflect.Value
	if !IsNil(val) {
		t.Errorf("Check val is nil failed")
		return
	}

	var iVal int
	log.Printf("IsNil(reflect.ValueOf(iVal)), val:%v", iVal)
	if IsNil(reflect.ValueOf(iVal)) {
		t.Errorf("Check int is nil failed")
		return
	}
	log.Printf("IsNil(reflect.ValueOf(&iVal)), val:%v", &iVal)
	if IsNil(reflect.ValueOf(&iVal)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	var iValPtr *int
	log.Printf("!IsNil(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	if !IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	iValPtr = &iVal
	log.Printf("IsNil(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	if IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&iValPtr)), val:%v", &iValPtr)
	if IsNil(reflect.ValueOf(&iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	var interfaceVal interface{}
	log.Printf("!IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	interfaceVal = iVal
	log.Printf("IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	if IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	interfaceVal = nil
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	var arrayIntVal []int
	log.Printf("IsNil(reflect.ValueOf(arrayIntVal)), val:%v", arrayIntVal)
	if IsNil(reflect.ValueOf(arrayIntVal)) {
		t.Errorf("Check arrayIntVal is nil failed")
		return
	}

	var arrayIntInterfaceVal interface{}
	arrayIntInterfaceVal = arrayIntVal
	log.Printf("IsNil(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	if IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal interface is nil failed")
		return
	}

	arrayIntInterfaceVal = &arrayIntVal
	log.Printf("IsNil(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	if IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil failed")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&arrayIntInterfaceVal)), val:%v", &arrayIntInterfaceVal)
	if IsNil(reflect.ValueOf(&arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil failed")
		return
	}

	var mapVal map[string]string
	log.Printf("IsNil(reflect.ValueOf(mapVal)), val:%v", mapVal)
	if IsNil(reflect.ValueOf(mapVal)) {
		t.Errorf("Check mapVal is nil failed")
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
	if IsNil(intVal) {
		t.Errorf("Check intVal is nil failed")
		return
	}

	ptrVal := dv.FieldByName("PtrVal")
	log.Printf("!IsNil(ptrVal), val:%v", ptrVal.Interface())
	if !IsNil(ptrVal) {
		t.Errorf("Check ptrVal is nil failed")
		return
	}

	interfaceVal := dv.FieldByName("InterfaceVal")
	log.Printf("!IsNil(interfaceVal), val:%v", interfaceVal.Interface())
	if !IsNil(interfaceVal) {
		t.Errorf("Check interfaceVal is nil failed")
		return
	}

	arrayVal := dv.FieldByName("ArrayVal")
	log.Printf("IsNil(arrayVal), val:%v", arrayVal.Interface())
	if IsNil(arrayVal) {
		t.Errorf("Check arrayVal is nil failed")
		return
	}

	ii := 10
	demo2 := Demo{PtrVal: &ii}
	dv2 := reflect.ValueOf(demo2)
	intVal2 := dv2.FieldByName("IntVal")
	log.Printf("IsNil(intVal2), val:%v", intVal2.Interface())
	if IsNil(intVal2) {
		t.Errorf("Check intVal2 is nil failed")
		return
	}

	ptrVal2 := dv2.FieldByName("PtrVal")
	log.Printf("IsNil(ptrVal2), val:%v", ptrVal2.Interface())
	if IsNil(ptrVal2) {
		t.Errorf("Check ptrVal2 is nil failed")
		return
	}

	interfaceVal2 := dv2.FieldByName("InterfaceVal")
	log.Printf("!IsNil(interfaceVal2), val:%v", interfaceVal2.Interface())
	if !IsNil(interfaceVal2) {
		t.Errorf("Check interfaceVal2 is nil failed")
		return
	}

	arrayVal2 := dv2.FieldByName("ArrayVal")
	log.Printf("IsNil(arrayVal2), val:%v", arrayVal2.Interface())
	if IsNil(arrayVal2) {
		t.Errorf("Check arrayVal2 is nil failed")
		return
	}
}
