// 简单的性能基准测试
package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/muidea/magicOrm/monitoring"
	"github.com/muidea/magicOrm/monitoring/database"
	"github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/validation"
)

// BenchmarkORMCollector 测试ORM收集器的性能
func BenchmarkORMCollector(b *testing.B) {
	collector := orm.NewCollector()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			collector.RecordOperation(
				monitoring.OperationInsert,
				fmt.Sprintf("User%d", i%100),
				time.Now(),
				nil,
				map[string]string{"test": "benchmark", "id": fmt.Sprintf("%d", i)},
			)
			i++
		}
	})
}

// BenchmarkValidationCollector 测试验证收集器的性能
func BenchmarkValidationCollector(b *testing.B) {
	collector := validation.NewCollector()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			collector.RecordValidation(
				"benchmark_validator",
				fmt.Sprintf("Model%d", i%100),
				"insert",
				time.Now(),
				nil,
				map[string]string{"test": "benchmark", "id": fmt.Sprintf("%d", i)},
			)
			i++
		}
	})
}

// BenchmarkDatabaseCollector 测试数据库收集器的性能
func BenchmarkDatabaseCollector(b *testing.B) {
	collector := database.NewCollector()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			collector.RecordQuery(
				"postgresql",
				"SELECT",
				10, // rowsAffected
				time.Now(),
				nil, // err
				map[string]string{"test": "benchmark", "id": fmt.Sprintf("%d", i)},
			)
			i++
		}
	})
}

// BenchmarkConcurrentOperations 测试并发操作性能
func BenchmarkConcurrentOperations(b *testing.B) {
	ormCollector := orm.NewCollector()
	valCollector := validation.NewCollector()
	dbCollector := database.NewCollector()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 混合不同类型的操作
			switch i % 3 {
			case 0:
				ormCollector.RecordOperation(
					monitoring.OperationInsert,
					fmt.Sprintf("ConcurrentModel%d", i%100),
					time.Now(),
					nil,
					map[string]string{"type": "orm", "id": fmt.Sprintf("%d", i)},
				)
			case 1:
				valCollector.RecordValidation(
					"concurrent_validator",
					fmt.Sprintf("ConcurrentModel%d", i%100),
					"insert",
					time.Now(),
					nil,
					map[string]string{"type": "validation", "id": fmt.Sprintf("%d", i)},
				)
			case 2:
				dbCollector.RecordQuery(
					"postgresql",
					"SELECT",
					5, // rowsAffected
					time.Now(),
					nil, // err
					map[string]string{"type": "database", "id": fmt.Sprintf("%d", i)},
				)
			}
			i++
		}
	})
}

// BenchmarkErrorHandling 测试错误处理性能
func BenchmarkErrorHandling(b *testing.B) {
	collector := orm.NewCollector()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			var err error
			if i%10 == 0 { // 10%的错误率
				err = fmt.Errorf("benchmark error %d", i)
			}

			collector.RecordOperation(
				monitoring.OperationUpdate,
				fmt.Sprintf("ErrorModel%d", i%100),
				time.Now(),
				err,
				map[string]string{"error_test": "true", "id": fmt.Sprintf("%d", i)},
			)
			i++
		}
	})
}

// BenchmarkGetMetrics 测试获取指标的性能
func BenchmarkGetMetrics(b *testing.B) {
	collector := orm.NewCollector()

	// 先记录一些数据
	for i := 0; i < 1000; i++ {
		collector.RecordOperation(
			monitoring.OperationInsert,
			fmt.Sprintf("MetricModel%d", i%100),
			time.Now(),
			nil,
			map[string]string{"preload": "true", "id": fmt.Sprintf("%d", i)},
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics, err := collector.GetMetrics()
		if err != nil {
			b.Errorf("获取指标失败: %v", err)
		}
		_ = metrics
	}
}
