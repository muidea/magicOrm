package test

import (
	"strconv"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// 简单对象，用于性能测试
type SimplePerf struct {
	ID    int    `orm:"id key auto" view:"detail,lite"`
	Name  string `orm:"name" view:"detail,lite"`
	Value int    `orm:"value" view:"detail,lite"`
}

// 复杂对象，包含嵌套结构，用于性能测试
type ComplexPerf struct {
	ID            int           `orm:"id key auto" view:"detail,lite"`
	Name          string        `orm:"name" view:"detail,lite"`
	Value         float64       `orm:"value" view:"detail,lite"`
	CreatedAt     time.Time     `orm:"createdAt" view:"detail,lite"`
	Tags          []string      `orm:"tags" view:"detail,lite"`
	SimpleList    []SimplePerf  `orm:"simpleList" view:"detail,lite"`
	SimplePtrList []*SimplePerf `orm:"simplePtrList" view:"detail,lite"`
}

// TestPerformance 性能测试主函数
func TestPerformance(t *testing.T) {
	// 跳过性能测试如果设置了环境变量
	if testing.Short() {
		t.Skip("skipping performance test in short mode.")
	}

	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("performance_local")

	o1, err := orm.NewOrm(localProvider, config, "performance_test")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	// 注册模型
	objList := []any{&SimplePerf{}, &ComplexPerf{}}
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

	// 批量插入简单对象性能测试
	t.Run("BulkInsertSimple", func(t *testing.T) {
		benchmarkBulkInsertSimple(t, o1, localProvider, 1000)
	})

	// 批量查询简单对象性能测试
	t.Run("BulkQuerySimple", func(t *testing.T) {
		benchmarkBulkQuerySimple(t, o1, localProvider)
	})

	// 插入复杂对象性能测试
	t.Run("InsertComplex", func(t *testing.T) {
		benchmarkInsertComplex(t, o1, localProvider, 50)
	})

	// 查询复杂对象性能测试
	t.Run("QueryComplex", func(t *testing.T) {
		benchmarkQueryComplex(t, o1, localProvider)
	})

	// 过滤器性能测试
	t.Run("FilterPerformance", func(t *testing.T) {
		benchmarkFilterPerformance(t, o1, localProvider)
	})

	// 清理测试数据
	cleanupPerformanceTest(t, o1, localProvider)
}

// 批量插入简单对象性能测试
func benchmarkBulkInsertSimple(t *testing.T, o1 orm.Orm, localProvider provider.Provider, count int) {
	startTime := time.Now()

	// 准备数据
	for i := 0; i < count; i++ {
		simpleObj := &SimplePerf{
			Name:  "Simple_" + strconv.Itoa(i),
			Value: i,
		}

		simpleObjModel, simpleObjErr := localProvider.GetEntityModel(simpleObj)
		if simpleObjErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", simpleObjErr.Error())
			return
		}

		_, simpleObjErr = o1.Insert(simpleObjModel)
		if simpleObjErr != nil {
			t.Errorf("insert simple object failed, err:%s", simpleObjErr.Error())
			return
		}
	}

	duration := time.Since(startTime)
	t.Logf("Bulk insert %d simple objects took: %v, avg: %v per insert",
		count, duration, duration/time.Duration(count))
}

// 批量查询简单对象性能测试
func benchmarkBulkQuerySimple(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	startTime := time.Now()

	simplePerfModel, _ := localProvider.GetEntityModel(&SimplePerf{})
	filter, err := localProvider.GetModelFilter(simplePerfModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("batch query failed, err:%s", queryErr.Error())
		return
	}

	count := len(modelList)
	duration := time.Since(startTime)
	t.Logf("Batch query %d simple objects took: %v, avg: %v per object",
		count, duration, duration/time.Duration(count))
}

