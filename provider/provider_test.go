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

	ret = checkFieldType(t, fType, "int", "int")
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

func checkBaseModel(t *testing.T, baseEntityModel model.Model) {
	modelName := baseEntityModel.GetName()
	if modelName != "Base" {
		t.Errorf("get model name failed, curName:%s", modelName)
		return
	}

	fields := baseEntityModel.GetFields()
	if len(fields) != 25 {
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
		if val.GetIndex() == 13 {
			ret = checkStringField(t, val)
			if !ret {
				t.Errorf("checkStringField failed")
				return
			}
		}

		if val.GetIndex() == 11 {
			ret = checkFloat32Field(t, val)
			if !ret {
				t.Errorf("checkFloatField failed")
				return
			}
		}

		if val.GetIndex() == 16 {
			ret = checkSliceField(t, val)
			if !ret {
				t.Errorf("checkSliceField failed")
				return
			}
		}
	}
}

func checkComposeModel(t *testing.T, extEntityModel model.Model) {
	modelName := extEntityModel.GetName()
	if modelName != "Compose" {
		t.Errorf("get model name failed, curName:%s", modelName)
		return
	}

	fields := extEntityModel.GetFields()
	if len(fields) != 7 {
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
	localProvider := NewLocalProvider("default", "abc")

	baseEntity := emptyBase
	composeEntity := emptyCompose

	_, err := localProvider.RegisterModel(baseEntity)
	if err != nil {
		t.Errorf("registerModel failed, err:%s", err.Error())
		return
	}
	_, err = localProvider.RegisterModel(composeEntity)
	if err != nil {
		t.Errorf("registerModel failed, err:%s", err.Error())
		return
	}

	defer localProvider.UnregisterModel(baseEntity)
	defer localProvider.UnregisterModel(composeEntity)

	baseEntityModel, baseEntityErr := localProvider.GetEntityModel(baseEntity)
	if baseEntityErr != nil {
		t.Errorf("get local entity model failed, err:%s", baseEntityErr.Error())
		return
	}

	checkBaseModel(t, baseEntityModel)

	composeEntityModel, composeEntityErr := localProvider.GetEntityModel(composeEntity)
	if composeEntityErr != nil {
		t.Errorf("get local entity model failed, err:%s", composeEntityErr.Error())
		return
	}

	checkComposeModel(t, composeEntityModel)
}

func TestRemoteProvider(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", "abc")

	baseEntity := baseVal
	composeEntity := composeVal

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

	composeObject, composeErr := remote.GetObject(composeEntity)
	if composeErr != nil {
		t.Errorf("GetObject failed, err:%s", composeErr.Error())
		return
	}

	composeVal, composeErr := remote.GetObjectValue(composeEntity)
	if composeErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", composeErr.Error())
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

	composeData, composeErr := remote.EncodeObject(composeObject)
	if composeErr != nil {
		t.Errorf("encode object faield, err:%s", composeErr.Error())
		return
	}
	composeObject, composeErr = remote.DecodeObject(composeData)
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

	composeValData, composeValErr := remote.EncodeObjectValue(composeVal)
	if composeValErr != nil {
		t.Errorf("encode object faield, err:%s", composeValErr.Error())
		return
	}
	composeVal, composeValErr = remote.DecodeObjectValue(composeValData)
	if composeValErr != nil {
		t.Errorf("decode object faield, err:%s", composeValErr.Error())
		return
	}

	remoteProvider.RegisterModel(baseObject)
	remoteProvider.RegisterModel(composeObject)
	defer remoteProvider.UnregisterModel(baseObject)
	defer remoteProvider.UnregisterModel(composeObject)

	baseEntityModel, baseEntityErr := remoteProvider.GetEntityModel(baseVal)
	if baseEntityErr != nil {
		t.Errorf("get remote entity model failed, err:%s", baseEntityErr.Error())
		return
	}

	checkBaseModel(t, baseEntityModel)

	composeEntityModel, composeEntityErr := remoteProvider.GetEntityModel(composeVal)
	if composeEntityErr != nil {
		t.Errorf("get remote entity model failed, err:%s", composeEntityErr.Error())
		return
	}

	checkComposeModel(t, composeEntityModel)
}

func TestUpdateRemoteProvider(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", "abc")

	baseEntity := baseVal
	composeEntity := composeVal

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

	composeObject, composeErr := remote.GetObject(composeEntity)
	if composeErr != nil {
		t.Errorf("GetObject failed, err:%s", composeErr.Error())
		return
	}

	composeVal, composeErr := remote.GetObjectValue(composeEntity)
	if composeErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", composeErr.Error())
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

	composeData, composeErr := remote.EncodeObject(composeObject)
	if composeErr != nil {
		t.Errorf("encode object faield, err:%s", composeErr.Error())
		return
	}
	composeObject, composeErr = remote.DecodeObject(composeData)
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

	composeValData, composeValErr := remote.EncodeObjectValue(composeVal)
	if composeValErr != nil {
		t.Errorf("encode object faield, err:%s", composeValErr.Error())
		return
	}
	composeVal, composeValErr = remote.DecodeObjectValue(composeValData)
	if composeValErr != nil {
		t.Errorf("decode object faield, err:%s", composeValErr.Error())
		return
	}

	remoteProvider.RegisterModel(baseObject)
	remoteProvider.RegisterModel(composeObject)
	defer remoteProvider.UnregisterModel(baseObject)
	defer remoteProvider.UnregisterModel(composeObject)

	base := &Base{}
	err := UpdateLocalEntity(baseVal, base)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	if base.ID != baseEntity.ID || base.Str != baseEntity.Str || base.F32 != baseEntity.F32 || len(base.StrArray) != len(baseEntity.StrArray) {
		t.Error("UpdateLocalEntity failed")
	}

	compose := &Compose{BasePtr: &Base{}}
	err = UpdateLocalEntity(composeVal, compose)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if compose.ID != composeEntity.ID || compose.Name != composeEntity.Name || compose.BasePtr.ID != composeEntity.BasePtr.ID || compose.BasePtr.F32 != composeEntity.BasePtr.F32 {
		t.Error("UpdateLocalEntity failed")
		return
	}
	if compose.BasePtr == nil {
		t.Error("UpdateLocalEntity failed")
		return
	}
	if compose.BasePtr.ID != composeEntity.BasePtr.ID {
		t.Error("UpdateLocalEntity failed")
		return
	}
	if len(compose.BasePtrArray) != len(composeEntity.BasePtrArray) {
		t.Error("UpdateLocalEntity failed")
		return
	}
}

func TestCompareProvider(t *testing.T) {
	remoteProvider := NewRemoteProvider("default", "abc")
	localProvider := NewLocalProvider("default", "abc")

	baseEntity := baseVal
	composeEntity := composeVal

	baseObject, baseErr := remote.GetObject(baseEntity)
	if baseErr != nil {
		t.Errorf("GetObject failed, err:%s", baseErr.Error())
		return
	}

	baseObjectVal, baseErr := remote.GetObjectValue(baseEntity)
	if baseErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", baseErr.Error())
		return
	}

	composeObject, composeErr := remote.GetObject(composeEntity)
	if composeErr != nil {
		t.Errorf("GetObject failed, err:%s", composeErr.Error())
		return
	}

	composeObjectVal, composeErr := remote.GetObjectValue(composeEntity)
	if composeErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", composeErr.Error())
		return
	}

	remoteProvider.RegisterModel(baseObject)
	remoteProvider.RegisterModel(composeObject)
	defer remoteProvider.UnregisterModel(baseObject)
	defer remoteProvider.UnregisterModel(composeObject)

	localProvider.RegisterModel(baseEntity)
	localProvider.RegisterModel(composeEntity)
	defer localProvider.UnregisterModel(baseEntity)
	defer localProvider.UnregisterModel(composeEntity)

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

	rBaseModel, rErr := remoteProvider.GetEntityModel(baseObjectVal)
	if rErr != nil {
		t.Errorf("GetEntityModel from remoteProvider failed, err:%s", rErr.Error())
		return
	}

	if !model.CompareModel(lBaseModel, rBaseModel) {
		t.Errorf("compareModel failed")
		return
	}

	lComposeModel, lErr := localProvider.GetEntityModel(composeEntity)
	if lErr != nil {
		t.Errorf("GetEntityModel from localProvider failed, err:%s", lErr.Error())
		return
	}

	rComposeModel, rErr := remoteProvider.GetEntityModel(composeObjectVal)
	if rErr != nil {
		t.Errorf("GetEntityModel from remoteProvider failed, err:%s", rErr.Error())
		return
	}

	rComposeType, rErr := remoteProvider.GetEntityType(composeObject)
	if rErr != nil {
		t.Errorf("GetEntityType from remoteProvider failed, err:%s", rErr.Error())
		return
	}
	rComposeVal, rErr := remoteProvider.GetEntityValue(composeObjectVal)
	if rErr != nil {
		t.Errorf("GetEntityValue from remoteProvider failed, err:%s", rErr.Error())
		return
	}

	r2ComposeModel, r2Err := remoteProvider.GetTypeModel(rComposeType)
	if r2Err != nil {
		t.Errorf("GetTypeModel from remoteProvider failed, err:%s", r2Err.Error())
		return
	}

	r2ComposeModel, r2Err = remote.SetModelValue(r2ComposeModel, rComposeVal)
	if r2Err != nil {
		t.Errorf("SetModelValue from remoteProvider failed, err:%s", r2Err.Error())
		return
	}

	if !model.CompareModel(lComposeModel, r2ComposeModel) {
		t.Errorf("compareModel failed")
		return
	}

	compose2Info := &Compose{BasePtr: &Base{}}
	r2ValPtr := r2ComposeModel.Interface(true).(*remote.ObjectValue)
	r2Err = UpdateLocalEntity(r2ValPtr, compose2Info)
	if r2Err != nil {
		t.Errorf("UpdateLocalEntity from remoteProvider failed, err:%s", r2Err.Error())
		return
	}

	if !model.CompareModel(lComposeModel, rComposeModel) {
		t.Errorf("compareModel failed")
		return
	}
}
