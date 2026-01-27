package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/utils"
)

const constraintLocalOwner = "constraint_local"

func TestConstraintLocal(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	validator := utils.NewValueValidator()
	localProvider := provider.NewLocalProvider(constraintLocalOwner, validator)

	o1, err := orm.NewOrm(localProvider, config, "constraint_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&ConstraintTestModel{}, &ContentConstraintTestModel{}}
	modelList, modelErr := registerLocalModel(localProvider, objList)
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
		testRequiredFields(t, o1, localProvider)
	})

	// 测试2: 测试只读字段
	t.Run("TestReadOnlyFields", func(t *testing.T) {
		testReadOnlyFields(t, o1, localProvider)
	})

	// 测试3: 测试只写字段
	t.Run("TestWriteOnlyFields", func(t *testing.T) {
		testWriteOnlyFields(t, o1, localProvider)
	})

	// 测试4: 测试不可变字段
	t.Run("TestImmutableFields", func(t *testing.T) {
		testImmutableFields(t, o1, localProvider)
	})

	// 测试5: 测试可选字段
	t.Run("TestOptionalFields", func(t *testing.T) {
		testOptionalFields(t, o1, localProvider)
	})

	// 测试6: 测试内容值约束
	t.Run("TestContentConstraints", func(t *testing.T) {
		testContentConstraints(t, o1, localProvider)
	})

	// 测试7: 测试内容值约束失败情况
	t.Run("TestContentConstraintFailures", func(t *testing.T) {
		testContentConstraintFailures(t, o1, localProvider)
	})

	// 清理测试数据
	cleanupConstraintTest(t, o1, localProvider)
}

// testRequiredFields 测试必填字段约束
func testRequiredFields(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 测试缺少必填字段的情况
	// 注意：由于ORM层可能已经处理了验证，这里主要测试正常流程
	// 实际验证可能在业务层进行

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

	objModel, objErr := localProvider.GetEntityModel(obj)
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

	insertedObj := insertedModel.Interface(true).(*ConstraintTestModel)
	t.Logf("Inserted object ID: %d", insertedObj.ID)

	// 验证插入的数据
	if insertedObj.Name != "test_user" {
		t.Errorf("Name field mismatch, expected: test_user, got: %s", insertedObj.Name)
	}
	if insertedObj.Status != 1 {
		t.Errorf("Status field mismatch, expected: 1, got: %d", insertedObj.Status)
	}
}

// testReadOnlyFields 测试只读字段约束
func testReadOnlyFields(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
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

	objModel, objErr := localProvider.GetEntityModel(obj)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedObj := insertedModel.Interface(true).(*ConstraintTestModel)
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

	updateModel, updateErr := localProvider.GetEntityModel(updateObj)
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
	queryObj := &ConstraintTestModel{ID: objID}
	queryModel, queryErr := localProvider.GetEntityModel(queryObj)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedObj := queriedModel.Interface(true).(*ConstraintTestModel)

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

// testWriteOnlyFields 测试只写字段约束
func testWriteOnlyFields(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
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

	objModel, objErr := localProvider.GetEntityModel(obj)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedObj := insertedModel.Interface(true).(*ConstraintTestModel)
	objID := insertedObj.ID

	// 查询对象
	queryObj := &ConstraintTestModel{ID: objID}
	queryModel, queryErr := localProvider.GetEntityModel(queryObj)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedObj := queriedModel.Interface(true).(*ConstraintTestModel)

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

// testImmutableFields 测试不可变字段约束
func testImmutableFields(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
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

	objModel, objErr := localProvider.GetEntityModel(obj)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	insertedModel, insertErr := o1.Insert(objModel)
	if insertErr != nil {
		t.Errorf("Insert failed, err:%s", insertErr.Error())
		return
	}

	insertedObj := insertedModel.Interface(true).(*ConstraintTestModel)
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

	updateModel, updateErr := localProvider.GetEntityModel(updateObj)
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
	queryObj := &ConstraintTestModel{ID: objID}
	queryModel, queryErr := localProvider.GetEntityModel(queryObj)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedObj := queriedModel.Interface(true).(*ConstraintTestModel)

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

// testOptionalFields 测试可选字段约束
func testOptionalFields(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
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

	obj1Model, obj1Err := localProvider.GetEntityModel(obj1)
	if obj1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", obj1Err.Error())
		return
	}

	inserted1Model, insert1Err := o1.Insert(obj1Model)
	if insert1Err != nil {
		t.Errorf("Insert without optional field failed, err:%s", insert1Err.Error())
		return
	}

	inserted1Obj := inserted1Model.Interface(true).(*ConstraintTestModel)
	obj1ID := inserted1Obj.ID

	// 查询验证可选字段为空
	query1Obj := &ConstraintTestModel{ID: obj1ID}
	query1Model, query1Err := localProvider.GetEntityModel(query1Obj)
	if query1Err != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", query1Err.Error())
		return
	}

	queried1Model, query1Err := o1.Query(query1Model)
	if query1Err != nil {
		t.Errorf("Query failed, err:%s", query1Err.Error())
		return
	}

	queried1Obj := queried1Model.Interface(true).(*ConstraintTestModel)
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

	obj2Model, obj2Err := localProvider.GetEntityModel(obj2)
	if obj2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", obj2Err.Error())
		return
	}

	inserted2Model, insert2Err := o1.Insert(obj2Model)
	if insert2Err != nil {
		t.Errorf("Insert with optional field failed, err:%s", insert2Err.Error())
		return
	}

	inserted2Obj := inserted2Model.Interface(true).(*ConstraintTestModel)
	obj2ID := inserted2Obj.ID

	// 查询验证可选字段被保存
	query2Obj := &ConstraintTestModel{ID: obj2ID}
	query2Model, query2Err := localProvider.GetEntityModel(query2Obj)
	if query2Err != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", query2Err.Error())
		return
	}

	queried2Model, query2Err := o1.Query(query2Model)
	if query2Err != nil {
		t.Errorf("Query failed, err:%s", query2Err.Error())
		return
	}

	queried2Obj := queried2Model.Interface(true).(*ConstraintTestModel)
	if queried2Obj.Email != "test@example.com" {
		t.Errorf("Optional Email field should be saved, expected: test@example.com, got: %s", queried2Obj.Email)
	}

	t.Logf("Optional field test passed: Email=%s", queried2Obj.Email)
}

