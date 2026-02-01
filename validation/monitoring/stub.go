// Package monitoring provides backward compatibility for validation monitoring.
// This is a stub implementation that forwards to the new unified monitoring system.
package monitoring

import (
	"fmt"
	"time"

	"github.com/muidea/magicOrm/monitoring/core"
	monitoringv2 "github.com/muidea/magicOrm/monitoring/validation"
)

// MetricsCollector is a stub for backward compatibility
type MetricsCollector struct {
	collector *core.Collector
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *core.MonitoringConfig) *MetricsCollector {
	collector := core.NewCollector(config)
	return &MetricsCollector{collector: collector}
}

// GetMetrics returns collected metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	if m.collector == nil {
		return make(map[string]interface{})
	}

	metrics := m.collector.GetMetrics()
	result := make(map[string]interface{})
	for k, v := range metrics {
		result[k] = v
	}
	return result
}

// ValidationLogger is a stub for backward compatibility
type ValidationLogger struct {
	monitor *monitoringv2.ValidationMonitor
}

// NewValidationLogger creates a new validation logger
func NewValidationLogger(collector *MetricsCollector, config *core.MonitoringConfig) *ValidationLogger {
	monitor := monitoringv2.NewValidationMonitor(collector.collector, config)
	return &ValidationLogger{monitor: monitor}
}

// RecordValidation records a validation operation
func (l *ValidationLogger) RecordValidation(
	operation string,
	modelName string,
	scenario string,
	duration time.Duration,
	err error,
	fields map[string]interface{},
) {
	if l.monitor == nil {
		return
	}

	// Convert scenario string to enum
	var scenarioEnum monitoringv2.Scenario
	switch scenario {
	case "insert":
		scenarioEnum = monitoringv2.ScenarioInsert
	case "update":
		scenarioEnum = monitoringv2.ScenarioUpdate
	case "query":
		scenarioEnum = monitoringv2.ScenarioQuery
	case "delete":
		scenarioEnum = monitoringv2.ScenarioDelete
	default:
		scenarioEnum = monitoringv2.ScenarioInsert
	}

	// Convert fields to string map
	stringFields := make(map[string]string)
	for k, v := range fields {
		if strVal, ok := v.(string); ok {
			stringFields[k] = strVal
		} else {
			// Convert non-string values
			stringFields[k] = fmt.Sprintf("%v", v)
		}
	}

	l.monitor.RecordValidation(
		operation,
		modelName,
		scenarioEnum,
		duration,
		err,
		stringFields,
	)
}

// Simple stub implementations for other types

type ExportConfig struct {
	Enabled bool
	Port    int
}

type MetricsExporter struct{}

func NewMetricsExporter(collector *MetricsCollector, logger *ValidationLogger, config ExportConfig) *MetricsExporter {
	return &MetricsExporter{}
}

func (e *MetricsExporter) Start() error { return nil }
func (e *MetricsExporter) Stop() error  { return nil }

// Helper function
func DefaultMonitoringConfig() *core.MonitoringConfig {
	config := core.DefaultMonitoringConfig()
	return &config
}
