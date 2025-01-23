package orm

import (
	"fmt"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

const maxDeepLevel = 3

// Orm orm interface
type Orm interface {
	Create(entity model.Model) *cd.Result
	Drop(entity model.Model) *cd.Result
	Insert(entity model.Model) (model.Model, *cd.Result)
	Update(entity model.Model) (model.Model, *cd.Result)
	Delete(entity model.Model) (model.Model, *cd.Result)
	Query(entity model.Model) (model.Model, *cd.Result)
	Count(filter model.Filter) (int64, *cd.Result)
	BatchQuery(filter model.Filter) ([]model.Model, *cd.Result)
	BeginTransaction() *cd.Result
	CommitTransaction() *cd.Result
	RollbackTransaction() *cd.Result
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
func NewExecutor(config executor.Config) (executor.Executor, *cd.Result) {
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
		name2Pool.Range(func(_, val any) bool {
			pool := val.(executor.Pool)
			pool.Uninitialized()

			return true
		})

		name2Pool = sync.Map{}
	})
}

func AddDatabase(dbServer, dbName, username, password, charSet string, maxConnNum int, owner string) (err *cd.Result) {
	config := NewConfig(dbServer, dbName, username, password, charSet)

	val, ok := name2Pool.Load(owner)
	if ok {
		pool := val.(executor.Pool)
		pool.IncReference()
		err = pool.CheckConfig(config)
		return
	}

	pool := NewPool()
	err = pool.Initialize(maxConnNum, config)
	if err != nil {
		log.Errorf("AddDatabase failed, pool.Initialize error:%s", err.Error())
		return
	}

	pool.IncReference()
	name2Pool.Store(owner, pool)
	return
}

func DelDatabase(owner string) {
	val, ok := name2Pool.Load(owner)
	if !ok {
		return
	}

	pool := val.(executor.Pool)
	if pool.DecReference() == 0 {
		pool.Uninitialized()
		name2Pool.Delete(owner)
	}
}

// NewOrm create new Orm
func NewOrm(provider provider.Provider, cfg executor.Config, prefix string) (Orm, *cd.Result) {
	executorVal, executorErr := NewExecutor(cfg)
	if executorErr != nil {
		log.Errorf("NewOrm failed, NewExecutor error:%s", executorErr.Error())
		return nil, cd.NewResult(cd.UnExpected, executorErr.Error())
	}

	orm := &impl{executor: executorVal, modelProvider: provider, modelCodec: codec.New(provider, prefix)}
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(provider provider.Provider, prefix string) (ret Orm, err *cd.Result) {
	val, ok := name2Pool.Load(provider.Owner())
	if !ok {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("can't find orm,name:%s", provider.Owner()))
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

	ret = &impl{executor: executorVal, modelProvider: provider, modelCodec: codec.New(provider, prefix)}
	return
}

// impl orm
type impl struct {
	executor      executor.Executor
	modelProvider provider.Provider
	modelCodec    codec.Codec
}

// BeginTransaction begin transaction
func (s *impl) BeginTransaction() (err *cd.Result) {
	if s.executor != nil {
		err = s.executor.BeginTransaction()
		if err != nil {
			log.Errorf("BeginTransaction failed, s.executor.BeginTransaction error:%s", err.Error())
		}
	}

	return
}

// CommitTransaction commit transaction
func (s *impl) CommitTransaction() (err *cd.Result) {
	if s.executor != nil {
		err = s.executor.CommitTransaction()
		if err != nil {
			log.Errorf("CommitTransaction failed, s.executor.CommitTransaction error:%s", err.Error())
		}
	}

	return
}

// RollbackTransaction rollback transaction
func (s *impl) RollbackTransaction() (err *cd.Result) {
	if s.executor != nil {
		err = s.executor.RollbackTransaction()
		if err != nil {
			log.Errorf("RollbackTransaction failed, s.executor.RollbackTransaction error:%s", err.Error())
		}
	}

	return
}

func (s *impl) finalTransaction(err *cd.Result) {
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
