package orm

import (
	"fmt"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

// MonitoredOrm wraps an Orm interface with monitoring capabilities
type MonitoredOrm struct {
	orm     Orm
	monitor *ORMMonitor
	config  *MonitoringConfig
}

// Orm is the interface that MonitoredOrm wraps
type Orm interface {
	Create(entity models.Model) *cd.Error
	Drop(entity models.Model) *cd.Error
	Insert(entity models.Model) (models.Model, *cd.Error)
	Update(entity models.Model) (models.Model, *cd.Error)
	Delete(entity models.Model) (models.Model, *cd.Error)
	Query(entity models.Model) (models.Model, *cd.Error)
	Count(filter models.Filter) (int64, *cd.Error)
	BatchQuery(filter models.Filter) ([]models.Model, *cd.Error)
	BeginTransaction() *cd.Error
	CommitTransaction() *cd.Error
	RollbackTransaction() *cd.Error
	Release()
}

// MonitoringConfig holds configuration for ORM monitoring
type MonitoringConfig struct {
	Enabled            bool
	RecordOperations   bool
	RecordQueries      bool
	RecordTransactions bool
	RecordCache        bool
	RecordDatabase     bool
	SampleRate         float64 // 0.0 to 1.0
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:            true,
		RecordOperations:   true,
		RecordQueries:      true,
		RecordTransactions: true,
		RecordCache:        true,
		RecordDatabase:     true,
		SampleRate:         1.0,
	}
}

// NewMonitoredOrm creates a new monitored ORM wrapper
func NewMonitoredOrm(orm Orm, monitor *ORMMonitor, config *MonitoringConfig) *MonitoredOrm {
	if config == nil {
		defaultConfig := DefaultMonitoringConfig()
		config = &defaultConfig
	}

	return &MonitoredOrm{
		orm:     orm,
		monitor: monitor,
		config:  config,
	}
}

// Create creates a table with monitoring
func (m *MonitoredOrm) Create(entity models.Model) *cd.Error {
	if !m.config.Enabled || !m.config.RecordOperations || !shouldSample(m.config.SampleRate) {
		return m.orm.Create(entity)
	}

	startTime := time.Now()
	err := m.orm.Create(entity)

	m.monitor.RecordOperation(
		OperationCreate,
		getModelName(entity),
		startTime,
		err,
		nil,
	)

	return err
}

// Drop drops a table with monitoring
func (m *MonitoredOrm) Drop(entity models.Model) *cd.Error {
	if !m.config.Enabled || !m.config.RecordOperations || !shouldSample(m.config.SampleRate) {
		return m.orm.Drop(entity)
	}

	startTime := time.Now()
	err := m.orm.Drop(entity)

	m.monitor.RecordOperation(
		OperationDrop,
		getModelName(entity),
		startTime,
		err,
		nil,
	)

	return err
}

// Insert inserts an entity with monitoring
func (m *MonitoredOrm) Insert(entity models.Model) (models.Model, *cd.Error) {
	if !m.config.Enabled || !m.config.RecordOperations || !shouldSample(m.config.SampleRate) {
		return m.orm.Insert(entity)
	}

	startTime := time.Now()
	result, err := m.orm.Insert(entity)

	m.monitor.RecordOperation(
		OperationInsert,
		getModelName(entity),
		startTime,
		err,
		nil,
	)

	return result, err
}

// Update updates an entity with monitoring
func (m *MonitoredOrm) Update(entity models.Model) (models.Model, *cd.Error) {
	if !m.config.Enabled || !m.config.RecordOperations || !shouldSample(m.config.SampleRate) {
		return m.orm.Update(entity)
	}

	startTime := time.Now()
	result, err := m.orm.Update(entity)

	m.monitor.RecordOperation(
		OperationUpdate,
		getModelName(entity),
		startTime,
		err,
		nil,
	)

	return result, err
}

// Delete deletes an entity with monitoring
func (m *MonitoredOrm) Delete(entity models.Model) (models.Model, *cd.Error) {
	if !m.config.Enabled || !m.config.RecordOperations || !shouldSample(m.config.SampleRate) {
		return m.orm.Delete(entity)
	}

	startTime := time.Now()
	result, err := m.orm.Delete(entity)

	m.monitor.RecordOperation(
		OperationDelete,
		getModelName(entity),
		startTime,
		err,
		nil,
	)

	return result, err
}

// Query queries an entity with monitoring
func (m *MonitoredOrm) Query(entity models.Model) (models.Model, *cd.Error) {
	if !m.config.Enabled || !m.config.RecordQueries || !shouldSample(m.config.SampleRate) {
		return m.orm.Query(entity)
	}

	startTime := time.Now()
	result, err := m.orm.Query(entity)

	// Record as simple query
	m.monitor.RecordQuery(
		getModelName(entity),
		QueryTypeSimple,
		1, // Assuming single row returned for Query()
		startTime,
		err,
		nil,
	)

	return result, err
}

// Count counts entities with monitoring
func (m *MonitoredOrm) Count(filter models.Filter) (int64, *cd.Error) {
	if !m.config.Enabled || !m.config.RecordQueries || !shouldSample(m.config.SampleRate) {
		return m.orm.Count(filter)
	}

	startTime := time.Now()
	count, err := m.orm.Count(filter)

	m.monitor.RecordOperation(
		OperationCount,
		getFilterModelName(filter),
		startTime,
		err,
		map[string]string{
			"filter_type": getFilterType(filter),
		},
	)

	// Also record as query with count
	if err == nil {
		m.monitor.RecordQuery(
			getFilterModelName(filter),
			QueryTypeFilter,
			0, // Count doesn't return rows
			startTime,
			err,
			map[string]string{
				"result_type": "count",
				"count_value": fmt.Sprintf("%d", count),
			},
		)
	}

	return count, err
}

