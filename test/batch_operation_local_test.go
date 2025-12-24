package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// 批量操作测试用的数据结构
type BatchItem struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	Name      string    `orm:"name index(name_idx)" view:"detail,lite"`
	Value     float64   `orm:"value" view:"detail,lite"`
	Status    int       `orm:"status" view:"detail,lite"`
	CreatedAt time.Time `orm:"createdAt" view:"detail,lite"`
}

// TestBatchOperations 测试批量操作功能
func TestBatchOperations(t *testing.T) {
	// 跳过测试如果设置了环境变量
	if testing.Short() {
		t.Skip("skipping batch operations test in short mode.")
	}

	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("batch_local")

	o1, err := orm.NewOrm(localProvider, config, "batch_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&BatchItem{}}
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

	// 测试批量插入
	t.Run("BatchInsert", func(t *testing.T) {
		testBatchInsert(t, o1, localProvider)
	})

	// 测试批量查询
	t.Run("BatchQuery", func(t *testing.T) {
		testBatchQuery(t, o1, localProvider)
	})

	// 测试批量更新
	t.Run("BatchUpdate", func(t *testing.T) {
		testBatchUpdate(t, o1, localProvider)
	})

	// 测试批量删除
	t.Run("BatchDelete", func(t *testing.T) {
		testBatchDelete(t, o1, localProvider)
	})

	// 测试批量操作性能
	t.Run("BatchPerformance", func(t *testing.T) {
		testBatchPerformance(t, o1, localProvider)
	})

	// 清理测试数据
	cleanupBatchTest(t, o1, localProvider)
}

// 测试批量插入
func testBatchInsert(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	itemCount := 100
	modelList := make([]models.Model, itemCount)

	insertStartTime := time.Now()

	// 准备批量插入的数据
	for i := 0; i < itemCount; i++ {
		item := &BatchItem{
			Name:      fmt.Sprintf("Batch_Insert_Item_%d", i),
			Value:     float64(i * 10),
			Status:    i % 5,
			CreatedAt: time.Now(),
		}

		itemModel, itemErr := localProvider.GetEntityModel(item)
		if itemErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", itemErr.Error())
			return
		}

		modelList[i] = itemModel
	}

	// 进行批量插入
	for _, m := range modelList {
		_, insertErr := o1.Insert(m)
		if insertErr != nil {
			t.Errorf("insert failed, err:%s", insertErr.Error())
			return
		}
	}

	insertDuration := time.Since(insertStartTime)
	t.Logf("Batch insert of %d items took: %v, avg: %v per item",
		itemCount, insertDuration, insertDuration/time.Duration(itemCount))

	// 验证插入的数据
	batchItemModel, _ := localProvider.GetEntityModel(&BatchItem{})
	filter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	filter.Like("name", "Batch_Insert_Item%")
	queryModelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("batch query failed, err:%s", queryErr.Error())
		return
	}

	if len(queryModelList) != itemCount {
		t.Errorf("Expected %d items after batch insert, but got %d", itemCount, len(queryModelList))
	}
}

