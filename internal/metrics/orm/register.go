package orm

import (
	"log/slog"

	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"
)

var ormMetricProvider *ORMMetricProvider
var ormMetricCollector *ORMMetricsCollector

// GetORMMetricsCollector returns the collector shared by runtime operations and monitoring.
func GetORMMetricsCollector() *ORMMetricsCollector { return ormMetricCollector }

// RegisterORMMetrics attaches the ORM runtime collector to the monitoring provider.
func RegisterORMMetrics(collector *ORMMetricsCollector) {
	// Retain the runtime collector even when monitoring starts later.
	ormMetricCollector = collector

	// 只有在GlobalManager存在时才注册provider
	if mgr := monitoring.GetGlobalManager(); mgr != nil {
		// 创建provider并传递collector
		ormMetricProvider = NewORMMetricProvider(ormMetricCollector)

		// 尝试注册ORMMetricProvider
		if err := monitoring.RegisterGlobalProvider(
			"magicorm_orm",
			func() types.MetricProvider {
				return ormMetricProvider
			},
			true, // 自动初始化
			100,  // 优先级
		); err != nil {
			ormMetricProvider = nil
			// 记录错误但不影响ORM初始化
			slog.Warn("Failed to register ORM metrics provider", "error", err.Error())
		} else {
			slog.Info("ORM metrics provider registered successfully")
		}
	} else {
		// GlobalManager不存在，只创建collector不注册provider
		slog.Debug("GlobalManager not available, ORM metrics collector created but provider not registered")
	}
}

// EnsureORMMetricProviderRegistered 尝试在GlobalManager可用时注册ORM metrics provider。
// 该方法是幂等的，可在monitoring.InitializeGlobalManager()之后显式调用。
func EnsureORMMetricProviderRegistered() {
	// 如果已经有provider，则不需要重复注册
	if ormMetricProvider != nil {
		return
	}
	if ormMetricCollector == nil {
		// 尚未初始化ORM，保持静默
		return
	}

	if monitoring.GetGlobalManager() == nil {
		// 监控系统尚未初始化，无法注册
		return
	}

	// 创建并注册provider
	ormMetricProvider = NewORMMetricProvider(ormMetricCollector)
	if err := monitoring.RegisterGlobalProvider(
		"magicorm_orm",
		func() types.MetricProvider {
			return ormMetricProvider
		},
		true,
		100,
	); err != nil {
		ormMetricProvider = nil
		slog.Warn("Failed to ensure ORM metrics provider registration", "error", err.Error())
	} else {
		slog.Info("ORM metrics provider ensured and registered successfully")
	}
}
