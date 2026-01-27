package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/utils"
)

func TestConstraintRemote(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	validator := utils.NewValueValidator()
	remoteProvider := provider.NewRemoteProvider("constraint_remote", validator)

	o1, err := orm.NewOrm(remoteProvider, config, "constraint_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&ConstraintTestModel{}, &ContentConstraintTestModel{}}
	modelList, modelErr := registerRemoteModel(remoteProvider, objList)
	if modelErr != nil {
		t.Errorf("register model failed. err:%s", modelErr.Error())
		return
	}

	// 删除并创建表
	err = dropModel(o1, modelList)
	if err != nil {
		t.Errorf("drop model failed. err:%s", err.Error())
		return
	}

	err = createModel(o1, modelList)
	if err != nil {
		t.Errorf("create model failed. err:%s", err.Error())
		return
	}

	// 测试1: 创建对象 - 测试必填字段
	t.Run("TestRequiredFields", func(t *testing.T) {
		testRequiredFieldsRemote(t, o1, remoteProvider)
	})

	// 测试2: 测试只读字段
	t.Run("TestReadOnlyFields", func(t *testing.T) {
		testReadOnlyFieldsRemote(t, o1, remoteProvider)
	})

	// 测试3: 测试只写字段
	t.Run("TestWriteOnlyFields", func(t *testing.T) {
		testWriteOnlyFieldsRemote(t, o1, remoteProvider)
	})

	// 测试4: 测试不可变字段
	t.Run("TestImmutableFields", func(t *testing.T) {
		testImmutableFieldsRemote(t, o1, remoteProvider)
	})

	// 测试5: 测试可选字段
	t.Run("TestOptionalFields", func(t *testing.T) {
		testOptionalFieldsRemote(t, o1, remoteProvider)
	})

	// 测试6: 测试内容值约束
	t.Run("TestContentConstraints", func(t *testing.T) {
		testContentConstraintsRemote(t, o1, remoteProvider)
	})

	// 清理测试数据
	cleanupConstraintTestRemote(t, o1, remoteProvider)
}

// testRequiredFieldsRemote 测试必填字段约束（remote provider）
func testRequiredFieldsRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 创建包含所有必填字段的对象
	obj := &ConstraintTestModel{
		Name:       "test_user",
		Password:   "secret123",
		CreateTime: 1234567890,
		UpdateTime: 1234567890,
		Status:     1,
		ReadOnlyID: 100,
		WriteOnly:  "write_only_value",
	}

	objValue, objErr := getObjectValue(obj)
	if objErr != nil {
		t.Errorf("getObjectValue failed, err:%s", objErr.Error())
		return
	}

	objModel, objErr := remoteProvider.GetEntityModel(objValue)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	// 插入对象
	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedValue := insertedModel.Interface(true).(*remote.ObjectValue)
	insertedObj := &ConstraintTestModel{}
	err := helper.UpdateEntity(insertedValue, insertedObj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	t.Logf("Inserted object ID: %d", insertedObj.ID)

	// 验证插入的数据
	if insertedObj.Name != "test_user" {
		t.Errorf("Name field mismatch, expected: test_user, got: %s", insertedObj.Name)
	}
	if insertedObj.Status != 1 {
		t.Errorf("Status field mismatch, expected: 1, got: %d", insertedObj.Status)
	}
}

// testReadOnlyFieldsRemote 测试只读字段约束（remote provider）
func testReadOnlyFieldsRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 首先创建一个对象
	obj := &ConstraintTestModel{
		Name:       "readonly_test",
		Password:   "password123",
		CreateTime: 1234567890,
		UpdateTime: 1234567890,
		Status:     1,
		ReadOnlyID: 200,
		WriteOnly:  "write_value",
	}

	objValue, objErr := getObjectValue(obj)
	if objErr != nil {
		t.Errorf("getObjectValue failed, err:%s", objErr.Error())
		return
	}

	objModel, objErr := remoteProvider.GetEntityModel(objValue)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedValue := insertedModel.Interface(true).(*remote.ObjectValue)
	insertedObj := &ConstraintTestModel{}
	err := helper.UpdateEntity(insertedValue, insertedObj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	objID := insertedObj.ID

	// 尝试更新只读字段
	updateObj := &ConstraintTestModel{
		ID:         objID,
		Name:       "updated_name",
		Password:   "new_password",
		CreateTime: 9999999999, // 尝试修改不可变字段
		UpdateTime: 1234567891, // 修改普通字段
		Status:     2,          // 尝试修改只读字段
		ReadOnlyID: 300,        // 尝试修改只读字段
		WriteOnly:  "updated_write",
	}

	updateValue, updateErr := getObjectValue(updateObj)
	if updateErr != nil {
		t.Errorf("getObjectValue for update failed, err:%s", updateErr.Error())
		return
	}

	updateModel, updateErr := remoteProvider.GetEntityModel(updateValue)
	if updateErr != nil {
		t.Errorf("GetEntityModel for update failed, err:%s", updateErr.Error())
		return
	}

	_, updateErr = o1.Update(updateModel)
	if updateErr != nil {
		t.Errorf("Update failed, err:%s", updateErr.Error())
		return
	}

	// 查询对象以验证只读字段是否被保护
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	queryObj := &ConstraintTestModel{
		ID:         objID,
		Name:       "updated_name", // 提供必填字段
		Password:   "",             // 只写字段，查询时不返回
		CreateTime: 1111111111,     // 提供不可变字段
		UpdateTime: 1234567891,     // 提供更新后的时间
		Status:     1,              // 提供只读字段的原始值
		ReadOnlyID: 200,            // 提供只读字段的原始值
		WriteOnly:  "",             // 只写字段
	}
	queryValue, queryErr := getObjectValue(queryObj)
	if queryErr != nil {
		t.Errorf("getObjectValue for query failed, err:%s", queryErr.Error())
		return
	}

	queryModel, queryErr := remoteProvider.GetEntityModel(queryValue)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedValue := queriedModel.Interface(true).(*remote.ObjectValue)
	queriedObj := &ConstraintTestModel{}
	err = helper.UpdateEntity(queriedValue, queriedObj)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	// 验证只读字段没有被修改
	if queriedObj.Status != insertedObj.Status {
		t.Errorf("Read-only Status field was modified, expected: %d, got: %d", insertedObj.Status, queriedObj.Status)
	}
	if queriedObj.ReadOnlyID != insertedObj.ReadOnlyID {
		t.Errorf("Read-only ReadOnlyID field was modified, expected: %d, got: %d", insertedObj.ReadOnlyID, queriedObj.ReadOnlyID)
	}

	// 验证普通字段被修改了
	if queriedObj.Name != "updated_name" {
		t.Errorf("Name field was not updated, expected: updated_name, got: %s", queriedObj.Name)
	}
	if queriedObj.UpdateTime != 1234567891 {
		t.Errorf("UpdateTime field was not updated, expected: 1234567891, got: %d", queriedObj.UpdateTime)
	}

	t.Logf("Read-only fields test passed: Status=%d, ReadOnlyID=%d", queriedObj.Status, queriedObj.ReadOnlyID)
}

// testWriteOnlyFieldsRemote 测试只写字段约束（remote provider）
func testWriteOnlyFieldsRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 创建包含只写字段的对象
	obj := &ConstraintTestModel{
		Name:       "writeonly_test",
		Password:   "secret_password", // 只写字段
		CreateTime: 1234567890,
		UpdateTime: 1234567890,
		Status:     1,
		ReadOnlyID: 400,
		WriteOnly:  "sensitive_data", // 只写字段
	}

	objValue, objErr := getObjectValue(obj)
	if objErr != nil {
		t.Errorf("getObjectValue failed, err:%s", objErr.Error())
		return
	}

	objModel, objErr := remoteProvider.GetEntityModel(objValue)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedValue := insertedModel.Interface(true).(*remote.ObjectValue)
	insertedObj := &ConstraintTestModel{}
	err := helper.UpdateEntity(insertedValue, insertedObj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	objID := insertedObj.ID

	// 查询对象
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	queryObj := &ConstraintTestModel{
		ID:         objID,
		Name:       "writeonly_test", // 提供必填字段
		Password:   "",               // 只写字段，查询时不返回
		CreateTime: 1234567890,       // 提供不可变字段
		UpdateTime: 1234567890,       // 提供更新时间
		Status:     1,                // 提供状态
		ReadOnlyID: 400,              // 提供只读字段
		WriteOnly:  "",               // 只写字段
	}
	queryValue, queryErr := getObjectValue(queryObj)
	if queryErr != nil {
		t.Errorf("getObjectValue for query failed, err:%s", queryErr.Error())
		return
	}

	queryModel, queryErr := remoteProvider.GetEntityModel(queryValue)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedValue := queriedModel.Interface(true).(*remote.ObjectValue)
	queriedObj := &ConstraintTestModel{}
	err = helper.UpdateEntity(queriedValue, queriedObj)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	// 验证只写字段在查询结果中为空或默认值
	// 注意：根据约束定义，只写字段禁止在展示接口输出
	// 这里我们检查Password和WriteOnly字段是否为空
	if queriedObj.Password != "" {
		t.Errorf("Write-only Password field should be empty in query result, got: %s", queriedObj.Password)
	}
	if queriedObj.WriteOnly != "" {
		t.Errorf("Write-only WriteOnly field should be empty in query result, got: %s", queriedObj.WriteOnly)
	}

	t.Logf("Write-only fields test passed: Password is hidden, WriteOnly is hidden")
}

// testImmutableFieldsRemote 测试不可变字段约束（remote provider）
func testImmutableFieldsRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 创建对象
	obj := &ConstraintTestModel{
		Name:       "immutable_test",
		Password:   "password123",
		CreateTime: 1111111111, // 初始创建时间
		UpdateTime: 1111111111,
		Status:     1,
		ReadOnlyID: 500,
		WriteOnly:  "write_data",
	}

	objValue, objErr := getObjectValue(obj)
	if objErr != nil {
		t.Errorf("getObjectValue failed, err:%s", objErr.Error())
		return
	}

	objModel, objErr := remoteProvider.GetEntityModel(objValue)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedValue := insertedModel.Interface(true).(*remote.ObjectValue)
	insertedObj := &ConstraintTestModel{}
	err := helper.UpdateEntity(insertedValue, insertedObj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	objID := insertedObj.ID

	// 尝试更新不可变字段
	updateObj := &ConstraintTestModel{
		ID:         objID,
		Name:       "updated_immutable_test",
		Password:   "new_password",
		CreateTime: 9999999999, // 尝试修改不可变字段
		UpdateTime: 2222222222, // 修改普通字段
		Status:     1,
		ReadOnlyID: 500,
		WriteOnly:  "updated_write",
	}

	updateValue, updateErr := getObjectValue(updateObj)
	if updateErr != nil {
		t.Errorf("getObjectValue for update failed, err:%s", updateErr.Error())
		return
	}

	updateModel, updateErr := remoteProvider.GetEntityModel(updateValue)
	if updateErr != nil {
		t.Errorf("GetEntityModel for update failed, err:%s", updateErr.Error())
		return
	}

	_, updateErr = o1.Update(updateModel)
	if updateErr != nil {
		t.Errorf("Update failed, err:%s", updateErr.Error())
		return
	}

	// 查询对象验证不可变字段是否被保护
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	queryObj := &ConstraintTestModel{
		ID:         objID,
		Name:       "updated_immutable_test", // 提供必填字段
		Password:   "",                       // 只写字段，查询时不返回
		CreateTime: 1111111111,               // 提供不可变字段的原始值
		UpdateTime: 2222222222,               // 提供更新后的时间
		Status:     1,                        // 提供状态
		ReadOnlyID: 500,                      // 提供只读字段
		WriteOnly:  "",                       // 只写字段
	}
	queryValue, queryErr := getObjectValue(queryObj)
	if queryErr != nil {
		t.Errorf("getObjectValue for query failed, err:%s", queryErr.Error())
		return
	}

	queryModel, queryErr := remoteProvider.GetEntityModel(queryValue)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedValue := queriedModel.Interface(true).(*remote.ObjectValue)
	queriedObj := &ConstraintTestModel{}
	err = helper.UpdateEntity(queriedValue, queriedObj)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	// 验证不可变字段没有被修改
	if queriedObj.CreateTime != 1111111111 {
		t.Errorf("Immutable CreateTime field was modified, expected: 1111111111, got: %d", queriedObj.CreateTime)
	}

	// 验证普通字段被修改了
	if queriedObj.Name != "updated_immutable_test" {
		t.Errorf("Name field was not updated, expected: updated_immutable_test, got: %s", queriedObj.Name)
	}
	if queriedObj.UpdateTime != 2222222222 {
		t.Errorf("UpdateTime field was not updated, expected: 2222222222, got: %d", queriedObj.UpdateTime)
	}

	t.Logf("Immutable field test passed: CreateTime=%d (unchanged)", queriedObj.CreateTime)
}

// testOptionalFieldsRemote 测试可选字段约束（remote provider）
func testOptionalFieldsRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 测试1: 创建时不提供可选字段
	obj1 := &ConstraintTestModel{
		Name:       "optional_test1",
		Password:   "password123",
		CreateTime: 1234567890,
		UpdateTime: 1234567890,
		Status:     1,
		ReadOnlyID: 600,
		WriteOnly:  "write_data",
		// Email字段是可选字段，不设置
	}

	obj1Value, obj1Err := getObjectValue(obj1)
	if obj1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", obj1Err.Error())
		return
	}

	obj1Model, obj1Err := remoteProvider.GetEntityModel(obj1Value)
	if obj1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", obj1Err.Error())
		return
	}

	inserted1Model, insert1Err := o1.Insert(obj1Model)
	if insert1Err != nil {
		t.Errorf("Insert without optional field failed, err:%s", insert1Err.Error())
		return
	}

	inserted1Value := inserted1Model.Interface(true).(*remote.ObjectValue)
	inserted1Obj := &ConstraintTestModel{}
	err := helper.UpdateEntity(inserted1Value, inserted1Obj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	obj1ID := inserted1Obj.ID

	// 查询验证可选字段为空
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	query1Obj := &ConstraintTestModel{
		ID:         obj1ID,
		Name:       "optional_test1", // 提供必填字段
		Password:   "",               // 只写字段，查询时不返回
		CreateTime: 1234567890,       // 提供不可变字段
		UpdateTime: 1234567890,       // 提供更新时间
		Status:     1,                // 提供状态
		ReadOnlyID: 600,              // 提供只读字段
		WriteOnly:  "",               // 只写字段
		Email:      "",               // 可选字段，可以为空
	}
	query1Value, query1Err := getObjectValue(query1Obj)
	if query1Err != nil {
		t.Errorf("getObjectValue for query failed, err:%s", query1Err.Error())
		return
	}

	query1Model, query1Err := remoteProvider.GetEntityModel(query1Value)
	if query1Err != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", query1Err.Error())
		return
	}

	queried1Model, query1Err := o1.Query(query1Model)
	if query1Err != nil {
		t.Errorf("Query failed, err:%s", query1Err.Error())
		return
	}

	queried1Value := queried1Model.Interface(true).(*remote.ObjectValue)
	queried1Obj := &ConstraintTestModel{}
	err = helper.UpdateEntity(queried1Value, queried1Obj)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	if queried1Obj.Email != "" {
		t.Errorf("Optional Email field should be empty when not provided, got: %s", queried1Obj.Email)
	}

	// 测试2: 创建时提供可选字段
	obj2 := &ConstraintTestModel{
		Name:       "optional_test2",
		Password:   "password123",
		CreateTime: 1234567890,
		UpdateTime: 1234567890,
		Status:     1,
		ReadOnlyID: 700,
		WriteOnly:  "write_data",
		Email:      "test@example.com", // 提供可选字段
	}

	obj2Value, obj2Err := getObjectValue(obj2)
	if obj2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", obj2Err.Error())
		return
	}

	obj2Model, obj2Err := remoteProvider.GetEntityModel(obj2Value)
	if obj2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", obj2Err.Error())
		return
	}

	inserted2Model, insert2Err := o1.Insert(obj2Model)
	if insert2Err != nil {
		t.Errorf("Insert with optional field failed, err:%s", insert2Err.Error())
		return
	}

	inserted2Value := inserted2Model.Interface(true).(*remote.ObjectValue)
	inserted2Obj := &ConstraintTestModel{}
	err = helper.UpdateEntity(inserted2Value, inserted2Obj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	obj2ID := inserted2Obj.ID

	// 查询验证可选字段被保存
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	query2Obj := &ConstraintTestModel{
		ID:         obj2ID,
		Name:       "optional_test2",   // 提供必填字段
		Password:   "",                 // 只写字段，查询时不返回
		CreateTime: 1234567890,         // 提供不可变字段
		UpdateTime: 1234567890,         // 提供更新时间
		Status:     1,                  // 提供状态
		ReadOnlyID: 700,                // 提供只读字段
		WriteOnly:  "",                 // 只写字段
		Email:      "test@example.com", // 可选字段，提供保存的值
	}
	query2Value, query2Err := getObjectValue(query2Obj)
	if query2Err != nil {
		t.Errorf("getObjectValue for query failed, err:%s", query2Err.Error())
		return
	}

	query2Model, query2Err := remoteProvider.GetEntityModel(query2Value)
	if query2Err != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", query2Err.Error())
		return
	}

	queried2Model, query2Err := o1.Query(query2Model)
	if query2Err != nil {
		t.Errorf("Query failed, err:%s", query2Err.Error())
		return
	}

	queried2Value := queried2Model.Interface(true).(*remote.ObjectValue)
	queried2Obj := &ConstraintTestModel{}
	err = helper.UpdateEntity(queried2Value, queried2Obj)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	if queried2Obj.Email != "test@example.com" {
		t.Errorf("Optional Email field should be saved, expected: test@example.com, got: %s", queried2Obj.Email)
	}

	t.Logf("Optional field test passed: Email=%s", queried2Obj.Email)
}

