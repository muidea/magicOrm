package provider

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/provider/helper"
)

// ComplexObj 测试结构体，包含嵌套类型和复杂字段
type ComplexObj struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	Name      string    `orm:"name" view:"detail,lite"`
	Value     float32   `orm:"value" view:"detail,lite"`
	TimeStamp time.Time `orm:"ts datetime" view:"detail,lite"`
	Items     []int     `orm:"items" view:"detail,lite"`
	Flag      bool      `orm:"flag" view:"detail,lite"`
	Namespace string    `orm:"namespace"`
}

// TestProviderReset 测试Provider的Reset方法
func TestProviderReset(t *testing.T) {
	// 测试LocalProvider的Reset
	localProvider := NewLocalProvider("test", nil)
	s := &Simple{}
	model1, err1 := localProvider.RegisterModel(s)
	if err1 != nil {
		t.Errorf("RegisterModel failed for local provider: %s", err1.Error())
		return
	}

	if model1 == nil {
		t.Errorf("Registered model should not be nil")
		return
	}

	// 检查模型是否已注册
	model2, err2 := localProvider.GetEntityModel(s, true)
	if err2 != nil {
		t.Errorf("GetEntityModel failed for local provider: %s", err2.Error())
		return
	}

	if model2 == nil {
		t.Errorf("Retrieved model should not be nil")
		return
	}

	// 重置Provider
	localProvider.Reset()

	// 重置后应该找不到已注册的模型
	_, err3 := localProvider.GetEntityModel(s, true)
	if err3 == nil {
		t.Errorf("After Reset, GetEntityModel should fail but didn't")
		return
	}

	// 测试RemoteProvider的Reset
	remoteProvider := NewRemoteProvider("test", nil)
	remoteComplexObj, remoteErr := helper.GetObject(&ComplexObj{})
	if remoteErr != nil {
		t.Errorf("Failed to get remote object: %s", remoteErr.Error())
		return
	}
	model4, err4 := remoteProvider.RegisterModel(remoteComplexObj)
	if err4 != nil {
		t.Errorf("RegisterModel failed for remote provider: %s", err4.Error())
		return
	}

	if model4 == nil {
		t.Errorf("Registered model should not be nil")
		return
	}

	// 重置Provider
	remoteProvider.Reset()

	// 重置后应该找不到已注册的模型
	_, err5 := remoteProvider.GetEntityModel(remoteComplexObj, true)
	if err5 == nil {
		t.Errorf("After Reset, GetEntityModel should fail but didn't")
		return
	}
}

// TestProviderOwner 测试Provider的Owner方法
func TestProviderOwner(t *testing.T) {
	owner := "test_owner_123"

	// 测试LocalProvider的Owner
	localProvider := NewLocalProvider(owner, nil)
	if localProvider.Owner() != owner {
		t.Errorf("LocalProvider.Owner() = %s, want %s", localProvider.Owner(), owner)
		return
	}

	// 测试RemoteProvider的Owner
	remoteProvider := NewRemoteProvider(owner, nil)
	if remoteProvider.Owner() != owner {
		t.Errorf("RemoteProvider.Owner() = %s, want %s", remoteProvider.Owner(), owner)
		return
	}
}

// TestUnregisterModel 测试UnregisterModel方法
func TestUnregisterModel(t *testing.T) {
	// 测试LocalProvider的UnregisterModel
	localProvider := NewLocalProvider("test", nil)
	s := &Simple{}

	// 注册模型
	_, err1 := localProvider.RegisterModel(s)
	if err1 != nil {
		t.Errorf("RegisterModel failed: %s", err1.Error())
		return
	}

	// 注销模型
	err2 := localProvider.UnregisterModel(s)
	if err2 != nil {
		t.Errorf("UnregisterModel failed: %s", err2.Error())
		return
	}

	// 尝试获取已注销的模型
	_, err3 := localProvider.GetEntityModel(s, true)
	if err3 == nil {
		t.Errorf("GetEntityModel after UnregisterModel should fail but didn't")
		return
	}

	// 测试RemoteProvider的UnregisterModel
	remoteProvider := NewRemoteProvider("test", nil)

	remoteComplexObj, remoteErr := helper.GetObject(&ComplexObj{})
	if remoteErr != nil {
		t.Errorf("Failed to get remote object: %s", remoteErr.Error())
		return
	}
	// 注册模型
	_, err4 := remoteProvider.RegisterModel(remoteComplexObj)
	if err4 != nil {
		t.Errorf("RegisterModel failed: %s", err4.Error())
		return
	}

	// 注销模型
	err5 := remoteProvider.UnregisterModel(remoteComplexObj)
	if err5 != nil {
		t.Errorf("UnregisterModel failed: %s", err5.Error())
		return
	}

	// 尝试获取已注销的模型
	_, err6 := remoteProvider.GetEntityModel(remoteComplexObj, true)
	if err6 == nil {
		t.Errorf("GetEntityModel after UnregisterModel should fail but didn't")
		return
	}
}

