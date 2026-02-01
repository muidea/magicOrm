package unified

import (
	"fmt"

	"github.com/muidea/magicOrm/monitoring/core"
	"github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/validation"
)

// MonitoringFactory creates and manages monitoring components
type MonitoringFactory struct {
	manager *MonitoringManager
}

// NewMonitoringFactory creates a new monitoring factory
func NewMonitoringFactory(config *core.MonitoringConfig) *MonitoringFactory {
	manager := NewMonitoringManager(config)
	return &MonitoringFactory{
		manager: manager,
	}
}

// GetManager returns the monitoring manager
func (f *MonitoringFactory) GetManager() *MonitoringManager {
	return f.manager
}

// GetCollector returns the core collector
func (f *MonitoringFactory) GetCollector() *core.Collector {
	return f.manager.GetCollector()
}

// GetExporter returns the exporter
func (f *MonitoringFactory) GetExporter() *core.Exporter {
	return f.manager.GetExporter()
}

// GetValidationMonitor returns the validation monitor
func (f *MonitoringFactory) GetValidationMonitor() *validation.ValidationMonitor {
	return f.manager.GetValidationMonitor()
}

// GetORMMonitor returns the ORM monitor
func (f *MonitoringFactory) GetORMMonitor() *orm.ORMMonitor {
	return f.manager.GetORMMonitor()
}

// Start starts all monitoring components
func (f *MonitoringFactory) Start() error {
	return f.manager.Start()
}

// Stop stops all monitoring components
func (f *MonitoringFactory) Stop() error {
	return f.manager.Stop()
}

// UpdateConfig updates the monitoring configuration
func (f *MonitoringFactory) UpdateConfig(config *core.MonitoringConfig) error {
	return f.manager.UpdateConfig(config)
}

// GetConfig returns the current configuration
func (f *MonitoringFactory) GetConfig() *core.MonitoringConfig {
	return f.manager.GetConfig()
}

// CreateMonitoredOrm creates a monitored ORM wrapper
func (f *MonitoringFactory) CreateMonitoredOrm(
	ormInterface orm.Orm,
	config *orm.MonitoringConfig,
) *orm.MonitoredOrm {

	ormMonitor := f.manager.GetORMMonitor()
	if ormMonitor == nil {
		// If ORM monitoring is disabled, return unwrapped ORM
		// or create a no-op monitor
		return orm.WrapOrmWithMonitoring(ormInterface, orm.DefaultORMMonitor(), config)
	}

	return orm.WrapOrmWithMonitoring(ormInterface, ormMonitor, config)
}

// CreateValidationMonitor creates a validation monitor
func (f *MonitoringFactory) CreateValidationMonitor() *validation.ValidationMonitor {
	return f.manager.GetValidationMonitor()
}

// CreateExporter creates and configures an exporter
func (f *MonitoringFactory) CreateExporter(config *core.ExportConfig) (*core.Exporter, error) {
	collector := f.manager.GetCollector()
	if collector == nil {
		collector = core.NewCollector(f.manager.GetConfig())
	}

	exporter := core.NewExporter(collector, config)
	return exporter, nil
}

// GetDefaultConfig returns the default monitoring configuration
func (f *MonitoringFactory) GetDefaultConfig() core.MonitoringConfig {
	return core.DefaultMonitoringConfig()
}

// GetDevelopmentConfig returns development monitoring configuration
func (f *MonitoringFactory) GetDevelopmentConfig() core.MonitoringConfig {
	return core.DevelopmentConfig()
}

// GetProductionConfig returns production monitoring configuration
func (f *MonitoringFactory) GetProductionConfig() core.MonitoringConfig {
	return core.ProductionConfig()
}

// GetHighLoadConfig returns high-load monitoring configuration
func (f *MonitoringFactory) GetHighLoadConfig() core.MonitoringConfig {
	return core.HighLoadConfig()
}

// GetStats returns comprehensive statistics
func (f *MonitoringFactory) GetStats() map[string]interface{} {
	return f.manager.GetManagerStats()
}

// Reset resets all monitoring components
func (f *MonitoringFactory) Reset() {
	f.manager.ResetStats()
}

// AddCustomLabels adds custom labels to all exported metrics
func (f *MonitoringFactory) AddCustomLabels(labels map[string]string) {
	f.manager.AddCustomLabels(labels)
}

