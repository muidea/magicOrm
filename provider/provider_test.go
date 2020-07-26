package provider

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote"
	"reflect"
	"testing"
)

type Base struct {
	ID    int      `orm:"id key auto"`
	Name  string   `orm:"name"`
	Price float32  `orm:"price"`
	Addr  []string `orm:"addr"`
}

type ExtInfo struct {
	ID    int     `orm:"id key auto"`
	Name  string  `orm:"name"`
	Info  Base    `orm:"info"`
	Ptr   *Base   `orm:"ptr"`
	Array []*Base `orm:"array"`
}

func checkModel(t *testing.T, baseEntityModel model.Model) {
	modelName := baseEntityModel.GetName()
	if modelName != "provider.Base" {
		t.Errorf("get model name failed, curName:%s", modelName)
		return
	}

	fields := baseEntityModel.GetFields()
	if len(fields) != 4 {
		t.Errorf("get model fields failed")
		return
	}

	pkField := baseEntityModel.GetPrimaryField()
	if pkField.GetName() != "ID" {
		t.Errorf("get pk field failed")
		return
	}

	ret := checkIntField(t, pkField)
	if !ret {
		t.Errorf("checkIntField failed")
		return
	}

	for _, val := range fields {
		if val.GetIndex() == 1 {
			ret = checkStringField(t, val)
			if !ret {
				t.Errorf("checkStringField failed")
				return
			}
		}

		if val.GetIndex() == 2 {
			ret = checkFloat32Field(t, val)
			if !ret {
				t.Errorf("checkFloatField failed")
				return
			}
		}

		if val.GetIndex() == 3 {
			ret = checkSliceField(t, val)
			if !ret {
				t.Errorf("checkSliceField failed")
				return
			}
		}
	}

}

func checkIntField(t *testing.T, intField model.Field) (ret bool) {
	fType := intField.GetType()

	ret = checkFieldType(t, fType, "int", nil)
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	if intField.IsAssigned() {
		t.Errorf("check is assigned failed")
		return
	}

	iVal := 10
	rVal := reflect.ValueOf(iVal)
	err := intField.UpdateValue(rVal)
	if err != nil {
		t.Errorf("SetValue failed, err:%s", err.Error())
		return
	}

	if !intField.IsAssigned() {
		t.Error("SetValue failed")
		return
	}

	if intField.GetValue().Get().Int() != 10 {
		t.Error("SetValue failed")
		return
	}

	ret = true
	return
}

func checkFloat32Field(t *testing.T, floatField model.Field) (ret bool) {
	fType := floatField.GetType()

	ret = checkFieldType(t, fType, "float32", nil)
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	if floatField.IsAssigned() {
		t.Errorf("check is assigned failed")
		return
	}

	fVal := float32(12.345)
	rVal := reflect.ValueOf(fVal)
	err := floatField.UpdateValue(rVal)
	if err != nil {
		t.Errorf("SetValue failed, err:%s", err.Error())
		return
	}

	if !floatField.IsAssigned() {
		t.Error("SetValue failed")
		return
	}

	if floatField.GetValue().Get().Float() != float64(fVal) {
		t.Error("SetValue failed")
		return
	}

	ret = true
	return
}

func checkStringField(t *testing.T, strField model.Field) (ret bool) {
	fType := strField.GetType()

	ret = checkFieldType(t, fType, "string", nil)
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	if strField.IsAssigned() {
		t.Errorf("check is assigned failed")
		return
	}

	strVal := "12.345"
	rVal := reflect.ValueOf(strVal)
	err := strField.UpdateValue(rVal)
	if err != nil {
		t.Errorf("SetValue failed, err:%s", err.Error())
		return
	}

	if !strField.IsAssigned() {
		t.Error("SetValue failed")
		return
	}

	if strField.GetValue().Get().String() != strVal {
		t.Error("SetValue failed")
		return
	}

	ret = true
	return
}

func checkSliceField(t *testing.T, sliceField model.Field) (ret bool) {
	fType := sliceField.GetType()

	sliceVal := []string{"12.345"}
	ret = checkFieldType(t, fType, reflect.TypeOf(sliceVal).String(), nil)
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	if sliceField.IsAssigned() {
		t.Errorf("check is assigned failed")
		return
	}

	rVal := reflect.ValueOf(sliceVal)
	err := sliceField.UpdateValue(rVal)
	if err != nil {
		t.Errorf("SetValue failed, err:%s", err.Error())
		return
	}

	if !sliceField.IsAssigned() {
		t.Error("SetValue failed")
		return
	}

	strV := fmt.Sprintf("%v", sliceVal)
	strR := fmt.Sprintf("%v", sliceField.GetValue().Get().Interface())
	if strV != strR {
		t.Errorf("SetValue failed")
		return
	}

	ret = true
	return
}

func checkFieldType(t *testing.T, fType model.Type, typeName string, typeDepend model.Type) bool {
	if fType.GetName() != typeName {
		t.Errorf("get field type name failed, curType:%s, expect type:%s", fType.GetName(), typeName)
		return false
	}

	if fType.Depend() != typeDepend {
		dependType := "nil"
		if typeDepend != nil {
			dependType = typeDepend.GetName()
		}

		t.Errorf("check depend type failed, currentType:%s, dependType:%s", fType.Depend().GetName(), dependType)
		return false
	}

	if fType.Interface().Type().String() != typeName {
		t.Errorf("illegal interface, expect type:%s", typeName)
		return false
	}

	return true
}

func TestLocalProvider(t *testing.T) {
	provider := NewLocalProvider("default")

	baseEntity := Base{}

	provider.RegisterModel(baseEntity)
	defer provider.UnregisterModel(baseEntity)

	baseEntityModel, baseEntityErr := provider.GetEntityModel(&baseEntity)
	if baseEntityErr != nil {
		t.Errorf("get local entity model failed, err:%s", baseEntityErr.Error())
		return
	}

	checkModel(t, baseEntityModel)
}

func TestRemoteProvider(t *testing.T) {
	provider := NewRemoteProvider("default")

	baseEntity := Base{}

	baseObject, baseErr := remote.GetObject(baseEntity)
	if baseErr != nil {
		t.Errorf("GetObject failed, err:%s", baseErr.Error())
		return
	}

	baseVal, baseErr := remote.GetObjectValue(baseEntity)
	if baseErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", baseErr.Error())
		return
	}

	provider.RegisterModel(baseObject)
	defer provider.UnregisterModel(baseObject)

	baseEntityModel, baseEntityErr := provider.GetEntityModel(baseVal)
	if baseEntityErr != nil {
		t.Errorf("get local entity model failed, err:%s", baseEntityErr.Error())
		return
	}

	checkModel(t, baseEntityModel)
}