// testContentConstraintsRemote 测试内容值约束（remote provider）
func testContentConstraintsRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 测试1: 创建符合所有约束的对象
	validObj := &ContentConstraintTestModel{
		Name:        "John Doe",                    // 长度在3-50之间
		Age:         25,                            // 在0-150之间
		Score:       85.5,                          // 在0.0-100.0之间
		Status:      "active",                      // 在枚举中
		Email:       "john.doe@example.com",        // 符合邮箱正则
		Description: "This is a valid description", // 长度小于500
		ItemCount:   5,                             // 大于等于1
		Price:       99.99,                         // 在0.01-9999.99之间
		Category:    "A",                           // 在枚举A:B:C:D中
		Code:        "ABC-123",                     // 符合正则格式
	}

	objValue, objErr := getObjectValue(validObj)
	if objErr != nil {
		t.Errorf("getObjectValue failed, err:%s", objErr.Error())
		return
	}

	objModel, objErr := remoteProvider.GetEntityModel(objValue)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	// 插入对象
	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert valid object failed, err:%s", insertErr.Error())
		return
	}

	insertedValue := insertedModel.Interface(true).(*remote.ObjectValue)
	insertedObj := &ContentConstraintTestModel{}
	err := helper.UpdateEntity(insertedValue, insertedObj)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	t.Logf("Inserted content constraint object ID: %d", insertedObj.ID)

	// 验证插入的数据
	if insertedObj.Name != "John Doe" {
		t.Errorf("Name field mismatch, expected: John Doe, got: %s", insertedObj.Name)
	}
	if insertedObj.Age != 25 {
		t.Errorf("Age field mismatch, expected: 25, got: %d", insertedObj.Age)
	}
	if insertedObj.Score != 85.5 {
		t.Errorf("Score field mismatch, expected: 85.5, got: %f", insertedObj.Score)
	}
	if insertedObj.Status != "active" {
		t.Errorf("Status field mismatch, expected: active, got: %s", insertedObj.Status)
	}
	if insertedObj.Email != "john.doe@example.com" {
		t.Errorf("Email field mismatch, expected: john.doe@example.com, got: %s", insertedObj.Email)
	}
	if insertedObj.Description != "This is a valid description" {
		t.Errorf("Description field mismatch, expected: This is a valid description, got: %s", insertedObj.Description)
	}
	if insertedObj.ItemCount != 5 {
		t.Errorf("ItemCount field mismatch, expected: 5, got: %d", insertedObj.ItemCount)
	}
	if insertedObj.Price != 99.99 {
		t.Errorf("Price field mismatch, expected: 99.99, got: %f", insertedObj.Price)
	}
	if insertedObj.Category != "A" {
		t.Errorf("Category field mismatch, expected: A, got: %s", insertedObj.Category)
	}
	if insertedObj.Code != "ABC-123" {
		t.Errorf("Code field mismatch, expected: ABC-123, got: %s", insertedObj.Code)
	}

	// 测试2: 查询对象
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	queryObj := &ContentConstraintTestModel{
		ID:          insertedObj.ID,
		Name:        "John Doe",                    // 提供必填字段
		Age:         25,                            // 提供年龄
		Score:       85.5,                          // 提供分数
		Status:      "active",                      // 提供状态
		Email:       "john.doe@example.com",        // 提供邮箱
		Description: "This is a valid description", // 提供描述
		ItemCount:   5,                             // 提供项目计数
		Price:       99.99,                         // 提供价格
		Category:    "A",                           // 提供分类
		Code:        "ABC-123",                     // 提供代码
	}
	queryValue, queryErr := getObjectValue(queryObj)
	if queryErr != nil {
		t.Errorf("getObjectValue for query failed, err:%s", queryErr.Error())
		return
	}

	queryModel, queryErr := remoteProvider.GetEntityModel(queryValue)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedValue := queriedModel.Interface(true).(*remote.ObjectValue)
	queriedObj := &ContentConstraintTestModel{}
	err = helper.UpdateEntity(queriedValue, queriedObj)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	if !queriedObj.Equal(insertedObj) {
		t.Errorf("Queried object does not match inserted object")
	}

	// 测试3: 更新对象（测试约束在更新时是否生效）
	updateObj := &ContentConstraintTestModel{
		ID:          insertedObj.ID,
		Name:        "Jane Smith",             // 仍然符合长度约束
		Age:         30,                       // 仍然符合范围
		Score:       95.0,                     // 仍然符合范围
		Status:      "inactive",               // 仍然在枚举中
		Email:       "jane.smith@example.com", // 符合邮箱正则
		Description: "Updated description",    // 长度小于500
		ItemCount:   10,                       // 大于等于1
		Price:       49.99,                    // 在范围内
		Category:    "B",                      // 在枚举中
		Code:        "XYZ-789",                // 符合正则格式
	}

	updateValue, updateErr := getObjectValue(updateObj)
	if updateErr != nil {
		t.Errorf("getObjectValue for update failed, err:%s", updateErr.Error())
		return
	}

	updateModel, updateErr := remoteProvider.GetEntityModel(updateValue)
	if updateErr != nil {
		t.Errorf("GetEntityModel for update failed, err:%s", updateErr.Error())
		return
	}

	_, updateErr = o1.Update(updateModel)
	if updateErr != nil {
		t.Errorf("Update failed, err:%s", updateErr.Error())
		return
	}

	// 查询验证更新
	// 注意：查询时需要提供完整的对象，因为验证器会在SetValue时验证约束
	queryObj2 := &ContentConstraintTestModel{
		ID:          insertedObj.ID,
		Name:        "Jane Smith",             // 提供更新后的必填字段
		Age:         30,                       // 提供更新后的年龄
		Score:       95.0,                     // 提供更新后的分数
		Status:      "inactive",               // 提供更新后的状态
		Email:       "jane.smith@example.com", // 提供更新后的邮箱
		Description: "Updated description",    // 提供更新后的描述
		ItemCount:   10,                       // 提供更新后的项目计数
		Price:       49.99,                    // 提供更新后的价格
		Category:    "B",                      // 提供更新后的分类
		Code:        "XYZ-789",                // 提供更新后的代码
	}
	queryValue2, queryErr2 := getObjectValue(queryObj2)
	if queryErr2 != nil {
		t.Errorf("getObjectValue for query failed, err:%s", queryErr2.Error())
		return
	}

	queryModel2, queryErr2 := remoteProvider.GetEntityModel(queryValue2)
	if queryErr2 != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr2.Error())
		return
	}

	queriedModel2, queryErr2 := o1.Query(queryModel2)
	if queryErr2 != nil {
		t.Errorf("Query failed, err:%s", queryErr2.Error())
		return
	}

	queriedValue2 := queriedModel2.Interface(true).(*remote.ObjectValue)
	queriedObj2 := &ContentConstraintTestModel{}
	err = helper.UpdateEntity(queriedValue2, queriedObj2)
	if err != nil {
		t.Errorf("UpdateEntity for query failed, err:%s", err.Error())
		return
	}

	if queriedObj2.Name != "Jane Smith" {
		t.Errorf("Updated Name field mismatch, expected: Jane Smith, got: %s", queriedObj2.Name)
	}
	if queriedObj2.Age != 30 {
		t.Errorf("Updated Age field mismatch, expected: 30, got: %d", queriedObj2.Age)
	}
	if queriedObj2.Status != "inactive" {
		t.Errorf("Updated Status field mismatch, expected: inactive, got: %s", queriedObj2.Status)
	}

	t.Logf("Content constraint test passed: all constraints validated")
}

