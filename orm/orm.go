package orm

import (
	"fmt"
	"sync"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

const maxDeepLevel = 3

// Orm orm interface
type Orm interface {
	Create(entity model.Model) error
	Drop(entity model.Model) error
	Insert(entity model.Model) (model.Model, error)
	Update(entity model.Model) (model.Model, error)
	Delete(entity model.Model) (model.Model, error)
	Query(entity model.Model) (model.Model, error)
	Count(filter model.Filter) (int64, error)
	BatchQuery(filter model.Filter) ([]model.Model, error)
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Release()
}

var name2Pool sync.Map
var name2PoolInitializeOnce sync.Once
var name2PoolUninitializedOnce sync.Once

// NewPool new executor pool
func NewPool() executor.Pool {
	return executor.NewPool()
}

// NewExecutor NewExecutor
func NewExecutor(config executor.Config) (executor.Executor, error) {
	return executor.NewExecutor(config)
}

func NewConfig(dbServer, dbName, username, password, charSet string) executor.Config {
	return executor.NewConfig(dbServer, dbName, username, password, charSet)
}

// Initialize InitOrm
func Initialize() {
	name2PoolInitializeOnce.Do(func() {
		name2Pool = sync.Map{}
	})
}

// Uninitialized orm
func Uninitialized() {
	name2PoolUninitializedOnce.Do(func() {
		name2Pool.Range(func(_, val interface{}) bool {
			pool := val.(executor.Pool)
			pool.Uninitialized()

			return true
		})

		name2Pool = sync.Map{}
	})
}

func AddDatabase(dbServer, dbName, username, password, charSet string, maxConnNum int, owner string) (err error) {
	config := NewConfig(dbServer, dbName, username, password, charSet)

	val, ok := name2Pool.Load(owner)
	if ok {
		pool := val.(executor.Pool)
		return pool.CheckConfig(config)
	}

	pool := NewPool()
	err = pool.Initialize(maxConnNum, config)
	if err != nil {
		log.Errorf("AddDatabase failed, pool.Initialize error:%s", err.Error())
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
func NewOrm(provider provider.Provider, cfg executor.Config, prefix string) (Orm, error) {
	executor, err := NewExecutor(cfg)
	if err != nil {
		log.Errorf("NewOrm failed, NewExecutor error:%s", err.Error())
		return nil, err
	}

	orm := &impl{executor: executor, modelProvider: provider, specialPrefix: prefix}
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(provider provider.Provider, prefix string) (ret Orm, err error) {
	val, ok := name2Pool.Load(provider.Owner())
	if !ok {
		err = fmt.Errorf("can't find orm,name:%s", provider.Owner())
		log.Errorf("GetOrm failed, error:%s", err.Error())
		return
	}

	pool := val.(executor.Pool)
	executorVal, executorErr := pool.GetExecutor()
	if executorErr != nil {
		err = executorErr
		log.Errorf("GetOrm failed, pool.GetExecutor error:%s", err.Error())
		return
	}

	ret = &impl{executor: executorVal, modelProvider: provider, specialPrefix: prefix}
	return
}

// impl orm
type impl struct {
	executor      executor.Executor
	modelProvider provider.Provider
	specialPrefix string
}

// BeginTransaction begin transaction
func (s *impl) BeginTransaction() (err error) {
	if s.executor != nil {
		err = s.executor.BeginTransaction()
		if err != nil {
			log.Errorf("BeginTransaction failed, s.executor.BeginTransaction error:%s", err.Error())
		}
	}

	return
}

// CommitTransaction commit transaction
func (s *impl) CommitTransaction() (err error) {
	if s.executor != nil {
		err = s.executor.CommitTransaction()
		if err != nil {
			log.Errorf("CommitTransaction failed, s.executor.CommitTransaction error:%s", err.Error())
		}
	}

	return
}

// RollbackTransaction rollback transaction
func (s *impl) RollbackTransaction() (err error) {
	if s.executor != nil {
		err = s.executor.RollbackTransaction()
		if err != nil {
			log.Errorf("RollbackTransaction failed, s.executor.RollbackTransaction error:%s", err.Error())
		}
	}

	return
}

func (s *impl) finalTransaction(err error) {
	if err == nil {
		err = s.executor.CommitTransaction()
		if err != nil {
			log.Errorf("finalTransaction failed, s.executor.CommitTransaction error:%s", err.Error())
		}
		return
	}

	err = s.executor.RollbackTransaction()
	if err != nil {
		log.Errorf("finalTransaction failed, s.executor.RollbackTransaction error:%s", err.Error())
	}
	return
}

func (s *impl) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}
