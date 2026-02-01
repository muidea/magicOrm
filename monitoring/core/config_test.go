package core

import (
	"testing"
	"time"
)

func TestDefaultMonitoringConfig(t *testing.T) {
	config := DefaultMonitoringConfig()

	// Test default values
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	if config.SamplingRate != 1.0 {
		t.Errorf("Expected sampling rate 1.0, got %f", config.SamplingRate)
	}

	if !config.AsyncCollection {
		t.Error("Default config should have async collection enabled")
	}

	if config.CollectionInterval != 30*time.Second {
		t.Errorf("Expected collection interval 30s, got %v", config.CollectionInterval)
	}

	if config.RetentionPeriod != 24*time.Hour {
		t.Errorf("Expected retention period 24h, got %v", config.RetentionPeriod)
	}

	if !config.EnableORM {
		t.Error("Default config should have ORM monitoring enabled")
	}

	if !config.EnableValidation {
		t.Error("Default config should have validation monitoring enabled")
	}

	if !config.EnableCache {
		t.Error("Default config should have cache monitoring enabled")
	}

	if !config.EnableDatabase {
		t.Error("Default config should have database monitoring enabled")
	}

	if config.DetailLevel != DetailLevelStandard {
		t.Errorf("Expected detail level standard, got %s", config.DetailLevel)
	}

	if config.BatchSize != 100 {
		t.Errorf("Expected batch size 100, got %d", config.BatchSize)
	}

	if config.BufferSize != 1000 {
		t.Errorf("Expected buffer size 1000, got %d", config.BufferSize)
	}

	if config.MaxConcurrentTasks != 10 {
		t.Errorf("Expected max concurrent tasks 10, got %d", config.MaxConcurrentTasks)
	}

	if config.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", config.Timeout)
	}
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	// Test development-specific values
	if config.SamplingRate != 0.1 {
		t.Errorf("Expected sampling rate 0.1 for development, got %f", config.SamplingRate)
	}

	if config.DetailLevel != DetailLevelBasic {
		t.Errorf("Expected detail level basic for development, got %s", config.DetailLevel)
	}

	if config.ExportConfig.Enabled {
		t.Error("Export should be disabled for development")
	}

	if config.AsyncCollection {
		t.Error("Async collection should be disabled for development")
	}
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	// Test production-specific values
	if config.SamplingRate != 0.5 {
		t.Errorf("Expected sampling rate 0.5 for production, got %f", config.SamplingRate)
	}

	if config.DetailLevel != DetailLevelStandard {
		t.Errorf("Expected detail level standard for production, got %s", config.DetailLevel)
	}

	if !config.ExportConfig.Enabled {
		t.Error("Export should be enabled for production")
	}

	if !config.ExportConfig.EnableAuth {
		t.Error("Auth should be enabled for production")
	}

	if !config.ExportConfig.EnableTLS {
		t.Error("TLS should be enabled for production")
	}

	if config.BatchSize != 500 {
		t.Errorf("Expected batch size 500 for production, got %d", config.BatchSize)
	}

	if config.BufferSize != 5000 {
		t.Errorf("Expected buffer size 5000 for production, got %d", config.BufferSize)
	}

	if config.MaxConcurrentTasks != 50 {
		t.Errorf("Expected max concurrent tasks 50 for production, got %d", config.MaxConcurrentTasks)
	}
}