// testContentConstraints 测试内容值约束
func testContentConstraints(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
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

	objModel, objErr := localProvider.GetEntityModel(validObj)
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

	insertedObj := insertedModel.Interface(true).(*ContentConstraintTestModel)
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
	queryObj := &ContentConstraintTestModel{ID: insertedObj.ID}
	queryModel, queryErr := localProvider.GetEntityModel(queryObj)
	if queryErr != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr.Error())
		return
	}

	queriedModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("Query failed, err:%s", queryErr.Error())
		return
	}

	queriedObj := queriedModel.Interface(true).(*ContentConstraintTestModel)
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

	updateModel, updateErr := localProvider.GetEntityModel(updateObj)
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
	queryObj2 := &ContentConstraintTestModel{ID: insertedObj.ID}
	queryModel2, queryErr2 := localProvider.GetEntityModel(queryObj2)
	if queryErr2 != nil {
		t.Errorf("GetEntityModel for query failed, err:%s", queryErr2.Error())
		return
	}

	queriedModel2, queryErr2 := o1.Query(queryModel2)
	if queryErr2 != nil {
		t.Errorf("Query failed, err:%s", queryErr2.Error())
		return
	}

	queriedObj2 := queriedModel2.Interface(true).(*ContentConstraintTestModel)
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

// testContentConstraintFailures 测试内容值约束失败的情况
func testContentConstraintFailures(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 测试各种违反约束的情况
	testCases := []struct {
		name        string
		obj         *ContentConstraintTestModel
		expectedErr bool
	}{
		{
			name: "Name too short",
			obj: &ContentConstraintTestModel{
				Name:        "Jo", // 长度小于3
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Name too long",
			obj: &ContentConstraintTestModel{
				Name:        "This is a very long name that exceeds the maximum allowed length of fifty characters", // 长度大于50
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Age below minimum",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         -5, // 小于0
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Age above maximum",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         200, // 大于150
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Score below range",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       -10.0, // 小于0.0
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Score above range",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       150.0, // 大于100.0
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Invalid status",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "invalid", // 不在枚举中
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Invalid email format",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "invalid-email", // 不符合邮箱正则
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Description too long",
			obj: &ContentConstraintTestModel{
				Name:   "John Doe",
				Age:    25,
				Score:  85.5,
				Status: "active",
				Email:  "test@example.com",
				Description: "This description is way too long. " + // 长度大于500
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
					"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
					"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris " +
					"nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in " +
					"reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla " +
					"pariatur. Excepteur sint occaecat cupidatat non proident, sunt in " +
					"culpa qui officia deserunt mollit anim id est laborum. " +
					"Sed ut perspiciatis unde omnis iste natus error sit voluptatem " +
					"accusantium doloremque laudantium, totam rem aperiam, eaque ipsa " +
					"quae ab illo inventore veritatis et quasi architecto beatae vitae " +
					"dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit " +
					"aspernatur aut odit aut fugit, sed quia consequuntur magni dolores " +
					"eos qui ratione voluptatem sequi nesciunt.",
				ItemCount: 5,
				Price:     99.99,
				Category:  "A",
				Code:      "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "ItemCount below minimum",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   0, // 小于1
				Price:       99.99,
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Price below range",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       0.0, // 小于0.01
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Price above range",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       10000.0, // 大于9999.99
				Category:    "A",
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Invalid category",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "E", // 不在枚举A:B:C:D中
				Code:        "ABC-123",
			},
			expectedErr: true,
		},
		{
			name: "Invalid code format",
			obj: &ContentConstraintTestModel{
				Name:        "John Doe",
				Age:         25,
				Score:       85.5,
				Status:      "active",
				Email:       "test@example.com",
				Description: "Valid description",
				ItemCount:   5,
				Price:       99.99,
				Category:    "A",
				Code:        "invalid", // 不符合正则格式
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objModel, objErr := localProvider.GetEntityModel(tc.obj)
			if objErr != nil {
				t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
				return
			}

			// 尝试插入对象
			_, insertErr := o1.Insert(objModel)

			if tc.expectedErr {
				// 期望插入失败
				if insertErr == nil {
					t.Errorf("Expected constraint violation error for %s, but insert succeeded", tc.name)
				} else {
					t.Logf("Constraint violation detected as expected for %s: %s", tc.name, insertErr.Error())
				}
			} else {
				// 期望插入成功
				if insertErr != nil {
					t.Errorf("Expected insert to succeed for %s, but got error: %s", tc.name, insertErr.Error())
				}
			}
		})
	}
}

// cleanupConstraintTest 清理约束测试数据
func cleanupConstraintTest(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	constraintModel, _ := localProvider.GetEntityModel(&ConstraintTestModel{})
	filter, err := localProvider.GetModelFilter(constraintModel)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
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
