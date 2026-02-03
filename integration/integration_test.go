// 简单的集成测试
package integration

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/monitoring/core"
)

func TestSimpleIntegration(t *testing.T) {
	t.Skip("Skipping due to magicCommon monitoring manager global state issues")
	// 创建测试配置 - 禁用导出以避免端口冲突
	config := core.DefaultMonitoringConfig()
	config.Namespace = "magicorm_test"
	config.AsyncCollection = false
	config.ExportConfig.Enabled = false // 禁用导出

	// 创建集成实例（单例）
	integration, err := NewSimpleIntegration(&config)
	if err != nil {
		t.Fatalf("创建集成失败: %v", err)
	}
	defer integration.Stop()

	// 测试基本功能
	t.Run("TestBasicOperations", func(t *testing.T) {
		// 这些操作应该不会崩溃
		integration.RecordORMOperation(
			"insert",
			"TestUser",
			true,
			100*time.Millisecond,
			nil,
			map[string]string{"test": "basic"},
		)

		integration.RecordValidationOperation(
			"test_validator",
			"TestModel",
			"insert",
			50*time.Millisecond,
			nil,
			map[string]string{"test": "basic"},
		)

		integration.RecordDatabaseOperation(
			"postgresql",
			"select",
			true,
			200*time.Millisecond,
			5,
			nil,
			map[string]string{"test": "basic"},
		)
	})

	t.Run("TestErrorOperations", func(t *testing.T) {
		integration.RecordORMOperation(
			"update",
			"TestUser",
			false,
			150*time.Millisecond,
			&testError{message: "test error"},
			map[string]string{"error_test": "true"},
		)
	})

	t.Run("TestManagerState", func(t *testing.T) {
		manager := integration.GetManager()
		if manager == nil {
			t.Error("管理器不应该为nil")
		}

		if !manager.IsRunning() {
			t.Error("管理器应该正在运行")
		}

		if !manager.IsInitialized() {
			t.Error("管理器应该已初始化")
		}
	})
}

// 测试错误类型
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