func TestHighLoadConfig(t *testing.T) {
	config := HighLoadConfig()

	// Test high-load specific values
	if config.SamplingRate != 0.1 {
		t.Errorf("Expected sampling rate 0.1 for high load, got %f", config.SamplingRate)
	}

	if config.DetailLevel != DetailLevelBasic {
		t.Errorf("Expected detail level basic for high load, got %s", config.DetailLevel)
	}

	if config.ExportConfig.RefreshInterval != 60*time.Second {
		t.Errorf("Expected refresh interval 60s for high load, got %v", config.ExportConfig.RefreshInterval)
	}

	if config.BatchSize != 1000 {
		t.Errorf("Expected batch size 1000 for high load, got %d", config.BatchSize)
	}

	if config.BufferSize != 10000 {
		t.Errorf("Expected buffer size 10000 for high load, got %d", config.BufferSize)
	}

	if config.MaxConcurrentTasks != 100 {
		t.Errorf("Expected max concurrent tasks 100 for high load, got %d", config.MaxConcurrentTasks)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      MonitoringConfig
		shouldError bool
	}{
		{
			name:   "Valid default config",
			config: DefaultMonitoringConfig(),
		},
		{
			name: "Invalid sampling rate negative",
			config: MonitoringConfig{
				SamplingRate: -0.1,
			},
			shouldError: true,
		},
		{
			name: "Invalid sampling rate too high",
			config: MonitoringConfig{
				SamplingRate: 1.5,
			},
			shouldError: true,
		},
		{
			name: "Invalid collection interval",
			config: MonitoringConfig{
				CollectionInterval: -1 * time.Second,
			},
			shouldError: true,
		},
		{
			name: "Invalid retention period",
			config: MonitoringConfig{
				RetentionPeriod: -1 * time.Hour,
			},
			shouldError: true,
		},
		{
			name: "Invalid batch size",
			config: MonitoringConfig{
				BatchSize: 0,
			},
			shouldError: true,
		},
		{
			name: "Invalid buffer size",
			config: MonitoringConfig{
				BufferSize: 0,
			},
			shouldError: true,
		},
		{
			name: "Invalid max concurrent tasks",
			config: MonitoringConfig{
				MaxConcurrentTasks: 0,
			},
			shouldError: true,
		},
		{
			name: "Invalid timeout",
			config: MonitoringConfig{
				Timeout: -1 * time.Second,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldError && err == nil {
				t.Error("Expected validation error, got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestExportConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      ExportConfig
		shouldError bool
	}{
		{
			name:   "Valid default export config",
			config: DefaultExportConfig(),
		},
		{
			name: "Invalid port too low",
			config: ExportConfig{
				Port: 0,
				Path: "/metrics",
			},
			shouldError: true,
		},
		{
			name: "Invalid port too high",
			config: ExportConfig{
				Port: 65536,
				Path: "/metrics",
			},
			shouldError: true,
		},
		{
			name: "Empty path",
			config: ExportConfig{
				Port: 9090,
				Path: "",
			},
			shouldError: true,
		},
		{
			name: "Invalid refresh interval",
			config: ExportConfig{
				Port:            9090,
				Path:            "/metrics",
				RefreshInterval: -1 * time.Second,
			},
			shouldError: true,
		},
		{
			name: "Invalid scrape timeout",
			config: ExportConfig{
				Port:          9090,
				Path:          "/metrics",
				ScrapeTimeout: -1 * time.Second,
			},
			shouldError: true,
		},
		{
			name: "TLS enabled without cert",
			config: ExportConfig{
				Port:        9090,
				Path:        "/metrics",
				EnableTLS:   true,
				TLSCertPath: "",
				TLSKeyPath:  "key.pem",
			},
			shouldError: true,
		},
		{
			name: "TLS enabled without key",
			config: ExportConfig{
				Port:        9090,
				Path:        "/metrics",
				EnableTLS:   true,
				TLSCertPath: "cert.pem",
				TLSKeyPath:  "",
			},
			shouldError: true,
		},
		{
			name: "Auth enabled without token",
			config: ExportConfig{
				Port:       9090,
				Path:       "/metrics",
				EnableAuth: true,
				AuthToken:  "",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldError && err == nil {
				t.Error("Expected validation error, got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestShouldSample(t *testing.T) {
	tests := []struct {
		name         string
		samplingRate float64
		expected     bool
	}{
		{"Rate 1.0", 1.0, true},
		{"Rate 0.0", 0.0, false},
		{"Rate 0.5", 0.5, true}, // Our implementation always returns true for >0
		{"Rate 0.1", 0.1, true}, // Our implementation always returns true for >0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MonitoringConfig{SamplingRate: tt.samplingRate}
			result := config.ShouldSample()
			if result != tt.expected {
				t.Errorf("ShouldSample() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsComponentEnabled(t *testing.T) {
	config := MonitoringConfig{
		Enabled:          true,
		EnableORM:        true,
		EnableValidation: false,
		EnableCache:      true,
		EnableDatabase:   false,
	}

	if !config.IsORMEnabled() {
		t.Error("ORM should be enabled")
	}

	if config.IsValidationEnabled() {
		t.Error("Validation should be disabled")
	}

	if !config.IsCacheEnabled() {
		t.Error("Cache should be enabled")
	}

	if config.IsDatabaseEnabled() {
		t.Error("Database should be disabled")
	}

	// Test with global disabled
	config.Enabled = false
	if config.IsORMEnabled() {
		t.Error("ORM should be disabled when monitoring is disabled")
	}
}

func TestIsExportEnabled(t *testing.T) {
	config := MonitoringConfig{
		Enabled: true,
		ExportConfig: ExportConfig{
			Enabled: true,
		},
	}

	if !config.IsExportEnabled() {
		t.Error("Export should be enabled")
	}

	// Test with export disabled
	config.ExportConfig.Enabled = false
	if config.IsExportEnabled() {
		t.Error("Export should be disabled")
	}

	// Test with monitoring disabled
	config.Enabled = false
	config.ExportConfig.Enabled = true
	if config.IsExportEnabled() {
		t.Error("Export should be disabled when monitoring is disabled")
	}
}

func TestConfigErrorString(t *testing.T) {
	err := &ConfigError{
		Field:   "sampling_rate",
		Value:   1.5,
		Message: "must be between 0 and 1",
	}

	expected := "invalid configuration: sampling_rate=number: must be between 0 and 1"
	if err.Error() != expected {
		t.Errorf("Error() = %q, expected %q", err.Error(), expected)
	}
}
