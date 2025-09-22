package test

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// TestLocalTransaction 测试本地事务功能
func TestLocalTransaction(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("transaction_local")

	o1, err := orm.NewOrm(localProvider, config, "transaction_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&Unit{}}
	modelList, modelErr := registerLocalModel(localProvider, objList)
	if modelErr != nil {
		t.Errorf("register model failed. err:%s", modelErr.Error())
		return
	}

	// 先删除再创建表结构
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

	// 开始事务
	txErr := o1.BeginTransaction()
	if txErr != nil {
		t.Errorf("begin transaction failed, err:%s", txErr.Error())
		return
	}

	// 准备数据
	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	obj1 := &Unit{
		I8:        8,
		I16:       1600,
		I32:       323200,
		I64:       uint64(78962222222),
		Name:      "Transaction Test 1",
		Value:     12.3456,
		F64:       12.45678,
		TimeStamp: now,
		Flag:      true,
		IArray:    []int{1, 2, 3},
		FArray:    []float32{1.1, 2.2, 3.3},
		StrArray:  []string{"a", "b", "c"},
	}

	obj2 := &Unit{
		I8:        9,
		I16:       1700,
		I32:       323300,
		I64:       uint64(78962222223),
		Name:      "Transaction Test 2",
		Value:     22.3456,
		F64:       22.45678,
		TimeStamp: now,
		Flag:      false,
		IArray:    []int{4, 5, 6},
		FArray:    []float32{4.4, 5.5, 6.6},
		StrArray:  []string{"d", "e", "f"},
	}

	// 在事务中插入第一个对象
	obj1Model, obj1Err := localProvider.GetEntityModel(obj1)
	if obj1Err != nil {
		o1.RollbackTransaction()
		t.Errorf("GetEntityModel failed, err:%s", obj1Err.Error())
		return
	}

	obj1Model, obj1Err = o1.Insert(obj1Model)
	if obj1Err != nil {
		o1.RollbackTransaction()
		t.Errorf("insert obj1 in transaction failed, err:%s", obj1Err.Error())
		return
	}
	obj1 = obj1Model.Interface(true).(*Unit)

	// 在事务中插入第二个对象
	obj2Model, obj2Err := localProvider.GetEntityModel(obj2)
	if obj2Err != nil {
		o1.RollbackTransaction()
		t.Errorf("GetEntityModel failed, err:%s", obj2Err.Error())
		return
	}

	obj2Model, obj2Err = o1.Insert(obj2Model)
	if obj2Err != nil {
		o1.RollbackTransaction()
		t.Errorf("insert obj2 in transaction failed, err:%s", obj2Err.Error())
		return
	}
	_ = obj2Model.Interface(true).(*Unit)

	// 提交事务
	commitErr := o1.CommitTransaction()
	if commitErr != nil {
		t.Errorf("commit transaction failed, err:%s", commitErr.Error())
		return
	}

	// 验证事务提交后的数据
	queryObj1 := &Unit{ID: obj1.ID}
	queryObj1Model, queryObj1Err := localProvider.GetEntityModel(queryObj1)
	if queryObj1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", queryObj1Err.Error())
		return
	}

	queryObj1Model, queryObj1Err = o1.Query(queryObj1Model)
	if queryObj1Err != nil {
		t.Errorf("query obj1 failed, err:%s", queryObj1Err.Error())
		return
	}
	queryObj1 = queryObj1Model.Interface(true).(*Unit)

	// 测试回滚功能
	tx2Err := o1.BeginTransaction()
	if tx2Err != nil {
		t.Errorf("begin transaction 2 failed, err:%s", tx2Err.Error())
		return
	}

	// 修改对象1
	queryObj1.Name = "Modified in transaction and will rollback"
	modObj1Model, modObj1Err := localProvider.GetEntityModel(queryObj1)
	if modObj1Err != nil {
		o1.RollbackTransaction()
		t.Errorf("GetEntityModel failed, err:%s", modObj1Err.Error())
		return
	}

	_, modObj1Err = o1.Update(modObj1Model)
	if modObj1Err != nil {
		o1.RollbackTransaction()
		t.Errorf("update obj1 in transaction failed, err:%s", modObj1Err.Error())
		return
	}

	// 回滚事务
	o1.RollbackTransaction()

	// 验证回滚后的数据
	checkObj1 := &Unit{ID: obj1.ID}
	checkObj1Model, checkObj1Err := localProvider.GetEntityModel(checkObj1)
	if checkObj1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", checkObj1Err.Error())
		return
	}

	checkObj1Model, checkObj1Err = o1.Query(checkObj1Model)
	if checkObj1Err != nil {
		t.Errorf("query obj1 after rollback failed, err:%s", checkObj1Err.Error())
		return
	}
	checkObj1 = checkObj1Model.Interface(true).(*Unit)

	// 验证回滚是否成功，名称应该保持原样
	if checkObj1.Name != "Transaction Test 1" {
		t.Errorf("transaction rollback failed, expected name 'Transaction Test 1', got: %s", checkObj1.Name)
		return
	}

	// 清理测试数据
	_, delErr1 := o1.Delete(obj1Model)
	if delErr1 != nil {
		t.Errorf("delete obj1 failed, err:%s", delErr1.Error())
		return
	}

	_, delErr2 := o1.Delete(obj2Model)
	if delErr2 != nil {
		t.Errorf("delete obj2 failed, err:%s", delErr2.Error())
		return
	}
}

