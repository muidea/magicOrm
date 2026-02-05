// Package metricsdb provides database-specific metric definitions for MagicORM.
package metricsdb

import (
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"
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
			// 记录错误但不影响初始化
		}
	}
}