// IsEnabled checks if monitoring is enabled
func (f *MonitoringFactory) IsEnabled() bool {
	return f.manager.IsEnabled()
}

// Enable enables monitoring
func (f *MonitoringFactory) Enable() {
	f.manager.Enable()
}

// Disable disables monitoring
func (f *MonitoringFactory) Disable() {
	f.manager.Disable()
}

// Convenience functions

// DefaultFactory creates a monitoring factory with default configuration
func DefaultFactory() *MonitoringFactory {
	config := core.DefaultMonitoringConfig()
	return NewMonitoringFactory(&config)
}

// DevelopmentFactory creates a monitoring factory for development
func DevelopmentFactory() *MonitoringFactory {
	config := core.DevelopmentConfig()
	return NewMonitoringFactory(&config)
}

// ProductionFactory creates a monitoring factory for production
func ProductionFactory() *MonitoringFactory {
	config := core.ProductionConfig()
	return NewMonitoringFactory(&config)
}

// HighLoadFactory creates a monitoring factory for high-load environments
func HighLoadFactory() *MonitoringFactory {
	config := core.HighLoadConfig()
	return NewMonitoringFactory(&config)
}

// FactoryWithConfig creates a monitoring factory with custom configuration
func FactoryWithConfig(config *core.MonitoringConfig) *MonitoringFactory {
	return NewMonitoringFactory(config)
}

// FactoryFromEnvironment creates a monitoring factory based on environment
func FactoryFromEnvironment(env string) *MonitoringFactory {
	switch env {
	case "development":
		return DevelopmentFactory()
	case "production":
		return ProductionFactory()
	case "highload":
		return HighLoadFactory()
	default:
		return DefaultFactory()
	}
}

// QuickStart starts monitoring with default configuration
func QuickStart() (*MonitoringFactory, error) {
	factory := DefaultFactory()
	if err := factory.Start(); err != nil {
		return nil, err
	}
	return factory, nil
}

// QuickStartWithConfig starts monitoring with custom configuration
func QuickStartWithConfig(config *core.MonitoringConfig) (*MonitoringFactory, error) {
	factory := NewMonitoringFactory(config)
	if err := factory.Start(); err != nil {
		return nil, err
	}
	return factory, nil
}

// Integration helpers

// IntegrateWithValidationSystem integrates monitoring with validation system
func (f *MonitoringFactory) IntegrateWithValidationSystem(validationSystem interface{}) error {
	// This would integrate the validation monitor with the actual validation system
	// Implementation depends on the validation system's API
	return nil
}

// IntegrateWithORMSystem integrates monitoring with ORM system
func (f *MonitoringFactory) IntegrateWithORMSystem(ormSystem interface{}) error {
	// This would integrate the ORM monitor with the actual ORM system
	// Implementation depends on the ORM system's API
	return nil
}

// GetMetricsEndpoint returns the metrics endpoint URL
func (f *MonitoringFactory) GetMetricsEndpoint() string {
	config := f.manager.GetConfig()
	if !config.IsExportEnabled() {
		return ""
	}

	protocol := "http"
	if config.ExportConfig.EnableTLS {
		protocol = "https"
	}

	return protocol + "://localhost:" + fmt.Sprintf("%d", config.ExportConfig.Port) + config.ExportConfig.Path
}

// GetHealthEndpoint returns the health endpoint URL
func (f *MonitoringFactory) GetHealthEndpoint() string {
	config := f.manager.GetConfig()
	if !config.IsExportEnabled() {
		return ""
	}

	protocol := "http"
	if config.ExportConfig.EnableTLS {
		protocol = "https"
	}

	return protocol + "://localhost:" + fmt.Sprintf("%d", config.ExportConfig.Port) + config.ExportConfig.HealthCheckPath
}

// GetInfoEndpoint returns the info endpoint URL
func (f *MonitoringFactory) GetInfoEndpoint() string {
	config := f.manager.GetConfig()
	if !config.IsExportEnabled() {
		return ""
	}

	protocol := "http"
	if config.ExportConfig.EnableTLS {
		protocol = "https"
	}

	return protocol + "://localhost:" + fmt.Sprintf("%d", config.ExportConfig.Port) + config.ExportConfig.InfoPath
}
