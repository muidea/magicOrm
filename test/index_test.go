package test

import (
	"strconv"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// 定义带索引的测试数据结构
type IndexTestItem struct {
	ID          int       `orm:"id key auto" view:"detail,lite"`
	Name        string    `orm:"name index(name_idx)" view:"detail,lite"` // 单列索引
	Value       float64   `orm:"value index(value_idx)" view:"detail,lite"` // 单列索引
	Category    string    `orm:"category index(cat_value_idx)" view:"detail,lite"` // 组合索引的一部分
	Status      int       `orm:"status index(cat_value_idx)" view:"detail,lite"` // 组合索引的一部分
	CreatedAt   time.Time `orm:"createdAt" view:"detail,lite"`
}

// TestIndexFeatures 测试索引功能
func TestIndexFeatures(t *testing.T) {
	// 跳过测试如果设置了环境变量
	if testing.Short() {
		t.Skip("skipping index test in short mode.")
	}

	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("index_local")

	o1, err := orm.NewOrm(localProvider, config, "index_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&IndexTestItem{}}
	modelList, modelErr := registerModel(localProvider, objList)
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

	// 测试单列索引性能
	t.Run("SingleColumnIndexPerformance", func(t *testing.T) {
		testSingleColumnIndexPerformance(t, o1, localProvider)
	})

	// 测试组合索引性能
	t.Run("CompositeIndexPerformance", func(t *testing.T) {
		testCompositeIndexPerformance(t, o1, localProvider)
	})

	// 清理测试数据
	cleanupIndexTest(t, o1, localProvider)
}

// 插入测试数据
func insertIndexTestData(t *testing.T, o1 orm.Orm, localProvider provider.Provider, count int) {
	categories := []string{"A", "B", "C", "D", "E"}
	statuses := []int{0, 1, 2, 3, 4}

	// 批量插入测试数据
	for i := 0; i < count; i++ {
		item := &IndexTestItem{
			Name:      "Index_" + strconv.Itoa(i),
			Value:     float64(i % 100),
			Category:  categories[i%len(categories)],
			Status:    statuses[i%len(statuses)],
			CreatedAt: time.Now(),
		}

		itemModel, itemErr := localProvider.GetEntityModel(item)
		if itemErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", itemErr.Error())
			return
		}

		_, itemErr = o1.Insert(itemModel)
		if itemErr != nil {
			t.Errorf("insert item failed, err:%s", itemErr.Error())
			return
		}
	}
}

// 测试单列索引性能
func testSingleColumnIndexPerformance(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 先插入足够多的测试数据
	dataCount := 500
	insertIndexTestData(t, o1, localProvider, dataCount)

	// 测试使用索引字段查询
	indexTestItemModel, _ := localProvider.GetEntityModel(&IndexTestItem{})

	// 测试使用 Name 索引的查询性能
	startTimeNameIndex := time.Now()
	nameFilter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	nameFilter.Equal("name", "Index_100")
	nameModelList, nameQueryErr := o1.BatchQuery(nameFilter)
	nameQueryDuration := time.Since(startTimeNameIndex)

	if nameQueryErr != nil {
		t.Errorf("name index query failed, err:%s", nameQueryErr.Error())
		return
	}

	t.Logf("Query with Name index took: %v, results: %d", nameQueryDuration, len(nameModelList))

	// 测试使用 Value 索引的查询性能
	startTimeValueIndex := time.Now()
	valueFilter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	valueFilter.Equal("value", 50.0)
	valueModelList, valueQueryErr := o1.BatchQuery(valueFilter)
	valueQueryDuration := time.Since(startTimeValueIndex)

	if valueQueryErr != nil {
		t.Errorf("value index query failed, err:%s", valueQueryErr.Error())
		return
	}

	t.Logf("Query with Value index took: %v, results: %d", valueQueryDuration, len(valueModelList))

	// 测试使用非索引字段的查询性能（CreatedAt）
	startTimeNoIndex := time.Now()
	noIndexFilter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	
	// 因为CreatedAt是时间类型，这里我们使用Above来构建一个范围查询
	noIndexFilter.Above("createdAt", time.Now().Add(-time.Hour))
	noIndexModelList, noIndexQueryErr := o1.BatchQuery(noIndexFilter)
	noIndexQueryDuration := time.Since(startTimeNoIndex)

	if noIndexQueryErr != nil {
		t.Errorf("no index query failed, err:%s", noIndexQueryErr.Error())
		return
	}

	t.Logf("Query without index took: %v, results: %d", noIndexQueryDuration, len(noIndexModelList))
}

// 测试组合索引性能
func testCompositeIndexPerformance(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	indexTestItemModel, _ := localProvider.GetEntityModel(&IndexTestItem{})

	// 测试使用组合索引的所有字段查询
	startTimeFullCompositeIndex := time.Now()
	fullCompositeFilter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	fullCompositeFilter.Equal("category", "A")
	fullCompositeFilter.Equal("status", 1)
	fullCompositeModelList, fullCompositeQueryErr := o1.BatchQuery(fullCompositeFilter)
	fullCompositeQueryDuration := time.Since(startTimeFullCompositeIndex)

	if fullCompositeQueryErr != nil {
		t.Errorf("full composite index query failed, err:%s", fullCompositeQueryErr.Error())
		return
	}

	t.Logf("Query with full composite index took: %v, results: %d", 
		fullCompositeQueryDuration, len(fullCompositeModelList))

	// 测试使用组合索引的第一个字段查询
	startTimePartialCompositeIndex := time.Now()
	partialCompositeFilter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	partialCompositeFilter.Equal("category", "B")
	partialCompositeModelList, partialCompositeQueryErr := o1.BatchQuery(partialCompositeFilter)
	partialCompositeQueryDuration := time.Since(startTimePartialCompositeIndex)

	if partialCompositeQueryErr != nil {
		t.Errorf("partial composite index query failed, err:%s", partialCompositeQueryErr.Error())
		return
	}

	t.Logf("Query with partial composite index took: %v, results: %d", 
		partialCompositeQueryDuration, len(partialCompositeModelList))

	// 测试使用组合索引的第二个字段查询（通常不会使用索引）
	startTimeNonLeadingIndex := time.Now()
	nonLeadingFilter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}
	nonLeadingFilter.Equal("status", 2)
	nonLeadingModelList, nonLeadingQueryErr := o1.BatchQuery(nonLeadingFilter)
	nonLeadingQueryDuration := time.Since(startTimeNonLeadingIndex)

	if nonLeadingQueryErr != nil {
		t.Errorf("non-leading composite index query failed, err:%s", nonLeadingQueryErr.Error())
		return
	}

	t.Logf("Query with non-leading composite index field took: %v, results: %d", 
		nonLeadingQueryDuration, len(nonLeadingModelList))
}

// 清理索引测试中创建的数据
func cleanupIndexTest(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	indexTestItemModel, _ := localProvider.GetEntityModel(&IndexTestItem{})
	filter, err := localProvider.GetModelFilter(indexTestItemModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("batch query failed, err:%s", queryErr.Error())
		return
	}

	for _, model := range modelList {
		_, delErr := o1.Delete(model)
		if delErr != nil {
			t.Errorf("delete item failed, err:%s", delErr.Error())
		}
	}
}
