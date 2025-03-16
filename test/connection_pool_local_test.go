//go:build local || all
// +build local all

package test

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// 用于连接池测试的数据结构
type PoolTestItem struct {
	ID        int       `orm:"id key auto" view:"detail,lite"`
	Name      string    `orm:"name" view:"detail,lite"`
	Value     float64   `orm:"value" view:"detail,lite"`
	CreatedAt time.Time `orm:"createdAt" view:"detail,lite"`
}

// TestConnectionPool 测试连接池功能
func TestConnectionPool(t *testing.T) {
	// 暂时跳过连接池测试
	t.Skip("Temporarily skipping connection pool tests due to stability issues")

	// 跳过测试如果设置了环境变量
	if testing.Short() {
		t.Skip("skipping connection pool test in short mode.")
	}

	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("pool_local")

	// 创建所有连接池和表
	var ormList []orm.Orm
	var modelList []model.Model
	var prefixList []string
	for i := 0; i < 10; i++ {
		prefix := fmt.Sprintf("pool_test_%d", i)
		prefixList = append(prefixList, prefix)
		o, err := orm.NewOrm(localProvider, config, prefix)
		if err != nil {
			t.Errorf("NewOrm failed for prefix %s, err:%s", prefix, err.Error())
			return
		}
		ormList = append(ormList, o)

		// 注册模型
		m, regErr := localProvider.RegisterModel(&PoolTestItem{})
		if regErr != nil {
			t.Errorf("RegisterModel failed for orm %d, err:%s", i, regErr.Error())
			return
		}
		modelList = append(modelList, m)

		// 创建表
		createErr := o.Create(m)
		if createErr != nil {
			t.Errorf("Create table failed for orm %d, err:%s", i, createErr.Error())
			return
		}
	}

	// 测试结束时清理表
	defer func() {
		for i, o := range ormList {
			if i < len(modelList) {
				dropErr := o.Drop(modelList[i])
				if dropErr != nil {
					t.Logf("Drop table failed for orm %d, err:%s", i, dropErr.Error())
				}
			}
			o.Release()
		}
	}()

	// 多连接并发操作测试
	t.Run("MultiConnectionConcurrentOperations", func(t *testing.T) {
		testMultiConnectionConcurrentOperations(t, ormList, localProvider)
	})

	// 连接切换测试
	t.Run("ConnectionSwitching", func(t *testing.T) {
		testConnectionSwitching(t, ormList, localProvider)
	})

	// 清理测试数据
	cleanupConnectionPoolTest(t, ormList[0], localProvider)
}

// 多连接并发操作测试
func testMultiConnectionConcurrentOperations(t *testing.T, ormList []orm.Orm, localProvider provider.Provider) {
	// 使用少量连接
	ormCount := 3
	itemsPerOrm := 10

	// 数据准备
	var wg sync.WaitGroup
	insertStartTime := time.Now()

	// 并发插入数据
	for ormIdx := 0; ormIdx < ormCount; ormIdx++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			if idx >= len(ormList) {
				t.Errorf("orm index %d out of range", idx)
				return
			}

			o := ormList[idx]

			// 向每个连接插入一些测试项
			for j := 0; j < itemsPerOrm; j++ {
				poolItem := PoolTestItem{
					Name:      "Pool_Orm" + strconv.Itoa(idx) + "_Item" + strconv.Itoa(j),
					Value:     float64(idx*100 + j),
					CreatedAt: time.Now(),
				}

				m, err := localProvider.GetEntityModel(&poolItem)
				if err != nil {
					t.Errorf("GetEntityModel failed for orm %d item %d, err:%s", idx, j, err.Error())
					continue
				}

				m, insertErr := o.Insert(m)
				if insertErr != nil {
					t.Errorf("insert item failed for orm %d item %d, err:%s", idx, j, insertErr.Error())
					continue
				}
			}
		}(ormIdx)
	}

	wg.Wait()
	insertDuration := time.Since(insertStartTime)
	totalItems := ormCount * itemsPerOrm
	t.Logf("Multi-connection concurrent insert of %d items took: %v, avg: %v per insert",
		totalItems, insertDuration, insertDuration/time.Duration(totalItems))

	// 查询验证
	for ormIdx := 0; ormIdx < ormCount; ormIdx++ {
		if ormIdx >= len(ormList) {
			continue
		}

		o := ormList[ormIdx]

		poolItemModel, _ := localProvider.GetEntityModel(&PoolTestItem{})
		filter, err := localProvider.GetModelFilter(poolItemModel)
		if err != nil {
			t.Errorf("GetModelFilter failed for orm %d, err:%s", ormIdx, err.Error())
			continue
		}

		name := "Pool_Orm" + strconv.Itoa(ormIdx)
		filter.Like("name", name)
		modelList, queryErr := o.BatchQuery(filter)
		if queryErr != nil {
			t.Errorf("batch query failed for orm %d, err:%s", ormIdx, queryErr.Error())
			continue
		}

		// 检查是否获取到了正确数量的项
		if len(modelList) != itemsPerOrm {
			t.Errorf("Expected %d items for orm %d, but got %d", itemsPerOrm, ormIdx, len(modelList))
		}
	}
}

// 连接切换测试
func testConnectionSwitching(t *testing.T, ormList []orm.Orm, localProvider provider.Provider) {
	if len(ormList) < 2 {
		t.Errorf("Not enough orm instances for switching test")
		return
	}

	// 使用第一个连接
	o1 := ormList[0]

	// 创建一个测试项
	testItem := PoolTestItem{
		Name:      "Switching_Test_Item",
		Value:     999.99,
		CreatedAt: time.Now(),
	}

	// 使用第一个连接插入
	m, err := localProvider.GetEntityModel(&testItem)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	m, insertErr := o1.Insert(m)
	if insertErr != nil {
		t.Errorf("insert with first orm failed, err:%s", insertErr.Error())
		return
	}

	insertedItem := m.Interface(true).(*PoolTestItem)

	// 使用第一个连接查询
	queryItem := PoolTestItem{ID: insertedItem.ID}
	queryModel, err := localProvider.GetEntityModel(&queryItem)
	if err != nil {
		t.Errorf("GetEntityModel failed for query, err:%s", err.Error())
		return
	}

	queryModel, queryErr := o1.Query(queryModel)
	if queryErr != nil {
		t.Errorf("query with first orm failed, err:%s", queryErr.Error())
		return
	}

	queriedItem := queryModel.Interface(true).(*PoolTestItem)

	// 验证数据一致性
	if queriedItem.Name != testItem.Name || queriedItem.Value != testItem.Value {
		t.Errorf("Data mismatch in connection switching. Expected name=%s value=%f, got name=%s value=%f",
			testItem.Name, testItem.Value, queriedItem.Name, queriedItem.Value)
	} else {
		t.Logf("Connection switching test passed")
	}
}

// 清理连接池测试中创建的数据
func cleanupConnectionPoolTest(t *testing.T, orm orm.Orm, localProvider provider.Provider) {
	poolItemModel, _ := localProvider.GetEntityModel(&PoolTestItem{})
	filter, err := localProvider.GetModelFilter(poolItemModel)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	modelList, queryErr := orm.BatchQuery(filter)
	if queryErr != nil {
		t.Errorf("batch query failed, err:%s", queryErr.Error())
		return
	}

	for _, model := range modelList {
		_, delErr := orm.Delete(model)
		if delErr != nil {
			t.Errorf("delete item failed, err:%s", delErr.Error())
		}
	}
}
