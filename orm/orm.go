package orm

import (
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Orm orm interface
type Orm interface {
	Create(entity model.Model) error
	Drop(entity model.Model) error
	Insert(entity model.Model) (model.Model, error)
	Update(entity model.Model) (model.Model, error)
	Delete(entity model.Model) (model.Model, error)
	Query(entity model.Model) (model.Model, error)
	Count(entity model.Model, filter model.Filter) (int64, error)
	BatchQuery(entity model.Model, filter model.Filter) ([]model.Model, error)
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Release()
}

// impl orm
type impl struct {
	executor      executor.Executor
	modelProvider provider.Provider
}

// NewFilter create new filter
func NewFilter(modelProvider provider.Provider) model.Filter {
	return &queryFilter{params: map[string]model.FilterItem{}, modelProvider: modelProvider}
}

// New create new impl
func New(executor executor.Executor, modelProvider provider.Provider) Orm {
	return &impl{executor: executor, modelProvider: modelProvider}
}

// BeginTransaction begin transaction
func (s *impl) BeginTransaction() (err error) {
	if s.executor != nil {
		err = s.executor.BeginTransaction()
	}

	return
}

// CommitTransaction commit transaction
func (s *impl) CommitTransaction() (err error) {
	if s.executor != nil {
		err = s.executor.CommitTransaction()
	}

	return
}

// RollbackTransaction rollback transaction
func (s *impl) RollbackTransaction() (err error) {
	if s.executor != nil {
		err = s.executor.RollbackTransaction()
	}

	return
}

func (s *impl) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}

// NewQueryFilter new query filter
func (s *impl) NewQueryFilter() model.Filter {
	return NewFilter(s.modelProvider)
}
