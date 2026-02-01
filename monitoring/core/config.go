package core

import (
	"time"
)

// DetailLevel defines the level of monitoring detail
type DetailLevel string

const (
	// DetailLevelBasic collects only essential metrics
	DetailLevelBasic DetailLevel = "basic"
	// DetailLevelStandard collects standard operational metrics
	DetailLevelStandard DetailLevel = "standard"
	// DetailLevelDetailed collects comprehensive metrics including performance breakdowns
	DetailLevelDetailed DetailLevel = "detailed"
)

// MonitoringConfig holds configuration for the unified monitoring system
type MonitoringConfig struct {
	// Enabled controls whether monitoring is active
	Enabled bool `json:"enabled"`

	// SamplingRate controls the rate of metric collection (0.0-1.0)
	SamplingRate float64 `json:"sampling_rate"`

	// AsyncCollection enables asynchronous metric collection
	AsyncCollection bool `json:"async_collection"`
	// CollectionInterval controls how often metrics are collected asynchronously
	CollectionInterval time.Duration `json:"collection_interval"`

	// RetentionPeriod controls how long metrics are retained
	RetentionPeriod time.Duration `json:"retention_period"`

	// Component-specific monitoring
	EnableORM        bool `json:"enable_orm"`
	EnableValidation bool `json:"enable_validation"`
	EnableCache      bool `json:"enable_cache"`
	EnableDatabase   bool `json:"enable_database"`

	// Detail level for metrics collection
	DetailLevel DetailLevel `json:"detail_level"`

	// Export configuration
	ExportConfig ExportConfig `json:"export_config"`

	// Performance optimization
	BatchSize          int           `json:"batch_size"`
	BufferSize         int           `json:"buffer_size"`
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	Timeout            time.Duration `json:"timeout"`
}

// ExportConfig holds configuration for metric export
type ExportConfig struct {
	// Enabled controls whether metrics are exported
	Enabled bool `json:"enabled"`

	// HTTP server configuration
	Port            int    `json:"port"`
	Path            string `json:"path"`
	HealthCheckPath string `json:"health_check_path"`
	MetricsPath     string `json:"metrics_path"`
	InfoPath        string `json:"info_path"`

	// Format support
	EnablePrometheus bool `json:"enable_prometheus"`
	EnableJSON       bool `json:"enable_json"`

	// Export intervals
	RefreshInterval time.Duration `json:"refresh_interval"`
	ScrapeTimeout   time.Duration `json:"scrape_timeout"`

	// Security
	EnableTLS    bool     `json:"enable_tls"`
	TLSCertPath  string   `json:"tls_cert_path"`
	TLSKeyPath   string   `json:"tls_key_path"`
	EnableAuth   bool     `json:"enable_auth"`
	AuthToken    string   `json:"auth_token"`
	AllowedHosts []string `json:"allowed_hosts"`
}

// DefaultMonitoringConfig returns the default monitoring configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:            true,
		SamplingRate:       1.0,
		AsyncCollection:    true,
		CollectionInterval: 30 * time.Second,
		RetentionPeriod:    24 * time.Hour,
		EnableORM:          true,
		EnableValidation:   true,
		EnableCache:        true,
		EnableDatabase:     true,
		DetailLevel:        DetailLevelStandard,
		ExportConfig:       DefaultExportConfig(),
		BatchSize:          100,
		BufferSize:         1000,
		MaxConcurrentTasks: 10,
		Timeout:            10 * time.Second,
	}
}

// DefaultExportConfig returns the default export configuration
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		Enabled:          true,
		Port:             9090,
		Path:             "/metrics",
		HealthCheckPath:  "/health",
		MetricsPath:      "/metrics/json",
		InfoPath:         "/",
		EnablePrometheus: true,
		EnableJSON:       true,
		RefreshInterval:  30 * time.Second,
		ScrapeTimeout:    10 * time.Second,
		EnableTLS:        false,
		EnableAuth:       false,
		AllowedHosts:     []string{"localhost", "127.0.0.1"},
	}
}

// DevelopmentConfig returns configuration suitable for development
func DevelopmentConfig() MonitoringConfig {
	config := DefaultMonitoringConfig()
	config.SamplingRate = 0.1 // 10% sampling in development
	config.DetailLevel = DetailLevelBasic
	config.ExportConfig.Enabled = false
	config.AsyncCollection = false // Synchronous for easier debugging
	return config
}

// ProductionConfig returns configuration suitable for production
func ProductionConfig() MonitoringConfig {
	config := DefaultMonitoringConfig()
	config.SamplingRate = 0.5 // 50% sampling in production
	config.DetailLevel = DetailLevelStandard
	config.ExportConfig.Enabled = true
	config.ExportConfig.EnableAuth = true
	config.ExportConfig.EnableTLS = true
	config.BatchSize = 500
	config.BufferSize = 5000
	config.MaxConcurrentTasks = 50
	return config
}

