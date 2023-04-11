package orm

import (
	"fmt"
	"sync"

	log "github.com/cihub/seelog"

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

var name2Pool sync.Map

// NewPool new executor pool
func NewPool() executor.Pool {
	return executor.NewPool()
}

// NewExecutor NewExecutor
func NewExecutor(config executor.Config) (executor.Executor, error) {
	return executor.NewExecutor(config)
}

func NewConfig(dbServer, dbName, username, password string) executor.Config {
	return executor.NewConfig(dbServer, dbName, username, password)
}

// Initialize InitOrm
func Initialize() {
	name2Pool = sync.Map{}
}

// Uninitialized orm
func Uninitialized() {
	name2Pool.Range(func(_, val interface{}) bool {
		pool := val.(executor.Pool)
		pool.Uninitialized()

		return true
	})

	name2Pool = sync.Map{}
}

func AddDatabase(dbServer, dbName, username, password string, maxConnNum int, owner string) (err error) {
	config := NewConfig(dbServer, dbName, username, password)

	val, ok := name2Pool.Load(owner)
	if ok {
		pool := val.(executor.Pool)
		return pool.CheckConfig(config)
	}

	pool := NewPool()
	err = pool.Initialize(maxConnNum, config)
	if err != nil {
		return
	}

	name2Pool.Store(owner, pool)
	return
}

func DelDatabase(owner string) {
	val, ok := name2Pool.Load(owner)
	if !ok {
		return
	}

	pool := val.(executor.Pool)
	pool.Uninitialized()
	name2Pool.Delete(owner)
}

// NewOrm create new Orm
func NewOrm(provider provider.Provider, cfg executor.Config) (Orm, error) {
	executor, err := NewExecutor(cfg)
	if err != nil {
		return nil, err
	}

	orm := &impl{executor: executor, modelProvider: provider}
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(provider provider.Provider) (ret Orm, err error) {
	val, ok := name2Pool.Load(provider.Owner())
	if !ok {
		err = fmt.Errorf("can't find orm,name:%s", provider.Owner())
		log.Errorf("get orm failed, err:%s", err.Error())
		return
	}

	pool := val.(executor.Pool)
	executorVal, executorErr := pool.GetExecutor()
	if executorErr != nil {
		err = executorErr
		return
	}

	ret = &impl{executor: executorVal, modelProvider: provider}
	return
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
