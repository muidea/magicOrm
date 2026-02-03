// Package core provides simplified monitoring configuration.
// This configuration is only for local metric collection, not for export.
package core

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/core"
)

// DetailLevel defines the level of monitoring detail
type DetailLevel = core.DetailLevel

const (
	// DetailLevelBasic collects only essential metrics
	DetailLevelBasic DetailLevel = core.DetailLevelBasic
	// DetailLevelStandard collects standard operational metrics
	DetailLevelStandard DetailLevel = core.DetailLevelStandard
	// DetailLevelDetailed collects comprehensive metrics including performance breakdowns
	DetailLevelDetailed DetailLevel = core.DetailLevelDetailed
)

// MonitoringConfig holds simplified configuration for metric collection only.
type MonitoringConfig struct {
	// Enabled controls whether monitoring is active
	Enabled bool `json:"enabled"`

	// Namespace for metric names (e.g., "magicorm")
	Namespace string `json:"namespace"`

	// SamplingRate controls the rate of metric collection (0.0-1.0)
	SamplingRate float64 `json:"sampling_rate"`

	// AsyncCollection enables asynchronous metric collection
	AsyncCollection bool `json:"async_collection"`

	// CollectionInterval controls how often metrics are collected asynchronously
	CollectionInterval time.Duration `json:"collection_interval"`

	// RetentionPeriod controls how long metrics are retained locally
	RetentionPeriod time.Duration `json:"retention_period"`

	// Detail level for metrics collection
	DetailLevel DetailLevel `json:"detail_level"`

	// Environment label for metrics
	Environment string `json:"environment"`

	// Performance optimization
	BatchSize          int           `json:"batch_size"`
	BufferSize         int           `json:"buffer_size"`
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	Timeout            time.Duration `json:"timeout"`
}

// DefaultMonitoringConfig returns default monitoring configuration for collection only.
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:            true,
		Namespace:          "magicorm",
		SamplingRate:       1.0,
		AsyncCollection:    true,
		CollectionInterval: 60 * time.Second,
		RetentionPeriod:    24 * time.Hour,
		DetailLevel:        DetailLevelStandard,
		Environment:        "development",
		BatchSize:          100,
		BufferSize:         1000,
		MaxConcurrentTasks: 10,
		Timeout:            30 * time.Second,
	}
}

// DevelopmentConfig returns configuration optimized for development environment.
func DevelopmentConfig() MonitoringConfig {
	config := DefaultMonitoringConfig()
	config.Environment = "development"
	config.SamplingRate = 0.1 // 10% sampling in development
	config.DetailLevel = DetailLevelDetailed
	config.RetentionPeriod = 1 * time.Hour
	return config
}

// ProductionConfig returns configuration optimized for production environment.
func ProductionConfig() MonitoringConfig {
	config := DefaultMonitoringConfig()
	config.Environment = "production"
	config.SamplingRate = 0.5 // 50% sampling in production
	config.DetailLevel = DetailLevelStandard
	config.RetentionPeriod = 7 * 24 * time.Hour // 7 days
	return config
}

// HighLoadConfig returns configuration optimized for high-load environments.
func HighLoadConfig() MonitoringConfig {
	config := ProductionConfig()
	config.SamplingRate = 0.1 // 10% sampling under high load
	config.DetailLevel = DetailLevelBasic
	config.BatchSize = 50
	config.BufferSize = 500
	return config
}

// Validate validates the monitoring configuration.
func (c *MonitoringConfig) Validate() error {
	if c.SamplingRate < 0 || c.SamplingRate > 1 {
		return &ConfigError{Field: "SamplingRate", Message: "must be between 0.0 and 1.0"}
	}

	if c.CollectionInterval < 1*time.Second {
		return &ConfigError{Field: "CollectionInterval", Message: "must be at least 1 second"}
	}

	if c.RetentionPeriod < 1*time.Minute {
		return &ConfigError{Field: "RetentionPeriod", Message: "must be at least 1 minute"}
	}

	if c.BatchSize <= 0 {
		return &ConfigError{Field: "BatchSize", Message: "must be positive"}
	}

	if c.BufferSize <= 0 {
		return &ConfigError{Field: "BufferSize", Message: "must be positive"}
	}

	if c.MaxConcurrentTasks <= 0 {
		return &ConfigError{Field: "MaxConcurrentTasks", Message: "must be positive"}
	}

	return nil
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + ": " + e.Message
}

// IsEnabled checks if monitoring is enabled.
func (c *MonitoringConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

// GetNamespace returns the namespace for metrics.
func (c *MonitoringConfig) GetNamespace() string {
	if c == nil || c.Namespace == "" {
		return "magicorm"
	}
	return c.Namespace
}

// GetEnvironment returns the environment label.
func (c *MonitoringConfig) GetEnvironment() string {
	if c == nil || c.Environment == "" {
		return "development"
	}
	return c.Environment
}
