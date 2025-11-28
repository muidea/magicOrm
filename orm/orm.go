package orm

import (
	"context"
	"fmt"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

const maxDeepLevel = 3

// Orm orm interface
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

var name2Pool sync.Map
var name2PoolInitializeOnce sync.Once
var name2PoolUninitializedOnce sync.Once

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
			pool := val.(database.Pool)
			pool.Uninitialized()

			return true
		})

		name2Pool = sync.Map{}
	})
}

func AddDatabase(dbServer, dbName, username, password string, maxConnNum int, owner string) (err *cd.Error) {
	config := NewConfig(dbServer, dbName, username, password)

	val, ok := name2Pool.Load(owner)
	if ok {
		pool := val.(database.Pool)
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

	pool := val.(database.Pool)
	if pool.DecReference() == 0 {
		pool.Uninitialized()
		name2Pool.Delete(owner)
	}
}

// NewOrm create new Orm
func NewOrm(provider provider.Provider, cfg database.Config, prefix string) (Orm, *cd.Error) {
	executorVal, executorErr := NewExecutor(cfg)
	if executorErr != nil {
		log.Errorf("NewOrm failed, NewExecutor error:%s", executorErr.Error())
		return nil, cd.NewError(cd.Unexpected, executorErr.Error())
	}

	orm := &impl{
		context:  context.Background(),
		executor: executorVal, modelProvider: provider, modelCodec: codec.New(provider, prefix)}
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(ctx context.Context, provider provider.Provider, prefix string) (ret Orm, err *cd.Error) {
	val, ok := name2Pool.Load(provider.Owner())
	if !ok {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't find orm,name:%s", provider.Owner()))
		log.Errorf("GetOrm failed, error:%s", err.Error())
		return
	}

	pool := val.(database.Pool)
	executorVal, executorErr := pool.GetExecutor(ctx)
	if executorErr != nil {
		err = executorErr
		log.Errorf("GetOrm failed, pool.GetExecutor error:%s", err.Error())
		return
	}

	ret = &impl{
		context:  ctx,
		executor: executorVal, modelProvider: provider, modelCodec: codec.New(provider, prefix)}
	return
}

// impl orm
type impl struct {
	context       context.Context
	executor      database.Executor
	modelProvider provider.Provider
	modelCodec    codec.Codec
}

// BeginTransaction begin transaction
func (s *impl) BeginTransaction() (err *cd.Error) {
	if s.executor != nil {
		err = s.executor.BeginTransaction()
		if err != nil {
			log.Errorf("BeginTransaction failed, s.executor.BeginTransaction error:%s", err.Error())
		}
	}

	return
}

// CommitTransaction commit transaction
func (s *impl) CommitTransaction() (err *cd.Error) {
	if s.executor != nil {
		err = s.executor.CommitTransaction()
		if err != nil {
			log.Errorf("CommitTransaction failed, s.executor.CommitTransaction error:%s", err.Error())
		}
	}

	return
}

// RollbackTransaction rollback transaction
func (s *impl) RollbackTransaction() (err *cd.Error) {
	if s.executor != nil {
		err = s.executor.RollbackTransaction()
		if err != nil {
			log.Errorf("RollbackTransaction failed, s.executor.RollbackTransaction error:%s", err.Error())
		}
	}

	return
}

func (s *impl) finalTransaction(err *cd.Error) {
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
}

func (s *impl) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}