// 测试批量查询
func testBatchQuery(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 进行各种条件的批量查询测试
	batchItemModel, _ := localProvider.GetEntityModel(&BatchItem{})

	// 1. 等值查询
	equalStartTime := time.Now()
	equalFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	equalFilter.Equal("status", 1)
	equalModelList, equalQueryErr := o1.BatchQuery(equalFilter)
	equalDuration := time.Since(equalStartTime)

	if equalQueryErr != nil {
		t.Errorf("equal batch query failed, err:%s", equalQueryErr.Error())
		return
	}

	t.Logf("Equal batch query returned %d items, took: %v", len(equalModelList), equalDuration)

	// 2. 范围查询
	rangeStartTime := time.Now()
	rangeFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	rangeFilter.Above("value", 200.0)
	rangeFilter.Below("value", 700.0)
	rangeModelList, rangeQueryErr := o1.BatchQuery(rangeFilter)
	rangeDuration := time.Since(rangeStartTime)

	if rangeQueryErr != nil {
		t.Errorf("range batch query failed, err:%s", rangeQueryErr.Error())
		return
	}

	t.Logf("Range batch query returned %d items, took: %v", len(rangeModelList), rangeDuration)

	// 3. 模糊查询
	likeStartTime := time.Now()
	likeFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	likeFilter.Like("name", "Item_5")
	likeModelList, likeQueryErr := o1.BatchQuery(likeFilter)
	likeDuration := time.Since(likeStartTime)

	if likeQueryErr != nil {
		t.Errorf("like batch query failed, err:%s", likeQueryErr.Error())
		return
	}

	t.Logf("Like batch query returned %d items, took: %v", len(likeModelList), likeDuration)

	// 4. 多条件组合查询
	combinedStartTime := time.Now()
	combinedFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	combinedFilter.Like("name", "Item_")
	combinedFilter.Above("value", 300.0)
	combinedFilter.Equal("status", 0)
	combinedModelList, combinedQueryErr := o1.BatchQuery(combinedFilter)
	combinedDuration := time.Since(combinedStartTime)

	if combinedQueryErr != nil {
		t.Errorf("combined batch query failed, err:%s", combinedQueryErr.Error())
		return
	}

	t.Logf("Combined batch query returned %d items, took: %v", len(combinedModelList), combinedDuration)

	// 5. 分页查询
	pageStartTime := time.Now()
	pageFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	pageFilter.Pagination(1, 20) // 第1页，每页20条
	pageModelList, pageQueryErr := o1.BatchQuery(pageFilter)
	pageDuration := time.Since(pageStartTime)

	if pageQueryErr != nil {
		t.Errorf("page batch query failed, err:%s", pageQueryErr.Error())
		return
	}

	t.Logf("Page batch query returned %d items, took: %v", len(pageModelList), pageDuration)
}

// 测试批量更新
func testBatchUpdate(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 首先查询要更新的数据
	batchItemModel, _ := localProvider.GetEntityModel(&BatchItem{})
	filter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	filter.Equal("status", 2)
	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("query for batch update failed, err:%s", queryErr.Error())
		return
	}

	updateCount := len(modelList)
	if updateCount == 0 {
		t.Logf("No items found for batch update test")
		return
	}

	// 更新这些数据
	updateStartTime := time.Now()

	for i, itemModel := range modelList {
		item := itemModel.Interface(true).(*BatchItem)
		item.Value = item.Value + 1000.0
		item.Name = item.Name + "_Updated"

		updatedModel, updateErr := localProvider.GetEntityModel(item)
		if updateErr != nil {
			t.Errorf("GetEntityModel for update failed, err:%s", updateErr.Error())
			continue
		}

		modelList[i] = updatedModel
	}

	// 执行批量更新
	for _, m := range modelList {
		_, updateErr := o1.Update(m)
		if updateErr != nil {
			t.Errorf("update failed, err:%s", updateErr.Error())
			return
		}
	}

	updateDuration := time.Since(updateStartTime)
	t.Logf("Batch update of %d items took: %v, avg: %v per item",
		updateCount, updateDuration, updateDuration/time.Duration(updateCount))

	// 验证更新结果
	updatedFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	updatedFilter.Like("name", "%_Updated")
	updatedModelList, updatedQueryErr := o1.BatchQuery(updatedFilter)
	if updatedQueryErr != nil {
		t.Errorf("query after batch update failed, err:%s", updatedQueryErr.Error())
		return
	}

	if len(updatedModelList) != updateCount {
		t.Errorf("Expected %d updated items, but got %d", updateCount, len(updatedModelList))
	}
}

