// Package validation provides validation-specific metric definitions for MagicORM.
package validation

import (
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"
)

var (
	// Global validation metrics collector
	validationMetricCollector *ValidationMetricsCollector
	// Global validation metric provider
	validationMetricProvider *ValidationMetricProvider
)

// GetValidationMetricsCollector returns the global validation metrics collector.
func GetValidationMetricsCollector() *ValidationMetricsCollector {
	return validationMetricCollector
}

// RegisterValidationMetrics registers validation metrics with the global monitoring system.
func RegisterValidationMetrics() {
	// 创建全局metrics收集器
	validationMetricCollector = NewValidationMetricsCollector()

	// 只有在GlobalManager存在时才注册provider
	if mgr := monitoring.GetGlobalManager(); mgr != nil {
		// 创建provider并传递collector
		validationMetricProvider = NewValidationMetricProviderWithCollector(validationMetricCollector)

		// 尝试注册ValidationMetricProvider
		if err := monitoring.RegisterGlobalProvider(
			"magicorm_validation",
			func() types.MetricProvider {
				return validationMetricProvider
			},
			true, // 自动初始化
			100,  // 优先级
		); err != nil {
			validationMetricProvider = nil
			// 记录错误但不影响初始化
		}
	}
}