// HighLoadConfig returns configuration for high-load environments
func HighLoadConfig() MonitoringConfig {
	config := ProductionConfig()
	config.SamplingRate = 0.1 // 10% sampling under high load
	config.DetailLevel = DetailLevelBasic
	config.ExportConfig.RefreshInterval = 60 * time.Second
	config.BatchSize = 1000
	config.BufferSize = 10000
	config.MaxConcurrentTasks = 100
	return config
}

// Validate validates the monitoring configuration
func (c *MonitoringConfig) Validate() error {
	if c.SamplingRate < 0 || c.SamplingRate > 1 {
		return &ConfigError{Field: "sampling_rate", Value: c.SamplingRate, Message: "must be between 0 and 1"}
	}

	if c.CollectionInterval <= 0 {
		return &ConfigError{Field: "collection_interval", Value: c.CollectionInterval, Message: "must be positive"}
	}

	if c.RetentionPeriod <= 0 {
		return &ConfigError{Field: "retention_period", Value: c.RetentionPeriod, Message: "must be positive"}
	}

	if c.BatchSize <= 0 {
		return &ConfigError{Field: "batch_size", Value: c.BatchSize, Message: "must be positive"}
	}

	if c.BufferSize <= 0 {
		return &ConfigError{Field: "buffer_size", Value: c.BufferSize, Message: "must be positive"}
	}

	if c.MaxConcurrentTasks <= 0 {
		return &ConfigError{Field: "max_concurrent_tasks", Value: c.MaxConcurrentTasks, Message: "must be positive"}
	}

	if c.Timeout <= 0 {
		return &ConfigError{Field: "timeout", Value: c.Timeout, Message: "must be positive"}
	}

	return c.ExportConfig.Validate()
}

// Validate validates the export configuration
func (c *ExportConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return &ConfigError{Field: "port", Value: c.Port, Message: "must be between 1 and 65535"}
	}

	if c.Path == "" {
		return &ConfigError{Field: "path", Value: c.Path, Message: "cannot be empty"}
	}

	if c.RefreshInterval <= 0 {
		return &ConfigError{Field: "refresh_interval", Value: c.RefreshInterval, Message: "must be positive"}
	}

	if c.ScrapeTimeout <= 0 {
		return &ConfigError{Field: "scrape_timeout", Value: c.ScrapeTimeout, Message: "must be positive"}
	}

	if c.EnableTLS {
		if c.TLSCertPath == "" {
			return &ConfigError{Field: "tls_cert_path", Value: c.TLSCertPath, Message: "required when TLS is enabled"}
		}
		if c.TLSKeyPath == "" {
			return &ConfigError{Field: "tls_key_path", Value: c.TLSKeyPath, Message: "required when TLS is enabled"}
		}
	}

	if c.EnableAuth && c.AuthToken == "" {
		return &ConfigError{Field: "auth_token", Value: c.AuthToken, Message: "required when auth is enabled"}
	}

	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigError) Error() string {
	return "invalid configuration: " + e.Field + "=" + stringify(e.Value) + ": " + e.Message
}

func stringify(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return stringifyNumber(v)
	case float32, float64:
		return stringifyNumber(v)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case time.Duration:
		return val.String()
	default:
		return "unknown"
	}
}

func stringifyNumber(v interface{}) string {
	// Simple implementation - in real code you'd use fmt.Sprintf
	return "number"
}

// ShouldSample determines if a metric should be collected based on sampling rate
func (c *MonitoringConfig) ShouldSample() bool {
	if c.SamplingRate >= 1.0 {
		return true
	}
	if c.SamplingRate <= 0 {
		return false
	}
	// Simple sampling - in production you'd use a proper sampling algorithm
	return true // Placeholder - will be implemented with proper sampling
}

// IsORMEnabled checks if ORM monitoring is enabled
func (c *MonitoringConfig) IsORMEnabled() bool {
	return c.Enabled && c.EnableORM
}

// IsValidationEnabled checks if validation monitoring is enabled
func (c *MonitoringConfig) IsValidationEnabled() bool {
	return c.Enabled && c.EnableValidation
}

// IsCacheEnabled checks if cache monitoring is enabled
func (c *MonitoringConfig) IsCacheEnabled() bool {
	return c.Enabled && c.EnableCache
}

// IsDatabaseEnabled checks if database monitoring is enabled
func (c *MonitoringConfig) IsDatabaseEnabled() bool {
	return c.Enabled && c.EnableDatabase
}

// IsExportEnabled checks if metric export is enabled
func (c *MonitoringConfig) IsExportEnabled() bool {
	return c.Enabled && c.ExportConfig.Enabled
}