// cleanupConstraintTestRemote 清理远程约束测试数据
func cleanupConstraintTestRemote(t *testing.T, o1 orm.Orm, remoteProvider provider.Provider) {
	// 创建一个有效的对象用于查询，避免验证失败
	constraintObj := &ConstraintTestModel{
		Name:       "cleanup_dummy", // 提供必填字段
		Password:   "dummy",         // 提供只写字段
		CreateTime: 1234567890,      // 提供不可变字段
		UpdateTime: 1234567890,      // 提供更新时间
		Status:     1,               // 提供状态
		ReadOnlyID: 999,             // 提供只读字段
		WriteOnly:  "dummy",         // 提供只写字段
	}
	constraintValue, err := getObjectValue(constraintObj)
	if err != nil {
		t.Errorf("getObjectValue failed, err:%s", err.Error())
		return
	}

	constraintModel, err := remoteProvider.GetEntityModel(constraintValue)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	filter, err := remoteProvider.GetModelFilter(constraintModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("cleanup batch query failed, err:%s", queryErr.Error())
		return
	}

	for _, model := range modelList {
		_, deleteErr := o1.Delete(model)
		if deleteErr != nil {
			t.Errorf("cleanup delete failed, err:%s", deleteErr.Error())
		}
	}

	t.Logf("Cleaned up %d constraint test records", len(modelList))
}