// 插入复杂对象性能测试
func benchmarkInsertComplex(t *testing.T, o1 orm.Orm, localProvider provider.Provider, count int) {
	// 先创建一些简单对象作为关联对象
	simpleObjList := []*SimplePerf{}
	for i := 0; i < 10; i++ {
		simpleObj := &SimplePerf{
			Name:  "ComplexRef_" + strconv.Itoa(i),
			Value: i * 100,
		}

		simpleObjModel, simpleObjErr := localProvider.GetEntityModel(simpleObj)
		if simpleObjErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", simpleObjErr.Error())
			return
		}

		simpleObjModel, simpleObjErr = o1.Insert(simpleObjModel)
		if simpleObjErr != nil {
			t.Errorf("insert simple object failed, err:%s", simpleObjErr.Error())
			return
		}

		simpleObjList = append(simpleObjList, simpleObjModel.Interface(true).(*SimplePerf))
	}

	startTime := time.Now()

	// 创建并插入复杂对象
	for i := 0; i < count; i++ {
		// 准备嵌套的简单对象列表
		simpleList := []SimplePerf{}
		for j := 0; j < 3; j++ { // 每个复杂对象包含3个简单对象
			simpleList = append(simpleList, SimplePerf{
				Name:  "Nested_" + strconv.Itoa(i) + "_" + strconv.Itoa(j),
				Value: i*10 + j,
			})
		}

		// 准备嵌套的简单对象指针列表
		simplePtrList := []*SimplePerf{}
		for j := 0; j < 2; j++ { // 每个复杂对象引用2个已存在的简单对象
			idx := (i + j) % len(simpleObjList)
			simplePtrList = append(simplePtrList, simpleObjList[idx])
		}

		// 创建复杂对象
		complexObj := &ComplexPerf{
			Name:          "Complex_" + strconv.Itoa(i),
			Value:         float64(i) * 1.5,
			CreatedAt:     time.Now(),
			Tags:          []string{"tag1", "tag2", "performance"},
			SimpleList:    simpleList,
			SimplePtrList: simplePtrList,
		}

		complexObjModel, complexObjErr := localProvider.GetEntityModel(complexObj)
		if complexObjErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", complexObjErr.Error())
			return
		}

		_, complexObjErr = o1.Insert(complexObjModel)
		if complexObjErr != nil {
			t.Errorf("insert complex object failed, err:%s", complexObjErr.Error())
			return
		}
	}

	duration := time.Since(startTime)
	t.Logf("Insert %d complex objects took: %v, avg: %v per insert",
		count, duration, duration/time.Duration(count))
}

// 查询复杂对象性能测试
func benchmarkQueryComplex(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	startTime := time.Now()

	complexPerfModel, _ := localProvider.GetEntityModel(&ComplexPerf{})
	filter, err := localProvider.GetModelFilter(complexPerfModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	modelList, queryErr := o1.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("batch query failed, err:%s", queryErr.Error())
		return
	}

	count := len(modelList)
	queryDuration := time.Since(startTime)
	t.Logf("Batch query %d complex objects took: %v, avg: %v per object",
		count, queryDuration, queryDuration/time.Duration(count))

	// 测试单个复杂对象的查询和序列化性能
	if count > 0 {
		complexObj := modelList[0].Interface(true).(*ComplexPerf)

		startDeserializeTime := time.Now()

		// 测试反序列化性能
		for i := 0; i < 100; i++ {
			_ = modelList[i%count].Interface(true).(*ComplexPerf)
		}

		deserializeDuration := time.Since(startDeserializeTime)
		t.Logf("Deserialize 100 complex objects took: %v, avg: %v per object",
			deserializeDuration, deserializeDuration/100)

		// 查询单个复杂对象性能
		singleQueryStart := time.Now()

		queryObj := &ComplexPerf{ID: complexObj.ID}
		queryObjModel, queryObjErr := localProvider.GetEntityModel(queryObj)
		if queryObjErr != nil {
			t.Errorf("GetEntityModel failed, err:%s", queryObjErr.Error())
			return
		}

		_, queryObjErr = o1.Query(queryObjModel)
		if queryObjErr != nil {
			t.Errorf("query complex object failed, err:%s", queryObjErr.Error())
			return
		}

		singleQueryDuration := time.Since(singleQueryStart)
		t.Logf("Single complex object query took: %v", singleQueryDuration)
	}
}

