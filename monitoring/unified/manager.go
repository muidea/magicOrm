package unified

import (
	"sync"
	"time"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/monitoring/core"
	monitoringDatabase "github.com/muidea/magicOrm/monitoring/database"
	"github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/validation"
)

// MonitoringManager provides unified management of all monitoring components
type MonitoringManager struct {
	mu sync.RWMutex

	config *core.MonitoringConfig

	// Core components
	collector *core.Collector
	exporter  *core.Exporter

	// Subsystem monitors
	validationMonitor *validation.ValidationMonitor
	ormMonitor        *orm.ORMMonitor
	databaseMonitor   *monitoringDatabase.DatabaseMonitor

	// State
	startTime time.Time
	enabled   bool
	stats     ManagerStats
}

// ManagerStats holds manager statistics
type ManagerStats struct {
	Uptime            time.Duration `json:"uptime"`
	MetricsCollected  int64         `json:"metrics_collected"`
	MetricsExported   int64         `json:"metrics_exported"`
	ValidationMetrics int64         `json:"validation_metrics"`
	ORMMetrics        int64         `json:"orm_metrics"`
	DatabaseMetrics   int64         `json:"database_metrics"`
	LastActivity      time.Time     `json:"last_activity"`
	Errors            int64         `json:"errors"`
}

// NewMonitoringManager creates a new unified monitoring manager
func NewMonitoringManager(config *core.MonitoringConfig) *MonitoringManager {
	if config == nil {
		defaultConfig := core.DefaultMonitoringConfig()
		config = &defaultConfig
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		// Use default config if validation fails
		defaultConfig := core.DefaultMonitoringConfig()
		config = &defaultConfig
	}

	manager := &MonitoringManager{
		config:    config,
		startTime: time.Now(),
		enabled:   config.Enabled,
		stats: ManagerStats{
			LastActivity: time.Now(),
		},
	}

	// Initialize components based on configuration
	manager.initializeComponents()

	return manager
}

// Start starts all monitoring components
func (m *MonitoringManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return nil
	}

	// Start exporter if enabled
	if m.config.IsExportEnabled() && m.exporter != nil {
		if err := m.exporter.Start(); err != nil {
			return err
		}
	}

	m.stats.LastActivity = time.Now()
	return nil
}

// Stop stops all monitoring components
func (m *MonitoringManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop exporter if running
	if m.exporter != nil {
		if err := m.exporter.Stop(); err != nil {
			return err
		}
	}

	m.enabled = false
	m.stats.LastActivity = time.Now()

	return nil
}

// Enable enables monitoring
func (m *MonitoringManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enabled = true
	m.config.Enabled = true
	m.stats.LastActivity = time.Now()
}

// Disable disables monitoring
func (m *MonitoringManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enabled = false
	m.config.Enabled = false
	m.stats.LastActivity = time.Now()
}

// IsEnabled returns whether monitoring is enabled
func (m *MonitoringManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// UpdateConfig updates the monitoring configuration
func (m *MonitoringManager) UpdateConfig(config *core.MonitoringConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate new configuration
	if err := config.Validate(); err != nil {
		return err
	}

	// Stop existing components if configuration changes significantly
	if m.config.Enabled != config.Enabled ||
		m.config.IsExportEnabled() != config.IsExportEnabled() {

		if m.exporter != nil {
			m.exporter.Stop()
		}
	}

	// Update configuration
	m.config = config
	m.enabled = config.Enabled

	// Reinitialize components if needed
	m.initializeComponents()

	// Restart if enabled
	if m.enabled && config.IsExportEnabled() && m.exporter != nil {
		if err := m.exporter.Start(); err != nil {
			return err
		}
	}

	m.stats.LastActivity = time.Now()
	return nil
}

// GetConfig returns the current configuration
func (m *MonitoringManager) GetConfig() *core.MonitoringConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config := *m.config
	return &config
}

// GetCollector returns the core collector
func (m *MonitoringManager) GetCollector() *core.Collector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.collector
}