// 测试批量删除
func testBatchDelete(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 首先查询要删除的数据
	batchItemModel, _ := localProvider.GetEntityModel(&BatchItem{})
	filter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	filter.Equal("status", 3)
	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("query for batch delete failed, err:%s", queryErr.Error())
		return
	}

	deleteCount := len(modelList)
	if deleteCount == 0 {
		t.Logf("No items found for batch delete test")
		return
	}

	// 执行批量删除
	deleteStartTime := time.Now()
	for _, m := range modelList {
		_, deleteErr := o1.Delete(m)
		if deleteErr != nil {
			t.Errorf("delete failed, err:%s", deleteErr.Error())
			return
		}
	}

	deleteDuration := time.Since(deleteStartTime)
	t.Logf("Batch delete of %d items took: %v, avg: %v per item",
		deleteCount, deleteDuration, deleteDuration/time.Duration(deleteCount))

	// 验证删除结果
	verifyFilter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	verifyFilter.Equal("status", 3)
	verifyModelList, verifyQueryErr := o1.BatchQuery(verifyFilter)
	if verifyQueryErr != nil {
		t.Errorf("query after batch delete failed, err:%s", verifyQueryErr.Error())
		return
	}

	if len(verifyModelList) != 0 {
		t.Errorf("Expected 0 items after delete, but got %d", len(verifyModelList))
	}
}

// 测试批量操作性能
func testBatchPerformance(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 清理现有数据
	cleanupBatchTest(t, o1, localProvider)

	// 大批量数据测试
	largeCount := 1000
	modelList := make([]models.Model, largeCount)

	// 准备批量插入的数据
	for i := 0; i < largeCount; i++ {
		item := &BatchItem{
			Name:      fmt.Sprintf("Perf_Item_%d", i),
			Value:     float64(i),
			Status:    i % 5,
			CreatedAt: time.Now(),
		}

		itemModel, itemErr := localProvider.GetEntityModel(item)
		if itemErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", itemErr.Error())
			return
		}

		modelList[i] = itemModel
	}

	// 测试批量插入性能
	insertStartTime := time.Now()
	for _, m := range modelList {
		_, insertErr := o1.Insert(m)
		if insertErr != nil {
			t.Errorf("large batch insert failed, err:%s", insertErr.Error())
			return
		}
	}
	insertDuration := time.Since(insertStartTime)
	t.Logf("Large batch insert of %d items took: %v, avg: %v per item",
		largeCount, insertDuration, insertDuration/time.Duration(largeCount))

	// 测试批量查询性能
	queryStartTime := time.Now()
	batchItemModel, _ := localProvider.GetEntityModel(&BatchItem{})
	filter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	filter.Like("name", "Perf_Item%")
	queryModelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("large batch query failed, err:%s", queryErr.Error())
		return
	}
	queryDuration := time.Since(queryStartTime)
	t.Logf("Large batch query of %d items took: %v, avg: %v per item",
		len(queryModelList), queryDuration, queryDuration/time.Duration(len(queryModelList)))

	// 测试单条插入与批量插入性能对比
	singleInsertStartTime := time.Now()
	for i := 0; i < 100; i++ {
		item := &BatchItem{
			Name:      fmt.Sprintf("Single_Insert_Item_%d", i),
			Value:     float64(i),
			Status:    i % 5,
			CreatedAt: time.Now(),
		}

		itemModel, itemErr := localProvider.GetEntityModel(item)
		if itemErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", itemErr.Error())
			continue
		}

		_, insertErr := o1.Insert(itemModel)
		if insertErr != nil {
			t.Errorf("single insert failed, err:%s", insertErr.Error())
			continue
		}
	}
	singleInsertDuration := time.Since(singleInsertStartTime)
	t.Logf("100 single inserts took: %v, avg: %v per item",
		singleInsertDuration, singleInsertDuration/100)

	// 计算性能提升
	singleItemAvg := singleInsertDuration / 100
	batchItemAvg := insertDuration / time.Duration(largeCount)
	speedup := float64(singleItemAvg) / float64(batchItemAvg)
	t.Logf("Batch insert is approximately %.2fx faster than single inserts", speedup)
}

// 清理批量操作测试中创建的数据
func cleanupBatchTest(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	batchItemModel, _ := localProvider.GetEntityModel(&BatchItem{})
	filter, err := localProvider.GetModelFilter(batchItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("cleanup batch query failed, err:%s", queryErr.Error())
		return
	}

	if len(modelList) > 0 {
		for _, m := range modelList {
			_, deleteErr := o1.Delete(m)
			if deleteErr != nil {
				t.Errorf("cleanup batch delete failed, err:%s", deleteErr.Error())
			}
		}
	}
}
