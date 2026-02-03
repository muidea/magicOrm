// 简单的集成测试 - 不使用 magicCommon 监控管理器
package integration

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/monitoring"
	"github.com/muidea/magicOrm/monitoring/database"
	"github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/validation"
)

func TestSimpleCollectors(t *testing.T) {
	t.Run("TestORMCollector", func(t *testing.T) {
		collector := orm.NewCollector()

		// 测试记录操作
		collector.RecordOperation(
			monitoring.OperationInsert,
			"TestUser",
			time.Now(),
			nil,
			map[string]string{"test": "orm"},
		)

		// 测试获取指标
		metrics, err := collector.GetMetrics()
		if err != nil {
			t.Errorf("获取指标失败: %v", err)
		}
		if metrics == nil {
			t.Error("指标不应该为nil")
		}
	})

	t.Run("TestValidationCollector", func(t *testing.T) {
		collector := validation.NewCollector()

		collector.RecordValidation(
			"test_validator",
			"TestModel",
			"insert",
			time.Now(),
			nil,
			map[string]string{"test": "validation"},
		)

		metrics, err := collector.GetMetrics()
		if err != nil {
			t.Errorf("获取指标失败: %v", err)
		}
		if metrics == nil {
			t.Error("指标不应该为nil")
		}
	})

	t.Run("TestDatabaseCollector", func(t *testing.T) {
		collector := database.NewCollector()

		collector.RecordQuery(
			"postgresql",
			"SELECT",
			10,
			time.Now(),
			nil,
			map[string]string{"test": "database"},
		)

		metrics, err := collector.GetMetrics()
		if err != nil {
			t.Errorf("获取指标失败: %v", err)
		}
		if metrics == nil {
			t.Error("指标不应该为nil")
		}
	})

	t.Run("TestErrorHandling", func(t *testing.T) {
		collector := orm.NewCollector()

		// 测试错误记录
		collector.RecordOperation(
			monitoring.OperationUpdate,
			"TestUser",
			time.Now(),
			&simpleTestError{message: "test error"},
			map[string]string{"error_test": "true"},
		)

		metrics, err := collector.GetMetrics()
		if err != nil {
			t.Errorf("获取指标失败: %v", err)
		}
		if metrics == nil {
			t.Error("指标不应该为nil")
		}
	})
}

// 测试错误类型
type simpleTestError struct {
	message string
}

func (e *simpleTestError) Error() string {
	return e.message
}
