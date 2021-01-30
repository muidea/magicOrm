package orm

import (
	"fmt"
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

var _config *ormConfig

// Initialize InitOrm
func Initialize(maxConnNum int, user, password, address, dbName string) error {
	cfg := &serverConfig{user: user, password: password, address: address, dbName: dbName}

	_config = newConfig()

	_config.updateServerConfig(cfg)

	return executor.InitializePool(maxConnNum, user, password, address, dbName)
}

// Uninitialize Uninitialize orm
func Uninitialize() {
	executor.UninitializePool()
}

// NewOrm create new Orm
func NewOrm(provider provider.Provider) (Orm, error) {
	cfg := _config.getServerConfig()
	if cfg == nil {
		return nil, fmt.Errorf("not define databaes server config")
	}

	executor, err := executor.NewExecutor(cfg.user, cfg.password, cfg.address, cfg.dbName)
	if err != nil {
		return nil, err
	}

	orm := &impl{executor: executor, modelProvider: provider}
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(provider provider.Provider) (Orm, error) {
	executor, err := executor.GetExecutor()
	if err != nil {
		return nil, err
	}

	orm := &impl{executor: executor, modelProvider: provider}
	return orm, nil
}

func GetFilter(provider provider.Provider) model.Filter {
	return &queryFilter{params: map[string]model.FilterItem{}, modelProvider: provider}
}

// impl orm
type impl struct {
	executor      executor.Executor
	modelProvider provider.Provider
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
