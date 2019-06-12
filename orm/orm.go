package orm

import (
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Orm orm
type Orm struct {
	executor      executor.Executor
	modelProvider provider.Provider
}

// NewFilter create new filter
func NewFilter(modelProvider provider.Provider) model.Filter {
	return &queryFilter{params: map[string]model.FilterItem{}, modelProvider: modelProvider}
}

// New create new Orm
func New(executor executor.Executor, modelProvider provider.Provider) *Orm {
	return &Orm{executor: executor, modelProvider: modelProvider}
}

// RegisterModel register model
func (s *Orm) RegisterModel(entity interface{}, owner string) error {
	return s.modelProvider.RegisterModel(entity)
}

// UnregisterModel unregister model
func (s *Orm) UnregisterModel(entity interface{}, owner string) {
	s.modelProvider.UnregisterModel(entity)
}

// BeginTransaction begin transaction
func (s *Orm) BeginTransaction() {
	if s.executor != nil {
		s.executor.BeginTransaction()
	}
}

// CommitTransaction commit transaction
func (s *Orm) CommitTransaction() {
	if s.executor != nil {
		s.executor.CommitTransaction()
	}
}

// RollbackTransaction rollbacktransaction
func (s *Orm) RollbackTransaction() {
	if s.executor != nil {
		s.executor.RollbackTransaction()
	}
}