// 过滤器性能测试
func benchmarkFilterPerformance(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	simplePerfModel, _ := localProvider.GetEntityModel(&SimplePerf{})

	// 1. 等值查询
	equalFilterStart := time.Now()

	equalFilter, err := localProvider.GetModelFilter(simplePerfModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	equalFilter.Equal("value", 10)
	equalModelList, equalQueryErr := o1.BatchQuery(equalFilter)
	if equalQueryErr != nil {
		t.Errorf("equal filter query failed, err:%s", equalQueryErr.Error())
		return
	}

	equalFilterDuration := time.Since(equalFilterStart)
	t.Logf("Equal filter query returned %d objects, took: %v",
		len(equalModelList), equalFilterDuration)

	// 2. 范围查询
	rangeFilterStart := time.Now()

	rangeFilter, err := localProvider.GetEntityFilter(&SimplePerf{}, models.MetaView)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	// 替换Between为Above和Below组合
	rangeFilter.Above("value", 100)
	rangeFilter.Below("value", 200)
	rangeModelList, rangeQueryErr := o1.BatchQuery(rangeFilter)
	if rangeQueryErr != nil {
		t.Errorf("range filter query failed, err:%s", rangeQueryErr.Error())
		return
	}

	rangeFilterDuration := time.Since(rangeFilterStart)
	t.Logf("Range filter query returned %d objects, took: %v",
		len(rangeModelList), rangeFilterDuration)

	// 3. 模糊查询
	likeFilterStart := time.Now()

	likeFilter, err := localProvider.GetModelFilter(simplePerfModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	likeFilter.Like("name", "Simple_1")
	likeModelList, likeQueryErr := o1.BatchQuery(likeFilter)
	if likeQueryErr != nil {
		t.Errorf("like filter query failed, err:%s", likeQueryErr.Error())
		return
	}

	likeFilterDuration := time.Since(likeFilterStart)
	t.Logf("Like filter query returned %d objects, took: %v",
		len(likeModelList), likeFilterDuration)

	// 4. 复合查询
	complexFilterStart := time.Now()

	complexFilter, err := localProvider.GetEntityFilter(&SimplePerf{}, models.MetaView)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	complexFilter.Above("value", 500) // 替换 GreaterThan
	complexFilter.Below("value", 800) // 替换 LessThan
	complexFilter.Like("name", "Simple_")
	complexFilter.Sort("value", false) // 降序
	complexModelList, complexQueryErr := o1.BatchQuery(complexFilter)
	if complexQueryErr != nil {
		t.Errorf("complex filter query failed, err:%s", complexQueryErr.Error())
		return
	}

	complexFilterDuration := time.Since(complexFilterStart)
	t.Logf("Complex filter query returned %d objects, took: %v",
		len(complexModelList), complexFilterDuration)

	// 5. 分页查询
	for pageSize := 10; pageSize <= 100; pageSize *= 10 {
		for pageIndex := 0; pageIndex < 3; pageIndex++ {
			pageFilterStart := time.Now()

			pageFilter, err := localProvider.GetModelFilter(simplePerfModel)
			if err != nil {
				t.Errorf("GetModelFilter failed, err:%s", err.Error())
				return
			}

			// 设置排序和分页
			pageFilter.Sort("value", true) // 升序
			pageFilter.Pagination(pageIndex, pageSize)
			pageModelList, pageQueryErr := o1.BatchQuery(pageFilter)
			if pageQueryErr != nil {
				t.Errorf("page filter query failed, err:%s", pageQueryErr.Error())
				return
			}

			pageFilterDuration := time.Since(pageFilterStart)
			t.Logf("Page filter query (page=%d, size=%d) returned %d objects, took: %v",
				pageIndex, pageSize, len(pageModelList), pageFilterDuration)
		}
	}
}

// 清理性能测试中创建的数据
func cleanupPerformanceTest(t *testing.T, o1 orm.Orm, localProvider provider.Provider) {
	// 删除复杂对象
	complexPerfModel, _ := localProvider.GetEntityModel(&ComplexPerf{})
	complexFilter, err := localProvider.GetModelFilter(complexPerfModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	complexModelList, complexQueryErr := o1.BatchQuery(complexFilter)
	if complexQueryErr != nil {
		t.Errorf("batch query complex objects failed, err:%s", complexQueryErr.Error())
		return
	}

	for _, model := range complexModelList {
		_, delErr := o1.Delete(model)
		if delErr != nil {
			t.Errorf("delete complex object failed, err:%s", delErr.Error())
		}
	}

	// 删除简单对象
	simplePerfModel, _ := localProvider.GetEntityModel(&SimplePerf{})
	simpleFilter, err := localProvider.GetModelFilter(simplePerfModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	simpleModelList, simpleQueryErr := o1.BatchQuery(simpleFilter)
	if simpleQueryErr != nil {
		t.Errorf("batch query simple objects failed, err:%s", simpleQueryErr.Error())
		return
	}

	for _, model := range simpleModelList {
		_, delErr := o1.Delete(model)
		if delErr != nil {
			t.Errorf("delete simple object failed, err:%s", delErr.Error())
		}
	}
}
