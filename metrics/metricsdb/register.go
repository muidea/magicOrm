// Package metricsdb provides database-specific metric definitions for MagicORM.
package metricsdb

import (
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"
	"log/slog"
)

var (
	// Global database metrics collector
	databaseMetricCollector *DatabaseMetricsCollector
	// Global database metric provider
	databaseMetricProvider *DatabaseMetricProvider
)

// GetDatabaseMetricsCollector returns the global database metrics collector.
func GetDatabaseMetricsCollector() *DatabaseMetricsCollector {
	return databaseMetricCollector
}

// RegisterDatabaseMetrics registers database metrics with the global monitoring system.
func RegisterDatabaseMetrics() {
	// 创建全局metrics收集器
	databaseMetricCollector = NewDatabaseMetricsCollector()

	// 只有在GlobalManager存在时才注册provider
	if mgr := monitoring.GetGlobalManager(); mgr != nil {
		// 创建provider并传递collector
		databaseMetricProvider = NewDatabaseMetricProviderWithCollector(databaseMetricCollector)

		// 尝试注册DatabaseMetricProvider
		if err := monitoring.RegisterGlobalProvider(
			"magicorm_database",
			func() types.MetricProvider {
				return databaseMetricProvider
			},
			true, // 自动初始化
			100,  // 优先级
		); err != nil {
			databaseMetricProvider = nil
			slog.Warn("Failed to register database metrics provider", "error", err.Error())
		}
	}
}

// EnsureDatabaseMetricProviderRegistered attempts to register the provider after GlobalManager becomes available.
func EnsureDatabaseMetricProviderRegistered() {
	if databaseMetricProvider != nil {
		return
	}
	if databaseMetricCollector == nil {
		return
	}
	if monitoring.GetGlobalManager() == nil {
		return
	}

	databaseMetricProvider = NewDatabaseMetricProviderWithCollector(databaseMetricCollector)
	if err := monitoring.RegisterGlobalProvider(
		"magicorm_database",
		func() types.MetricProvider {
			return databaseMetricProvider
		},
		true,
		100,
	); err != nil {
		databaseMetricProvider = nil
		slog.Warn("Failed to ensure database metrics provider registration", "error", err.Error())
	} else {
		slog.Info("Database metrics provider ensured and registered successfully")
	}
}
