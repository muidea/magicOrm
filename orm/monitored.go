package orm

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// MonitoredOrmConfig holds configuration for monitored ORM
type MonitoredOrmConfig struct {
	// Enable monitoring (deprecated - MagicORM no longer collects data)
	Enabled bool

	// Custom labels for metrics (deprecated - MagicORM no longer collects data)
	CustomLabels map[string]string
}

// DefaultMonitoredOrmConfig returns default monitored ORM configuration
func DefaultMonitoredOrmConfig() MonitoredOrmConfig {
	return MonitoredOrmConfig{
		Enabled:      false, // MagicORM no longer collects data
		CustomLabels: make(map[string]string),
	}
}

// NewMonitoredOrm creates a new ORM instance with monitoring
func NewMonitoredOrm(
	provider provider.Provider,
	cfg database.Config,
	prefix string,
	monitoredConfig MonitoredOrmConfig,
) (Orm, *cd.Error) {

	// Create base ORM
	baseOrm, err := NewOrm(provider, cfg, prefix)
	if err != nil {
		return nil, err
	}

	// Wrap with monitoring if enabled
	if !monitoredConfig.Enabled {
		return baseOrm, nil
	}

	return wrapOrmWithMonitoring(baseOrm, monitoredConfig), nil
}

// wrapOrmWithMonitoring wraps an ORM with monitoring capabilities
func wrapOrmWithMonitoring(orm Orm, config MonitoredOrmConfig) Orm {
	// MagicORM no longer collects data - return the original ORM
	// Monitoring is now handled by magicCommon/monitoring system
	return orm
}

// Helper functions for monitoring integration
// Note: MagicORM no longer collects data - these are now no-op functions
// Data collection is handled by magicCommon/monitoring system

// RecordOperation is a helper to record ORM operations for monitoring
func RecordOperation(
	operation string,
	modelName string,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	// MagicORM no longer collects data - this is a no-op
	// Data collection is handled by magicCommon/monitoring system
}

// RecordQuery is a helper to record ORM queries for monitoring
func RecordQuery(
	modelName string,
	queryType string,
	rowsReturned int,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	// MagicORM no longer collects data - this is a no-op
	// Data collection is handled by magicCommon/monitoring system
}

// RecordTransaction is a helper to record transactions for monitoring
func RecordTransaction(
	operation string,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	// MagicORM no longer collects data - this is a no-op
	// Data collection is handled by magicCommon/monitoring system
}

// GetModelName extracts model name for monitoring
func GetModelName(model models.Model) string {
	if model == nil {
		return "unknown"
	}

	// Try to get the model name
	// This is a simplified implementation
	return "model"
}

// GetFilterModelName extracts model name from filter for monitoring
func GetFilterModelName(filter models.Filter) string {
	if filter == nil {
		return "unknown"
	}

	// Try to get the model name from filter
	// This is a simplified implementation
	return "filter_model"
}

// Convenience functions

// NewMonitoredOrmWithDefaultConfig creates monitored ORM with default configuration
func NewMonitoredOrmWithDefaultConfig(
	provider provider.Provider,
	cfg database.Config,
	prefix string,
) (Orm, *cd.Error) {

	config := DefaultMonitoredOrmConfig()
	return NewMonitoredOrm(provider, cfg, prefix, config)
}

// WrapExistingOrmWithMonitoring wraps an existing ORM with monitoring
func WrapExistingOrmWithMonitoring(orm Orm, config MonitoredOrmConfig) Orm {
	return wrapOrmWithMonitoring(orm, config)
}

// WrapExistingOrmWithDefaultMonitoring wraps an existing ORM with default monitoring
func WrapExistingOrmWithDefaultMonitoring(orm Orm) Orm {
	config := DefaultMonitoredOrmConfig()
	return wrapOrmWithMonitoring(orm, config)
}

// Monitoring integration points for existing ORM implementation
// Note: MagicORM no longer collects data - these functions are now deprecated
// Data collection is handled by magicCommon/monitoring system

var globalMonitoringEnabled = false
var globalMonitoringConfig = DefaultMonitoredOrmConfig()

// EnableGlobalMonitoring enables global monitoring for all ORM operations
// Deprecated: MagicORM no longer collects data
func EnableGlobalMonitoring(config MonitoredOrmConfig) {
	globalMonitoringEnabled = false // Always disabled - MagicORM no longer collects data
}

// DisableGlobalMonitoring disables global monitoring
// Deprecated: MagicORM no longer collects data
func DisableGlobalMonitoring() {
	globalMonitoringEnabled = false
}

// IsGlobalMonitoringEnabled returns whether global monitoring is enabled
// Deprecated: MagicORM no longer collects data
func IsGlobalMonitoringEnabled() bool {
	return false // Always false - MagicORM no longer collects data
}

// GetGlobalMonitoringConfig returns the global monitoring configuration
// Deprecated: MagicORM no longer collects data
func GetGlobalMonitoringConfig() MonitoredOrmConfig {
	return DefaultMonitoredOrmConfig()
}

// CreateOrmWithGlobalMonitoring creates an ORM with global monitoring configuration
// Deprecated: MagicORM no longer collects data
func CreateOrmWithGlobalMonitoring(
	provider provider.Provider,
	cfg database.Config,
	prefix string,
) (Orm, *cd.Error) {

	// MagicORM no longer collects data - always return base ORM
	return NewOrm(provider, cfg, prefix)
}