// GetExporter returns the exporter
func (m *MonitoringManager) GetExporter() *core.Exporter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.exporter
}

// GetValidationMonitor returns the validation monitor
func (m *MonitoringManager) GetValidationMonitor() *validation.ValidationMonitor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.validationMonitor
}

// GetORMMonitor returns the ORM monitor
func (m *MonitoringManager) GetORMMonitor() *orm.ORMMonitor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ormMonitor
}

// GetStats returns manager statistics
func (m *MonitoringManager) GetStats() ManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.stats
	stats.Uptime = time.Since(m.startTime)

	// Update metrics counts from collector
	if m.collector != nil {
		collectorStats := m.collector.GetStats()
		stats.MetricsCollected = collectorStats.MetricsCollected
	}

	// Update export stats from exporter
	if m.exporter != nil {
		exporterStats := m.exporter.GetStats()
		stats.MetricsExported = exporterStats.RequestsTotal
	}

	return stats
}

// RecordActivity records manager activity
func (m *MonitoringManager) RecordActivity() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.LastActivity = time.Now()
}

// RecordError records an error in monitoring system
func (m *MonitoringManager) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.Errors++
	m.stats.LastActivity = time.Now()
}

// ResetStats resets all statistics
func (m *MonitoringManager) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = ManagerStats{
		LastActivity: time.Now(),
	}

	// Reset component stats if they support it
	if m.collector != nil {
		m.collector.Reset()
	}

	if m.exporter != nil {
		m.exporter.ResetStats()
	}
}

// AddCustomLabel adds a custom label to all exported metrics
func (m *MonitoringManager) AddCustomLabel(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.exporter != nil {
		m.exporter.WithLabel(key, value)
	}
}

// AddCustomLabels adds multiple custom labels to all exported metrics
func (m *MonitoringManager) AddCustomLabels(labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.exporter != nil {
		m.exporter.WithLabels(labels)
	}
}

// GetMetrics returns all collected metrics
func (m *MonitoringManager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.collector == nil {
		return make(map[string]interface{})
	}

	// Convert metrics to interface{} map
	metrics := m.collector.GetMetrics()
	result := make(map[string]interface{})
	for k, v := range metrics {
		result[k] = v
	}
	return result
}

// Cleanup performs cleanup of old metrics
func (m *MonitoringManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.collector != nil {
		m.collector.Cleanup()
	}

	m.stats.LastActivity = time.Now()
}

// Private methods

func (m *MonitoringManager) initializeComponents() {
	// Create or update core collector
	if m.collector == nil {
		m.collector = core.NewCollector(m.config)
	}

	// Create or update exporter if export is enabled
	if m.config.IsExportEnabled() && m.exporter == nil {
		exportConfig := m.config.ExportConfig
		m.exporter = core.NewExporter(m.collector, &exportConfig)
	} else if !m.config.IsExportEnabled() && m.exporter != nil {
		// Stop and clear exporter if export is disabled
		m.exporter.Stop()
		m.exporter = nil
	}

	// Create or update validation monitor if enabled
	if m.config.IsValidationEnabled() && m.validationMonitor == nil {
		m.validationMonitor = validation.NewValidationMonitor(m.collector, m.config)
	} else if !m.config.IsValidationEnabled() && m.validationMonitor != nil {
		m.validationMonitor = nil
	}

	// Create or update ORM monitor if enabled
	if m.config.IsORMEnabled() && m.ormMonitor == nil {
		m.ormMonitor = orm.NewORMMonitor(m.collector, m.config)
	} else if !m.config.IsORMEnabled() && m.ormMonitor != nil {
		m.ormMonitor = nil
	}

	// Create or update database monitor if enabled
	if m.config.IsDatabaseEnabled() && m.databaseMonitor == nil {
		m.databaseMonitor = monitoringDatabase.NewDatabaseMonitor(m.collector, m.config)
	} else if !m.config.IsDatabaseEnabled() && m.databaseMonitor != nil {
		m.databaseMonitor = nil
	}

	// Update stats
	m.stats.LastActivity = time.Now()
}

