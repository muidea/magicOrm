package database

import (
	"context"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/monitoring/core"
)

// DatabaseMonitoringFactory creates monitored database components
type DatabaseMonitoringFactory struct {
	collector *core.Collector
	config    *core.MonitoringConfig
	monitor   *DatabaseMonitor
}

// NewDatabaseMonitoringFactory creates a new database monitoring factory
func NewDatabaseMonitoringFactory(collector *core.Collector, config *core.MonitoringConfig) *DatabaseMonitoringFactory {
	return &DatabaseMonitoringFactory{
		collector: collector,
		config:    config,
	}
}

// CreateDatabaseMonitor creates a database monitor
func (f *DatabaseMonitoringFactory) CreateDatabaseMonitor() *DatabaseMonitor {
	if f.monitor == nil {
		if f.collector == nil {
			config := core.DefaultMonitoringConfig()
			collector := core.NewCollector(&config)
			f.collector = collector
			f.config = &config
		}
		f.monitor = NewDatabaseMonitor(f.collector, f.config)
	}
	return f.monitor
}

// WrapExecutor wraps an executor with monitoring
func (f *DatabaseMonitoringFactory) WrapExecutor(executor database.Executor, dbType string) *MonitoredExecutor {
	monitor := f.CreateDatabaseMonitor()
	return NewMonitoredExecutor(executor, monitor, dbType)
}

// WrapPool wraps a pool with monitoring
func (f *DatabaseMonitoringFactory) WrapPool(pool database.Pool, dbType string) *MonitoredPool {
	monitor := f.CreateDatabaseMonitor()
	return NewMonitoredPool(pool, monitor, dbType)
}

// CreateMonitoredPostgresPool creates a monitored PostgreSQL pool
func (f *DatabaseMonitoringFactory) CreateMonitoredPostgresPool(ctx context.Context, maxConnNum int, config database.Config) (database.Pool, error) {
	// This would need to integrate with the actual PostgreSQL pool creation
	// For now, returns a wrapper around an existing pool
	monitor := f.CreateDatabaseMonitor()

	// In a real implementation, this would create the actual pool
	// and wrap it with monitoring
	return &MonitoredPool{
		// Pool implementation would go here
		monitor: monitor,
		dbType:  "postgresql",
	}, nil
}

// CreateMonitoredMySQLPool creates a monitored MySQL pool
func (f *DatabaseMonitoringFactory) CreateMonitoredMySQLPool(ctx context.Context, maxConnNum int, config database.Config) (database.Pool, error) {
	monitor := f.CreateDatabaseMonitor()

	return &MonitoredPool{
		// Pool implementation would go here
		monitor: monitor,
		dbType:  "mysql",
	}, nil
}

// GetDatabaseMetrics returns database-specific metrics
func (f *DatabaseMonitoringFactory) GetDatabaseMetrics() (map[string][]core.Metric, error) {
	monitor := f.CreateDatabaseMonitor()
	if monitor == nil || monitor.collector == nil {
		return nil, nil
	}

	return monitor.collector.GetMetrics(), nil
}

// ResetDatabaseMetrics resets database metrics
func (f *DatabaseMonitoringFactory) ResetDatabaseMetrics() {
	monitor := f.CreateDatabaseMonitor()
	if monitor != nil && monitor.collector != nil {
		monitor.collector.Reset()
	}
}

// GetDatabaseStats returns database monitoring statistics
func (f *DatabaseMonitoringFactory) GetDatabaseStats() core.CollectorStats {
	monitor := f.CreateDatabaseMonitor()
	if monitor == nil || monitor.collector == nil {
		return core.CollectorStats{}
	}

	return monitor.collector.GetStats()
}

// SimpleDatabaseMonitoringFactory is a simplified factory for basic use cases
type SimpleDatabaseMonitoringFactory struct {
	monitor *DatabaseMonitor
}

// NewSimpleDatabaseMonitoringFactory creates a simple database monitoring factory
func NewSimpleDatabaseMonitoringFactory(config *core.MonitoringConfig) *SimpleDatabaseMonitoringFactory {
	collector := core.NewCollector(config)
	monitor := NewDatabaseMonitor(collector, config)

	return &SimpleDatabaseMonitoringFactory{
		monitor: monitor,
	}
}

// GetMonitor returns the database monitor
func (f *SimpleDatabaseMonitoringFactory) GetMonitor() *DatabaseMonitor {
	return f.monitor
}

// WrapExecutor wraps an executor with monitoring
func (f *SimpleDatabaseMonitoringFactory) WrapExecutor(executor database.Executor, dbType string) *MonitoredExecutor {
	return NewMonitoredExecutor(executor, f.monitor, dbType)
}

// WrapPool wraps a pool with monitoring
func (f *SimpleDatabaseMonitoringFactory) WrapPool(pool database.Pool, dbType string) *MonitoredPool {
	return NewMonitoredPool(pool, f.monitor, dbType)
}
