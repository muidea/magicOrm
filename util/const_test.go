package util

import (
	"log"
	"reflect"
	"testing"
)

func TestNilValue(t *testing.T) {
	var iVal int
	log.Printf("IsNil(reflect.ValueOf(iVal)), val:%v", iVal)
	if IsNil(reflect.ValueOf(iVal)) {
		t.Errorf("Check int is nil faield")
		return
	}
	log.Printf("IsNil(reflect.ValueOf(&iVal)), val:%v", &iVal)
	if IsNil(reflect.ValueOf(&iVal)) {
		t.Errorf("Check int ptr is nil faield")
		return
	}

	var iValPtr *int
	log.Printf("!IsNil(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	if !IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil faield")
		return
	}

	iValPtr = &iVal
	log.Printf("IsNil(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	if IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil faield")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&iValPtr)), val:%v", &iValPtr)
	if IsNil(reflect.ValueOf(&iValPtr)) {
		t.Errorf("Check int ptr is nil faield")
		return
	}

	var interfaceVal interface{}
	log.Printf("!IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil faield")
		return
	}

	interfaceVal = iVal
	log.Printf("IsNil(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	if IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil faield")
		return
	}

	var arrayIntVal []int
	log.Printf("IsNil(reflect.ValueOf(arrayIntVal)), val:%v", arrayIntVal)
	if IsNil(reflect.ValueOf(arrayIntVal)) {
		t.Errorf("Check arrayIntVal is nil faield")
		return
	}

	var arrayIntInterfaceVal interface{}
	arrayIntInterfaceVal = arrayIntVal
	log.Printf("IsNil(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	if IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal interface is nil faield")
		return
	}

	arrayIntInterfaceVal = &arrayIntVal
	log.Printf("IsNil(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	if IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil faield")
		return
	}

	log.Printf("IsNil(reflect.ValueOf(&arrayIntInterfaceVal)), val:%v", &arrayIntInterfaceVal)
	if IsNil(reflect.ValueOf(&arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil faield")
		return
	}
}