// TestLocalBatchOperation 测试批量操作功能
func TestLocalBatchOperation(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("batch_local")

	o1, err := orm.NewOrm(localProvider, config, "batch_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&Unit{}}
	modelList, modelErr := registerLocalModel(localProvider, objList)
	if modelErr != nil {
		t.Errorf("register model failed. err:%s", modelErr.Error())
		return
	}

	// 先删除再创建表结构
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

	// 批量插入数据
	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	batchSize := 20
	unitList := make([]*Unit, 0, batchSize)
	modelList = make([]model.Model, 0, batchSize)

	// 准备数据
	for i := 0; i < batchSize; i++ {
		unit := &Unit{
			I8:        int8(i % 127),
			I16:       int16(i * 100),
			I32:       int32(i * 1000),
			I64:       uint64(i * 10000),
			Name:      "Batch Item " + string(rune(65+i)), // A, B, C, ...
			Value:     float32(i) * 1.5,
			F64:       float64(i) * 2.5,
			TimeStamp: now,
			Flag:      i%2 == 0,
			IArray:    []int{i, i + 1, i + 2},
			FArray:    []float32{float32(i) * 0.5, float32(i) * 1.5, float32(i) * 2.5},
			StrArray:  []string{string(rune(65 + i)), string(rune(66 + i)), string(rune(67 + i))},
		}
		unitList = append(unitList, unit)

		unitModel, unitErr := localProvider.GetEntityModel(unit)
		if unitErr != nil {
			err = unitErr
			t.Errorf("GetEntityModel failed. err:%s", err.Error())
			return
		}
		modelList = append(modelList, unitModel)
	}

	// 批量插入 (单独插入多条)
	insertedModelList := make([]model.Model, 0, batchSize)
	for idx := 0; idx < batchSize; idx++ {
		unitModel, unitErr := o1.Insert(modelList[idx])
		if unitErr != nil {
			err = unitErr
			t.Errorf("Insert failed. err:%s", err.Error())
			return
		}
		insertedModelList = append(insertedModelList, unitModel)
		unitList[idx] = unitModel.Interface(true).(*Unit)
	}

	// 使用过滤器批量查询
	unitModel, _ := localProvider.GetEntityModel(&Unit{})
	filter, err := localProvider.GetModelFilter(unitModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	// 查询偶数ID的记录
	filter.Equal("flag", true)
	bqModelList, bqModelErr := o1.BatchQuery(filter)
	if bqModelErr != nil {
		t.Errorf("BatchQuery failed, err:%s", bqModelErr.Error())
		return
	}
	if len(bqModelList) != batchSize/2 {
		t.Errorf("batch query flag=true failed, expected %d records, got %d", batchSize/2, len(bqModelList))
		return
	}

	// 使用LIKE查询
	filter2, err := localProvider.GetModelFilter(unitModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}
	filter2.Like("name", "Batch Item")
	bq2ModelList, bq2ModelErr := o1.BatchQuery(filter2)
	if bq2ModelErr != nil {
		t.Errorf("BatchQuery with LIKE failed, err:%s", bq2ModelErr.Error())
		return
	}
	if len(bq2ModelList) != batchSize {
		t.Errorf("batch query with LIKE failed, expected %d records, got %d", batchSize, len(bq2ModelList))
		return
	}

	// 分页查询测试
	filter3, err := localProvider.GetModelFilter(unitModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}
	pageSize := 5
	filter3.Pagination(0, pageSize) // 第一页，每页5条
	bq3ModelList, bq3ModelErr := o1.BatchQuery(filter3)
	if bq3ModelErr != nil {
		t.Errorf("BatchQuery with pagination failed, err:%s", bq3ModelErr.Error())
		return
	}
	if len(bq3ModelList) != pageSize {
		t.Errorf("batch query with pagination failed, expected %d records, got %d", pageSize, len(bq3ModelList))
		return
	}

	// 排序查询测试
	filter4, err := localProvider.GetModelFilter(unitModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}
	filter4.Sort("value", false) // value降序
	bq4ModelList, bq4ModelErr := o1.BatchQuery(filter4)
	if bq4ModelErr != nil {
		t.Errorf("BatchQuery with sorting failed, err:%s", bq4ModelErr.Error())
		return
	}
	if len(bq4ModelList) != batchSize {
		t.Errorf("batch query with sorting failed, expected %d records, got %d", batchSize, len(bq4ModelList))
		return
	}

	// 验证排序结果
	for i := 0; i < len(bq4ModelList)-1; i++ {
		unit1 := bq4ModelList[i].Interface(true).(*Unit)
		unit2 := bq4ModelList[i+1].Interface(true).(*Unit)
		if unit1.Value < unit2.Value {
			t.Errorf("sorting failed, expected desc order but got asc")
			return
		}
	}

	// 批量删除测试
	for idx := 0; idx < batchSize; idx++ {
		_, delErr := o1.Delete(insertedModelList[idx])
		if delErr != nil {
			err = delErr
			t.Errorf("Delete failed. err:%s", err.Error())
			return
		}
	}

	// 验证删除结果
	count, countErr := o1.Count(filter2)
	if countErr != nil {
		t.Errorf("count object after batch delete failed, err:%s", countErr.Error())
		return
	}
	if count != 0 {
		t.Errorf("batch delete failed, expected 0 records, got %d", count)
		return
	}
}
