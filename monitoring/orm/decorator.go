// Package orm provides ORM monitoring decorators.
// This is a simplified version that only provides monitoring decoration,
// relying on external monitoring systems for collection and export.
package orm

import (
	"fmt"
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/monitoring"
)

// Use types directly from the monitoring package

// MonitoredOrm is a simplified decorator that adds monitoring to ORM operations.
type MonitoredOrm struct {
	orm       Orm
	collector ORMCollector
	enabled   bool
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

// NewMonitoredOrm creates a new monitored ORM wrapper.
func NewMonitoredOrm(orm Orm, collector ORMCollector) *MonitoredOrm {
	return &MonitoredOrm{
		orm:       orm,
		collector: collector,
		enabled:   collector != nil,
	}
}

// Create creates a table with monitoring.
func (m *MonitoredOrm) Create(entity models.Model) *cd.Error {
	startTime := time.Now()
	err := m.orm.Create(entity)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationCreate,
			getModelName(entity),
			startTime,
			err,
			nil,
		)
	}

	return err
}

// Drop drops a table with monitoring.
func (m *MonitoredOrm) Drop(entity models.Model) *cd.Error {
	startTime := time.Now()
	err := m.orm.Drop(entity)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationDrop,
			getModelName(entity),
			startTime,
			err,
			nil,
		)
	}

	return err
}

// Insert inserts an entity with monitoring.
func (m *MonitoredOrm) Insert(entity models.Model) (models.Model, *cd.Error) {
	startTime := time.Now()
	result, err := m.orm.Insert(entity)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationInsert,
			getModelName(entity),
			startTime,
			err,
			nil,
		)
	}

	return result, err
}

// Update updates an entity with monitoring.
func (m *MonitoredOrm) Update(entity models.Model) (models.Model, *cd.Error) {
	startTime := time.Now()
	result, err := m.orm.Update(entity)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationUpdate,
			getModelName(entity),
			startTime,
			err,
			nil,
		)
	}

	return result, err
}

// Delete deletes an entity with monitoring.
func (m *MonitoredOrm) Delete(entity models.Model) (models.Model, *cd.Error) {
	startTime := time.Now()
	result, err := m.orm.Delete(entity)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationDelete,
			getModelName(entity),
			startTime,
			err,
			nil,
		)
	}

	return result, err
}

// Query queries an entity with monitoring.
func (m *MonitoredOrm) Query(entity models.Model) (models.Model, *cd.Error) {
	startTime := time.Now()
	result, err := m.orm.Query(entity)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationQuery,
			getModelName(entity),
			startTime,
			err,
			nil,
		)
	}

	return result, err
}

// Count counts entities with monitoring.
func (m *MonitoredOrm) Count(filter models.Filter) (int64, *cd.Error) {
	startTime := time.Now()
	count, err := m.orm.Count(filter)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationCount,
			getFilterModelName(filter),
			startTime,
			err,
			nil,
		)
	}

	return count, err
}

// BatchQuery performs batch query with monitoring.
func (m *MonitoredOrm) BatchQuery(filter models.Filter) ([]models.Model, *cd.Error) {
	startTime := time.Now()
	results, err := m.orm.BatchQuery(filter)

	if m.enabled {
		m.collector.RecordOperation(
			monitoring.OperationBatch,
			getFilterModelName(filter),
			startTime,
			err,
			map[string]string{"batch_size": fmt.Sprintf("%d", len(results))},
		)
	}

	return results, err
}

// BeginTransaction begins a transaction with monitoring.
func (m *MonitoredOrm) BeginTransaction() *cd.Error {
	startTime := time.Now()
	err := m.orm.BeginTransaction()

	if m.enabled {
		m.collector.RecordTransaction("begin", startTime, err, nil)
	}

	return err
}

// CommitTransaction commits a transaction with monitoring.
func (m *MonitoredOrm) CommitTransaction() *cd.Error {
	startTime := time.Now()
	err := m.orm.CommitTransaction()

	if m.enabled {
		m.collector.RecordTransaction("commit", startTime, err, nil)
	}

	return err
}

// RollbackTransaction rolls back a transaction with monitoring.
func (m *MonitoredOrm) RollbackTransaction() *cd.Error {
	startTime := time.Now()
	err := m.orm.RollbackTransaction()

	if m.enabled {
		m.collector.RecordTransaction("rollback", startTime, err, nil)
	}

	return err
}

// Release releases resources.
func (m *MonitoredOrm) Release() {
	m.orm.Release()
}

// Helper functions

func getModelName(model models.Model) string {
	if model == nil {
		return "unknown"
	}

	// Try to get model name from reflection
	modelType := reflect.TypeOf(model)
	if modelType == nil {
		return "unknown"
	}

	// Handle pointer types
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return modelType.Name()
}

func getFilterModelName(filter models.Filter) string {
	if filter == nil {
		return "unknown"
	}

	// This is a simplified implementation
	// In real code, you would extract model name from filter
	return "filter"
}
