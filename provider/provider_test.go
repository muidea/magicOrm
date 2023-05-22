package provider

import (
	"github.com/muidea/magicOrm/provider/local"
	"testing"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote"
)

func checkIntField(t *testing.T, intField model.Field) (ret bool) {
	fType := intField.GetType()

	ret = checkFieldType(t, fType, "int", "int")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkFloat32Field(t *testing.T, floatField model.Field) (ret bool) {
	fType := floatField.GetType()

	ret = checkFieldType(t, fType, "float32", "float32")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkStringField(t *testing.T, strField model.Field) (ret bool) {
	fType := strField.GetType()

	ret = checkFieldType(t, fType, "string", "string")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkSliceField(t *testing.T, sliceField model.Field) (ret bool) {
	fType := sliceField.GetType()

	ret = checkFieldType(t, fType, "string", "string")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkSliceStructField(t *testing.T, sliceField model.Field) (ret bool) {
	fType := sliceField.GetType()

	ret = checkFieldType(t, fType, "Base", "Base")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkStructField(t *testing.T, structField model.Field) (ret bool) {
	fType := structField.GetType()

	ret = checkFieldType(t, fType, "Base", "Base")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkStructPtrField(t *testing.T, structField model.Field) (ret bool) {
	fType := structField.GetType()

	ret = checkFieldType(t, fType, "Base", "Base")
	if !ret {
		t.Errorf("check field type failed")
		return
	}

	ret = true
	return
}

func checkFieldType(t *testing.T, fType model.Type, typeName, typeDepend string) bool {
	if fType.GetName() != typeName {
		t.Errorf("get field type name failed, curType:%s, expect type:%s", fType.GetName(), typeName)
		return false
	}

	if fType.Elem() != nil && typeDepend != "" {
		if fType.Elem().GetName() != typeDepend {
			t.Errorf("check depend type failed, currentType:%s, dependType:%s", fType.Elem().GetName(), typeDepend)
			return false
		}
	}
	if fType.Elem() == nil && typeDepend != "" {
		t.Errorf("check depend type failed, currentType:%s, dependType:%s", "nil", typeDepend)
		return false
	}
	if fType.Elem() != nil && typeDepend == "" {
		t.Errorf("check depend type failed, currentType:%s, dependType:%s", fType.Elem().GetName(), "nil")
		return false
	}

	return true
}

/*
	type Base struct {
		ID    int      `orm:"id key auto"`
		Str  string   `orm:"name"`
		Price float32  `orm:"price"`
		Addr  []string `orm:"addr"`
	}
*/
func checkBaseModel(t *testing.T, baseEntityModel model.Model) {
	modelName := baseEntityModel.GetName()
	if modelName != "Base" {
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

/*
	type Compose struct {
		ID    int     `orm:"id key auto"`
		Str  string  `orm:"name"`
		Info  Base    `orm:"info"`
		Ptr   *Base   `orm:"ptr"`
		Array []*Base `orm:"array"`
	}
*/
func checkExtModel(t *testing.T, extEntityModel model.Model) {
	modelName := extEntityModel.GetName()
	if modelName != "Compose" {
		t.Errorf("get model name failed, curName:%s", modelName)
		return
	}

	fields := extEntityModel.GetFields()
	if len(fields) != 5 {
		t.Errorf("get model fields failed")
		return
	}

	pkField := extEntityModel.GetPrimaryField()
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
			ret = checkStructField(t, val)
			if !ret {
				t.Errorf("checkStructField failed")
				return
			}
		}

		if val.GetIndex() == 3 {
			ret = checkStructPtrField(t, val)
			if !ret {
				t.Errorf("checkStructPtrField failed")
				return
			}
		}

		if val.GetIndex() == 4 {
			ret = checkSliceStructField(t, val)
			if !ret {
				t.Errorf("checkSliceStructField failed")
				return
			}
		}
	}

}

func TestLocalProvider(t *testing.T) {
	provider := NewLocalProvider("default", "abc")

	baseEntity := &Base{}
	extEntity := &Compose{}

	_, err := provider.RegisterModel(baseEntity)
	if err != nil {
		t.Errorf("registerModel failed, err:%s", err.Error())
		return
	}
	_, err = provider.RegisterModel(extEntity)
	if err != nil {
		t.Errorf("registerModel failed, err:%s", err.Error())
		return
	}

	defer provider.UnregisterModel(baseEntity)
	defer provider.UnregisterModel(extEntity)

	baseEntityModel, baseEntityErr := provider.GetEntityModel(baseEntity)
	if baseEntityErr != nil {
		t.Errorf("get local entity model failed, err:%s", baseEntityErr.Error())
		return
	}

	checkBaseModel(t, baseEntityModel)

	extEntityModel, extEntityErr := provider.GetEntityModel(extEntity)
	if extEntityErr != nil {
		t.Errorf("get local entity model failed, err:%s", extEntityErr.Error())
		return
	}

	checkExtModel(t, extEntityModel)
}

func TestRemoteProvider(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", "abc")

	baseEntity := &Base{}
	extEntity := &Compose{}

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

	extObject, extErr := remote.GetObject(extEntity)
	if extErr != nil {
		t.Errorf("GetObject failed, err:%s", extErr.Error())
		return
	}

	extVal, extErr := remote.GetObjectValue(extEntity)
	if extErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", extErr.Error())
		return
	}

	baseData, baseErr := remote.EncodeObject(baseObject)
	if baseErr != nil {
		t.Errorf("encode object faield, err:%s", baseErr.Error())
		return
	}
	baseObject, baseErr = remote.DecodeObject(baseData)
	if baseErr != nil {
		t.Errorf("decode object faield, err:%s", baseErr.Error())
		return
	}

	extData, extErr := remote.EncodeObject(extObject)
	if extErr != nil {
		t.Errorf("encode object faield, err:%s", extErr.Error())
		return
	}
	extObject, extErr = remote.DecodeObject(extData)
	if baseErr != nil {
		t.Errorf("decode object faield, err:%s", baseErr.Error())
		return
	}

	baseValData, baseValErr := remote.EncodeObjectValue(baseVal)
	if baseValErr != nil {
		t.Errorf("encode object faield, err:%s", baseValErr.Error())
		return
	}
	baseVal, baseValErr = remote.DecodeObjectValue(baseValData)
	if baseValErr != nil {
		t.Errorf("decode object faield, err:%s", baseValErr.Error())
		return
	}

	extValData, extValErr := remote.EncodeObjectValue(extVal)
	if extValErr != nil {
		t.Errorf("encode object faield, err:%s", extValErr.Error())
		return
	}
	extVal, extValErr = remote.DecodeObjectValue(extValData)
	if extValErr != nil {
		t.Errorf("decode object faield, err:%s", extValErr.Error())
		return
	}

	remoteProvider.RegisterModel(baseObject)
	remoteProvider.RegisterModel(extObject)
	defer remoteProvider.UnregisterModel(baseObject)
	defer remoteProvider.UnregisterModel(extObject)

	baseEntityModel, baseEntityErr := remoteProvider.GetEntityModel(baseVal)
	if baseEntityErr != nil {
		t.Errorf("get remote entity model failed, err:%s", baseEntityErr.Error())
		return
	}

	checkBaseModel(t, baseEntityModel)

	extEntityModel, extEntityErr := remoteProvider.GetEntityModel(extVal)
	if extEntityErr != nil {
		t.Errorf("get remote entity model failed, err:%s", extEntityErr.Error())
		return
	}

	checkExtModel(t, extEntityModel)
}

func TestUpdateRemoteProvider(t *testing.T) {
	provider := NewRemoteProvider("default", "abc")

	baseEntity := &Base{ID: 123, Str: "test int", F32: 12.35, StrArray: []string{"qq", "ar", "yt"}}
	extEntity := &Compose{ID: 234, Name: "hello", Base: *baseEntity, BasePtr: baseEntity, BasePtrArray: []*Base{baseEntity}}

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

	extObject, extErr := remote.GetObject(extEntity)
	if extErr != nil {
		t.Errorf("GetObject failed, err:%s", extErr.Error())
		return
	}

	extVal, extErr := remote.GetObjectValue(extEntity)
	if extErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", extErr.Error())
		return
	}

	baseData, baseErr := remote.EncodeObject(baseObject)
	if baseErr != nil {
		t.Errorf("encode object faield, err:%s", baseErr.Error())
		return
	}
	baseObject, baseErr = remote.DecodeObject(baseData)
	if baseErr != nil {
		t.Errorf("decode object faield, err:%s", baseErr.Error())
		return
	}

	extData, extErr := remote.EncodeObject(extObject)
	if extErr != nil {
		t.Errorf("encode object faield, err:%s", extErr.Error())
		return
	}
	extObject, extErr = remote.DecodeObject(extData)
	if baseErr != nil {
		t.Errorf("decode object faield, err:%s", baseErr.Error())
		return
	}

	baseValData, baseValErr := remote.EncodeObjectValue(baseVal)
	if baseValErr != nil {
		t.Errorf("encode object faield, err:%s", baseValErr.Error())
		return
	}
	baseVal, baseValErr = remote.DecodeObjectValue(baseValData)
	if baseValErr != nil {
		t.Errorf("decode object faield, err:%s", baseValErr.Error())
		return
	}

	extValData, extValErr := remote.EncodeObjectValue(extVal)
	if extValErr != nil {
		t.Errorf("encode object faield, err:%s", extValErr.Error())
		return
	}
	extVal, extValErr = remote.DecodeObjectValue(extValData)
	if extValErr != nil {
		t.Errorf("decode object faield, err:%s", extValErr.Error())
		return
	}

	provider.RegisterModel(baseObject)
	provider.RegisterModel(extObject)
	defer provider.UnregisterModel(baseObject)
	defer provider.UnregisterModel(extObject)

	base := &Base{}
	err := UpdateEntity(baseVal, base)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	if base.ID != baseEntity.ID || base.Str != baseEntity.Str || base.F32 != baseEntity.F32 || len(base.StrArray) != len(baseEntity.StrArray) {
		t.Error("UpdateEntity failed")
	}

	ext := &Compose{BasePtr: &Base{}}
	err = UpdateEntity(extVal, ext)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if ext.ID != extEntity.ID || ext.Name != extEntity.Name || ext.BasePtr.ID != extEntity.BasePtr.ID || ext.BasePtr.F32 != extEntity.BasePtr.F32 {
		t.Error("UpdateEntity failed")
		return
	}
	if ext.BasePtr == nil {
		t.Error("UpdateEntity failed")
		return
	}
	if ext.BasePtr.ID != extEntity.BasePtr.ID {
		t.Error("UpdateEntity failed")
		return
	}
	if len(ext.BasePtrArray) != len(extEntity.BasePtrArray) {
		t.Error("UpdateEntity failed")
		return
	}
}

func TestCompareProvider(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", "abc")
	localProvider := NewLocalProvider("default", "abc")

	baseEntity := &Base{ID: 123, Str: "test int", F32: 12.35, StrArray: []string{"qq", "ar", "yt"}}
	extEntity := &Compose{ID: 234, Name: "hello", Base: *baseEntity, BasePtr: baseEntity, BasePtrArray: []*Base{baseEntity}}

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

	extObject, extErr := remote.GetObject(extEntity)
	if extErr != nil {
		t.Errorf("GetObject failed, err:%s", extErr.Error())
		return
	}

	extVal, extErr := remote.GetObjectValue(extEntity)
	if extErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", extErr.Error())
		return
	}

	remoteProvider.RegisterModel(baseObject)
	remoteProvider.RegisterModel(extObject)
	defer remoteProvider.UnregisterModel(baseObject)
	defer remoteProvider.UnregisterModel(extObject)

	localProvider.RegisterModel(baseEntity)
	localProvider.RegisterModel(extEntity)
	defer localProvider.UnregisterModel(baseEntity)
	defer localProvider.UnregisterModel(extEntity)

	lBaseModel, lErr := localProvider.GetEntityModel(baseEntity)
	if lErr != nil {
		t.Errorf("GetEntityModel from localProvider failed, err:%s", lErr.Error())
		return
	}

	lBaseType, lErr := localProvider.GetEntityType(baseEntity)
	if lErr != nil {
		t.Errorf("GetEntityType from localProvider failed, err:%s", lErr.Error())
		return
	}

	lBaseVal, lErr := localProvider.GetEntityValue(baseEntity)
	if lErr != nil {
		t.Errorf("GetEntityValue from localProvider failed, err:%s", lErr.Error())
		return
	}

	l2BaseModel, l2Err := localProvider.GetTypeModel(lBaseType)
	if l2Err != nil {
		t.Errorf("GetTypeModel from localProvider failed, err:%s", l2Err.Error())
		return
	}
	l2BaseModel, l2Err = local.SetModelValue(l2BaseModel, lBaseVal)
	if l2Err != nil {
		t.Errorf("SetModelValue from localProvider failed, err:%s", l2Err.Error())
		return
	}
	if !model.CompareModel(lBaseModel, l2BaseModel) {
		t.Errorf("compareModel failed")
		return
	}

	rBaseModel, rErr := remoteProvider.GetEntityModel(baseVal)
	if rErr != nil {
		t.Errorf("GetEntityModel from remoteProvider failed, err:%s", rErr.Error())
		return
	}

	if !model.CompareModel(lBaseModel, rBaseModel) {
		t.Errorf("compareModel failed")
		return
	}

	lExtModel, lErr := localProvider.GetEntityModel(extEntity)
	if lErr != nil {
		t.Errorf("GetEntityModel from localProvider failed, err:%s", lErr.Error())
		return
	}

	rExtModel, rErr := remoteProvider.GetEntityModel(extVal)
	if rErr != nil {
		t.Errorf("GetEntityModel from remoteProvider failed, err:%s", rErr.Error())
		return
	}

	rExtType, rErr := remoteProvider.GetEntityType(extObject)
	if rErr != nil {
		t.Errorf("GetEntityType from remoteProvider failed, err:%s", rErr.Error())
		return
	}
	rExtVal, rErr := remoteProvider.GetEntityValue(extVal)
	if rErr != nil {
		t.Errorf("GetEntityValue from remoteProvider failed, err:%s", rErr.Error())
		return
	}

	r2ExtModel, r2Err := remoteProvider.GetTypeModel(rExtType)
	if r2Err != nil {
		t.Errorf("GetTypeModel from remoteProvider failed, err:%s", r2Err.Error())
		return
	}

	ext2Info := &Compose{BasePtr: &Base{}}
	r2Val := r2ExtModel.Interface(false).(remote.ObjectValue)
	r2Err = UpdateEntity(&r2Val, ext2Info)
	if r2Err != nil {
		t.Errorf("UpdateEntity from remoteProvider failed, err:%s", r2Err.Error())
		return
	}

	r2ExtModel, r2Err = remote.SetModelValue(r2ExtModel, rExtVal)
	if r2Err != nil {
		t.Errorf("SetModelValue from remoteProvider failed, err:%s", r2Err.Error())
		return
	}
	if !model.CompareModel(lExtModel, r2ExtModel) {
		t.Errorf("compareModel failed")
		return
	}

	r2ValPtr := r2ExtModel.Interface(true).(*remote.ObjectValue)
	r2Err = UpdateEntity(r2ValPtr, ext2Info)
	if r2Err != nil {
		t.Errorf("UpdateEntity from remoteProvider failed, err:%s", r2Err.Error())
		return
	}

	if !model.CompareModel(lExtModel, rExtModel) {
		t.Errorf("compareModel failed")
		return
	}
}