// TestGetTypeModel 测试GetTypeModel方法
func TestGetTypeModel(t *testing.T) {
	localProvider := NewLocalProvider("test", nil)
	s := &Simple{}

	// 注册模型
	_, err1 := localProvider.RegisterModel(s)
	if err1 != nil {
		t.Errorf("RegisterModel failed: %s", err1.Error())
		return
	}

	// 获取类型
	typeVal, typeErr := localProvider.GetEntityType(s)
	if typeErr != nil {
		t.Errorf("GetEntityType failed: %s", typeErr.Error())
		return
	}

	// 从类型获取模型
	modelVal, modelErr := localProvider.GetTypeModel(typeVal)
	if modelErr != nil {
		t.Errorf("GetTypeModel failed: %s", modelErr.Error())
		return
	}

	if modelVal == nil {
		t.Errorf("Model should not be nil")
		return
	}

	// 验证模型字段
	obj := modelVal.Interface(true).(*Simple)
	rt := reflect.TypeOf(obj)

	expectedFields := []string{"ID", "I8", "I16", "I32", "I64", "Name", "Value", "F64", "TimeStamp", "Flag", "Namespace"}
	for _, field := range expectedFields {
		if _, found := rt.Elem().FieldByName(field); !found {
			t.Errorf("Expected field %s not found in model", field)
			return
		}
	}
}

// TestGetValueModel 测试GetValueModel方法
func TestGetValueModel(t *testing.T) {
	localProvider := NewLocalProvider("test", nil)
	s := &Simple{
		ID:        123,
		Name:      "test_name",
		Value:     123.456,
		F64:       789.012,
		TimeStamp: time.Now(),
		Flag:      true,
		Namespace: "test_namespace",
	}

	// 注册模型
	_, err1 := localProvider.RegisterModel(s)
	if err1 != nil {
		t.Errorf("RegisterModel failed: %s", err1.Error())
		return
	}

	// 获取值
	_, valueErr := localProvider.GetEntityValue(s)
	if valueErr != nil {
		t.Errorf("GetEntityValue failed: %s", valueErr.Error())
		return
	}

	// 获取类型
	_, typeErr := localProvider.GetEntityType(s)
	if typeErr != nil {
		t.Errorf("GetEntityType failed: %s", typeErr.Error())
		return
	}
}

// TestErrorCases 测试异常情况处理
func TestErrorCases(t *testing.T) {
	localProvider := NewLocalProvider("test", nil)

	// 测试nil值的情况
	_, err1 := localProvider.RegisterModel(nil)
	if err1 == nil {
		t.Errorf("RegisterModel with nil should fail but didn't")
		return
	}

	_, err2 := localProvider.GetEntityType(nil)
	if err2 == nil {
		t.Errorf("GetEntityType with nil should fail but didn't")
		return
	}

	_, err3 := localProvider.GetEntityValue(nil)
	if err3 == nil {
		t.Errorf("GetEntityValue with nil should fail but didn't")
		return
	}

	_, err4 := localProvider.GetEntityModel(nil, true)
	if err4 == nil {
		t.Errorf("GetEntityModel with nil should fail but didn't")
		return
	}

	err5 := localProvider.UnregisterModel(nil)
	if err5 == nil {
		t.Errorf("UnregisterModel with nil should fail but didn't")
		return
	}
}
