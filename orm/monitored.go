package orm

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/monitoring/core"
	monitoringorm "github.com/muidea/magicOrm/monitoring/orm"
	"github.com/muidea/magicOrm/monitoring/unified"
	"github.com/muidea/magicOrm/provider"
)

// MonitoredOrmConfig holds configuration for monitored ORM
type MonitoredOrmConfig struct {
	// Enable monitoring
	Enabled bool

	// Monitoring configuration
	MonitoringConfig *core.MonitoringConfig

	// ORM-specific monitoring configuration
	ORMMonitoringConfig *monitoringorm.MonitoringConfig

	// Custom labels for metrics
	CustomLabels map[string]string

	// Auto-start metrics exporter
	AutoStartExporter bool
}

// DefaultMonitoredOrmConfig returns default monitored ORM configuration
func DefaultMonitoredOrmConfig() MonitoredOrmConfig {
	monitoringConfig := core.DefaultMonitoringConfig()
	ormMonitoringConfig := monitoringorm.DefaultMonitoringConfig()

	return MonitoredOrmConfig{
		Enabled:             true,
		MonitoringConfig:    &monitoringConfig,
		ORMMonitoringConfig: &ormMonitoringConfig,
		CustomLabels:        make(map[string]string),
		AutoStartExporter:   true,
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
	// Create monitoring factory
	factory := unified.NewMonitoringFactory(config.MonitoringConfig)

	// Add custom labels
	if len(config.CustomLabels) > 0 {
		factory.AddCustomLabels(config.CustomLabels)
	}

	// Get ORM monitor
	ormMonitor := factory.GetORMMonitor()
	if ormMonitor == nil {
		// Create default monitor if factory doesn't provide one
		ormMonitor = monitoringorm.DefaultORMMonitor()
	}

	// Create monitored ORM wrapper
	monitoredOrm := monitoringorm.NewMonitoredOrm(
		wrapOrmInterface(orm),
		ormMonitor,
		config.ORMMonitoringConfig,
	)

	// Start exporter if auto-start is enabled
	if config.AutoStartExporter {
		exporter := factory.GetExporter()
		if exporter != nil {
			exporter.Start()
		}
	}

	// Store factory for later access if needed
	// (This would require modifying the monitored ORM wrapper to store the factory)

	return convertToOrmInterface(monitoredOrm)
}

// wrapOrmInterface wraps the Orm interface for the monitoring package
func wrapOrmInterface(orm Orm) monitoringorm.Orm {
	return &ormWrapper{orm: orm}
}

// convertToOrmInterface converts monitored ORM back to Orm interface
func convertToOrmInterface(monitoredOrm *monitoringorm.MonitoredOrm) Orm {
	return &monitoredOrmWrapper{monitoredOrm: monitoredOrm}
}

// Wrapper types for interface conversion

type ormWrapper struct {
	orm Orm
}

func (w *ormWrapper) Create(entity models.Model) *cd.Error {
	return w.orm.Create(entity)
}

func (w *ormWrapper) Drop(entity models.Model) *cd.Error {
	return w.orm.Drop(entity)
}

func (w *ormWrapper) Insert(entity models.Model) (models.Model, *cd.Error) {
	return w.orm.Insert(entity)
}

func (w *ormWrapper) Update(entity models.Model) (models.Model, *cd.Error) {
	return w.orm.Update(entity)
}

func (w *ormWrapper) Delete(entity models.Model) (models.Model, *cd.Error) {
	return w.orm.Delete(entity)
}

func (w *ormWrapper) Query(entity models.Model) (models.Model, *cd.Error) {
	return w.orm.Query(entity)
}

func (w *ormWrapper) Count(filter models.Filter) (int64, *cd.Error) {
	return w.orm.Count(filter)
}

func (w *ormWrapper) BatchQuery(filter models.Filter) ([]models.Model, *cd.Error) {
	return w.orm.BatchQuery(filter)
}

func (w *ormWrapper) BeginTransaction() *cd.Error {
	return w.orm.BeginTransaction()
}

func (w *ormWrapper) CommitTransaction() *cd.Error {
	return w.orm.CommitTransaction()
}

func (w *ormWrapper) RollbackTransaction() *cd.Error {
	return w.orm.RollbackTransaction()
}

func (w *ormWrapper) Release() {
	w.orm.Release()
}

type monitoredOrmWrapper struct {
	monitoredOrm *monitoringorm.MonitoredOrm
}

func (w *monitoredOrmWrapper) Create(entity models.Model) *cd.Error {
	return w.monitoredOrm.Create(entity)
}

func (w *monitoredOrmWrapper) Drop(entity models.Model) *cd.Error {
	return w.monitoredOrm.Drop(entity)
}

func (w *monitoredOrmWrapper) Insert(entity models.Model) (models.Model, *cd.Error) {
	return w.monitoredOrm.Insert(entity)
}

func (w *monitoredOrmWrapper) Update(entity models.Model) (models.Model, *cd.Error) {
	return w.monitoredOrm.Update(entity)
}

func (w *monitoredOrmWrapper) Delete(entity models.Model) (models.Model, *cd.Error) {
	return w.monitoredOrm.Delete(entity)
}

func (w *monitoredOrmWrapper) Query(entity models.Model) (models.Model, *cd.Error) {
	return w.monitoredOrm.Query(entity)
}

func (w *monitoredOrmWrapper) Count(filter models.Filter) (int64, *cd.Error) {
	return w.monitoredOrm.Count(filter)
}

func (w *monitoredOrmWrapper) BatchQuery(filter models.Filter) ([]models.Model, *cd.Error) {
	return w.monitoredOrm.BatchQuery(filter)
}

func (w *monitoredOrmWrapper) BeginTransaction() *cd.Error {
	return w.monitoredOrm.BeginTransaction()
}

func (w *monitoredOrmWrapper) CommitTransaction() *cd.Error {
	return w.monitoredOrm.CommitTransaction()
}

func (w *monitoredOrmWrapper) RollbackTransaction() *cd.Error {
	return w.monitoredOrm.RollbackTransaction()
}

func (w *monitoredOrmWrapper) Release() {
	w.monitoredOrm.Release()
}

// Helper functions for monitoring integration

// RecordOperation is a helper to record ORM operations for monitoring
func RecordOperation(
	operation monitoringorm.OperationType,
	modelName string,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	// This would integrate with a global monitoring system
	// For now, it's a no-op
}

// RecordQuery is a helper to record ORM queries for monitoring
func RecordQuery(
	modelName string,
	queryType monitoringorm.QueryType,
	rowsReturned int,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	// This would integrate with a global monitoring system
	// For now, it's a no-op
}

// RecordTransaction is a helper to record transactions for monitoring
func RecordTransaction(
	operation string,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	// This would integrate with a global monitoring system
	// For now, it's a no-op
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

// These functions can be called from the existing ORM implementation
// to add monitoring without modifying the core logic

var globalMonitoringEnabled = false
var globalMonitoringConfig = DefaultMonitoredOrmConfig()

// EnableGlobalMonitoring enables global monitoring for all ORM operations
func EnableGlobalMonitoring(config MonitoredOrmConfig) {
	globalMonitoringEnabled = true
	globalMonitoringConfig = config
}

// DisableGlobalMonitoring disables global monitoring
func DisableGlobalMonitoring() {
	globalMonitoringEnabled = false
}

// IsGlobalMonitoringEnabled returns whether global monitoring is enabled
func IsGlobalMonitoringEnabled() bool {
	return globalMonitoringEnabled
}

// GetGlobalMonitoringConfig returns the global monitoring configuration
func GetGlobalMonitoringConfig() MonitoredOrmConfig {
	return globalMonitoringConfig
}

// CreateOrmWithGlobalMonitoring creates an ORM with global monitoring configuration
func CreateOrmWithGlobalMonitoring(
	provider provider.Provider,
	cfg database.Config,
	prefix string,
) (Orm, *cd.Error) {

	if globalMonitoringEnabled {
		return NewMonitoredOrm(provider, cfg, prefix, globalMonitoringConfig)
	}

	return NewOrm(provider, cfg, prefix)
}
