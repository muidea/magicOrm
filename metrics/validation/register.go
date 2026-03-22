// Package validation provides validation-specific metric definitions for MagicORM.
package validation

import (
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/types"
	"log/slog"
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
			slog.Warn("Failed to register validation metrics provider", "error", err.Error())
		}
	}
}

// EnsureValidationMetricProviderRegistered attempts to register the provider after GlobalManager becomes available.
func EnsureValidationMetricProviderRegistered() {
	if validationMetricProvider != nil {
		return
	}
	if validationMetricCollector == nil {
		return
	}
	if monitoring.GetGlobalManager() == nil {
		return
	}

	validationMetricProvider = NewValidationMetricProviderWithCollector(validationMetricCollector)
	if err := monitoring.RegisterGlobalProvider(
		"magicorm_validation",
		func() types.MetricProvider {
			return validationMetricProvider
		},
		true,
		100,
	); err != nil {
		validationMetricProvider = nil
		slog.Warn("Failed to ensure validation metrics provider registration", "error", err.Error())
	} else {
		slog.Info("Validation metrics provider ensured and registered successfully")
	}
}