// BatchQuery performs batch query with monitoring
func (m *MonitoredOrm) BatchQuery(filter models.Filter) ([]models.Model, *cd.Error) {
	if !m.config.Enabled || !m.config.RecordQueries || !shouldSample(m.config.SampleRate) {
		return m.orm.BatchQuery(filter)
	}

	startTime := time.Now()
	results, err := m.orm.BatchQuery(filter)

	rowsReturned := 0
	if err == nil && results != nil {
		rowsReturned = len(results)
	}

	m.monitor.RecordBatchOperation(
		OperationBatch,
		getFilterModelName(filter),
		rowsReturned,
		startTime,
		rowsReturned,
		0, // Assuming all successful for batch query
		err,
		map[string]string{
			"filter_type": getFilterType(filter),
			"operation":   "batch_query",
		},
	)

	// Also record as query
	m.monitor.RecordQuery(
		getFilterModelName(filter),
		QueryTypeBatch,
		rowsReturned,
		startTime,
		err,
		map[string]string{
			"filter_type": getFilterType(filter),
		},
	)

	return results, err
}

// BeginTransaction begins a transaction with monitoring
func (m *MonitoredOrm) BeginTransaction() *cd.Error {
	if !m.config.Enabled || !m.config.RecordTransactions || !shouldSample(m.config.SampleRate) {
		return m.orm.BeginTransaction()
	}

	startTime := time.Now()
	err := m.orm.BeginTransaction()

	m.monitor.RecordTransaction(
		"begin",
		startTime,
		err,
		nil,
	)

	return err
}

// CommitTransaction commits a transaction with monitoring
func (m *MonitoredOrm) CommitTransaction() *cd.Error {
	if !m.config.Enabled || !m.config.RecordTransactions || !shouldSample(m.config.SampleRate) {
		return m.orm.CommitTransaction()
	}

	startTime := time.Now()
	err := m.orm.CommitTransaction()

	m.monitor.RecordTransaction(
		"commit",
		startTime,
		err,
		nil,
	)

	return err
}

// RollbackTransaction rolls back a transaction with monitoring
func (m *MonitoredOrm) RollbackTransaction() *cd.Error {
	if !m.config.Enabled || !m.config.RecordTransactions || !shouldSample(m.config.SampleRate) {
		return m.orm.RollbackTransaction()
	}

	startTime := time.Now()
	err := m.orm.RollbackTransaction()

	m.monitor.RecordTransaction(
		"rollback",
		startTime,
		err,
		nil,
	)

	return err
}

// Release releases resources with monitoring
func (m *MonitoredOrm) Release() {
	m.orm.Release()

	// Record release operation if monitoring is enabled
	if m.config.Enabled && m.config.RecordOperations && shouldSample(m.config.SampleRate) {
		// Note: Release doesn't have error or duration typically
		m.monitor.RecordOperation(
			"release",
			"",
			time.Now(),
			nil,
			nil,
		)
	}
}

// RecordCacheAccess records cache access (can be called by underlying implementation)
func (m *MonitoredOrm) RecordCacheAccess(cacheType, operation string, hit bool, duration time.Duration) {
	if !m.config.Enabled || !m.config.RecordCache || !shouldSample(m.config.SampleRate) {
		return
	}

	m.monitor.RecordCacheAccess(
		cacheType,
		operation,
		hit,
		duration,
		nil,
	)
}

// RecordDatabaseOperation records database operation (can be called by underlying implementation)
func (m *MonitoredOrm) RecordDatabaseOperation(dbType, operation string, startTime time.Time, err error) {
	if !m.config.Enabled || !m.config.RecordDatabase || !shouldSample(m.config.SampleRate) {
		return
	}

	m.monitor.RecordDatabaseOperation(
		dbType,
		operation,
		startTime,
		err,
		nil,
	)
}

// RecordConnectionPool records connection pool statistics (can be called by underlying implementation)
func (m *MonitoredOrm) RecordConnectionPool(
	dbType string,
	active, idle, waiting, max int,
) {
	if !m.config.Enabled || !m.config.RecordDatabase || !shouldSample(m.config.SampleRate) {
		return
	}

	m.monitor.RecordConnectionPool(
		dbType,
		active,
		idle,
		waiting,
		max,
		nil,
	)
}

// Helper functions

func getModelName(entity models.Model) string {
	if entity == nil {
		return "unknown"
	}

	// Try to get model name
	// This is a simplified implementation
	return "model"
}

func getFilterModelName(filter models.Filter) string {
	if filter == nil {
		return "unknown"
	}

	// Try to get model name from filter
	// This is a simplified implementation
	return "filter_model"
}

func getFilterType(filter models.Filter) string {
	if filter == nil {
		return "unknown"
	}

	// Determine filter type
	// This is a simplified implementation
	return "standard"
}

func shouldSample(sampleRate float64) bool {
	if sampleRate >= 1.0 {
		return true
	}
	if sampleRate <= 0.0 {
		return false
	}

	// Simple sampling - in production use proper sampling algorithm
	// For now, always sample if rate > 0
	return true
}

// Convenience functions

// WrapOrmWithMonitoring wraps an existing ORM with monitoring
func WrapOrmWithMonitoring(orm Orm, monitor *ORMMonitor, config *MonitoringConfig) *MonitoredOrm {
	return NewMonitoredOrm(orm, monitor, config)
}

// WrapOrmWithDefaultMonitoring wraps an existing ORM with default monitoring
func WrapOrmWithDefaultMonitoring(orm Orm) *MonitoredOrm {
	monitor := DefaultORMMonitor()
	config := DefaultMonitoringConfig()
	return NewMonitoredOrm(orm, monitor, &config)
}
