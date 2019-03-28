package remote

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestSimpleObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240, Add: []int{12, 34, 45}}

	objVal, objErr := GetObjectValue(obj)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	data, err := json.Marshal(objVal)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}

	log.Print(*objVal)

	log.Print(string(data))

	val := &ObjectValue{}
	err = json.Unmarshal(data, val)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}
	log.Print(*val)

	objValue := reflect.Indirect(reflect.ValueOf(val))
	log.Printf("objValue canSet:%v", objValue.CanSet())
	itemsVal := objValue.FieldByName("Items")
	log.Printf("itemsVal canSet:%v", itemsVal.CanSet())

	nameVal := itemsVal.Index(0).FieldByName("Value")
	nameVal.Set(reflect.ValueOf("aabbccdd"))

	log.Print(itemsVal.Interface())

	log.Print(*val)
}

func TestExtObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc}
	ext := &ExtObj{Name: "extObj", Obj: obj}

	objVal, objErr := GetObjectValue(ext)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	data, err := json.Marshal(&objVal)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}

	log.Print(objVal)

	log.Print(string(data))
}