// Convenience functions

// DefaultMonitoringManager creates a monitoring manager with default configuration
func DefaultMonitoringManager() *MonitoringManager {
	config := core.DefaultMonitoringConfig()
	return NewMonitoringManager(&config)
}

// DevelopmentMonitoringManager creates a monitoring manager for development
func DevelopmentMonitoringManager() *MonitoringManager {
	config := core.DevelopmentConfig()
	return NewMonitoringManager(&config)
}

// ProductionMonitoringManager creates a monitoring manager for production
func ProductionMonitoringManager() *MonitoringManager {
	config := core.ProductionConfig()
	return NewMonitoringManager(&config)
}

// HighLoadMonitoringManager creates a monitoring manager for high-load environments
func HighLoadMonitoringManager() *MonitoringManager {
	config := core.HighLoadConfig()
	return NewMonitoringManager(&config)
}

// GetManagerStats returns comprehensive statistics from all components
func (m *MonitoringManager) GetManagerStats() map[string]interface{} {
	stats := m.GetStats()
	config := m.GetConfig()

	result := map[string]interface{}{
		"manager": stats,
		"config": map[string]interface{}{
			"enabled":           config.Enabled,
			"sampling_rate":     config.SamplingRate,
			"enable_orm":        config.EnableORM,
			"enable_validation": config.EnableValidation,
			"enable_cache":      config.EnableCache,
			"enable_database":   config.EnableDatabase,
			"export_enabled":    config.IsExportEnabled(),
			"detail_level":      config.DetailLevel,
		},
	}

	// Add collector stats if available
	if m.collector != nil {
		collectorStats := m.collector.GetStats()
		result["collector"] = collectorStats
	}

	// Add exporter stats if available
	if m.exporter != nil {
		exporterStats := m.exporter.GetStats()
		result["exporter"] = exporterStats
	}

	// Add validation monitor stats if available
	if m.validationMonitor != nil {
		validationStats := m.validationMonitor.GetStats()
		result["validation"] = validationStats
	}

	// Add ORM monitor stats if available
	if m.ormMonitor != nil {
		ormStats := m.ormMonitor.GetStats()
		result["orm"] = ormStats
	}

	// Add database monitor stats if available
	if m.databaseMonitor != nil {
		collector := m.databaseMonitor.GetCollector()
		if collector != nil {
			databaseStats := collector.GetStats()
			result["database"] = databaseStats
		}
	}

	return result
}

// GetDatabaseMonitor returns the database monitor
func (m *MonitoringManager) GetDatabaseMonitor() *monitoringDatabase.DatabaseMonitor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.databaseMonitor
}

// SetDatabaseMonitor sets the database monitor
func (m *MonitoringManager) SetDatabaseMonitor(monitor *monitoringDatabase.DatabaseMonitor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.databaseMonitor = monitor
	m.stats.LastActivity = time.Now()
}

// GetDatabaseMonitoringFactory returns a database monitoring factory
func (m *MonitoringManager) GetDatabaseMonitoringFactory() *monitoringDatabase.DatabaseMonitoringFactory {
	return monitoringDatabase.NewDatabaseMonitoringFactory(m.collector, m.config)
}

// WrapDatabaseExecutor wraps a database executor with monitoring
func (m *MonitoringManager) WrapDatabaseExecutor(executor database.Executor, dbType string) *monitoringDatabase.MonitoredExecutor {
	factory := m.GetDatabaseMonitoringFactory()
	return factory.WrapExecutor(executor, dbType)
}

// WrapDatabasePool wraps a database pool with monitoring
func (m *MonitoringManager) WrapDatabasePool(pool database.Pool, dbType string) *monitoringDatabase.MonitoredPool {
	factory := m.GetDatabaseMonitoringFactory()
	return factory.WrapPool(pool, dbType)
}
