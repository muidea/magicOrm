package orm

import (
	"fmt"
	"sync"

	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	ormImpl "github.com/muidea/magicOrm/orm"
)

// Orm orm interface
type Orm interface {
	RegisterModel(entity interface{}, owner string) error
	UnregisterModel(entity interface{}, owner string)
	Create(entity interface{}, owner string) error
	Drop(entity interface{}, owner string) error
	Insert(entity interface{}, owner string) error
	Update(entity interface{}, owner string) error
	Delete(entity interface{}, owner string) error
	Query(entity interface{}, owner string) error
	Count(entity interface{}, filter model.Filter, owner string) (int64, error)
	BatchQuery(sliceEntity interface{}, filter model.Filter, owner string) error
	QueryFilter(owner string) model.Filter
	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()
	Release()
}

var _config *ormConfig

func init() {
}

type orm struct {
	executor        executor.Executor
	ownerOrmImplMap map[string]*ormImpl.Orm
	ownerOrmLock    sync.RWMutex
}

// Initialize InitOrm
func Initialize(maxConnNum int, user, password, address, dbName string, localProvider bool) error {
	cfg := &serverConfig{user: user, password: password, address: address, dbName: dbName}

	_config = newConfig(localProvider)

	_config.updateServerConfig(cfg)

	executor.InitializePool(maxConnNum, user, password, address, dbName)

	return nil
}

// Uninitialize Uninitialize orm
func Uninitialize() {
	_config = nil

	executor.UninitializePool()
}

// New create new Orm
func New() (Orm, error) {
	cfg := _config.getServerConfig()
	if cfg == nil {
		return nil, fmt.Errorf("not define databaes server config")
	}

	executor, err := executor.NewExecutor(cfg.user, cfg.password, cfg.address, cfg.dbName)
	if err != nil {
		return nil, err
	}

	return &orm{executor: executor, ownerOrmImplMap: map[string]*ormImpl.Orm{}}, nil
}

// Get get orm from pool
func Get() (Orm, error) {
	executor, err := executor.GetExecutor()
	if err != nil {
		return nil, err
	}

	return &orm{executor: executor, ownerOrmImplMap: map[string]*ormImpl.Orm{}}, nil
}

func (s *orm) RegisterModel(entity interface{}, owner string) error {
	ormPtr := s.getOrm(owner)
	return ormPtr.RegisterModel(entity)
}

func (s *orm) UnregisterModel(entity interface{}, owner string) {
	ormPtr := s.getOrm(owner)

	ormPtr.UnregisterModel(entity)
}

func (s *orm) Create(entity interface{}, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.Create(entity)
	return
}

func (s *orm) Drop(entity interface{}, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.Drop(entity)
	return
}

func (s *orm) Insert(entity interface{}, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.Insert(entity)
	return
}

func (s *orm) Query(entity interface{}, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.Query(entity)
	return
}

func (s *orm) Delete(entity interface{}, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.Delete(entity)
	return
}

func (s *orm) Update(entity interface{}, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.Update(entity)
	return
}

func (s *orm) Count(entity interface{}, filter model.Filter, owner string) (ret int64, err error) {
	ormPtr := s.getOrm(owner)

	ret, err = ormPtr.Count(entity, filter)
	return
}

func (s *orm) BatchQuery(sliceEntity interface{}, filter model.Filter, owner string) (err error) {
	ormPtr := s.getOrm(owner)

	err = ormPtr.BatchQuery(sliceEntity, filter)
	return
}

func (s *orm) QueryFilter(owner string) model.Filter {
	ormPtr := s.getOrm(owner)

	return ormPtr.NewQueryFilter()
}

func (s *orm) BeginTransaction() {
	if s.executor != nil {
		s.executor.BeginTransaction()
	}
}

func (s *orm) CommitTransaction() {
	if s.executor != nil {
		s.executor.CommitTransaction()
	}
}

func (s *orm) RollbackTransaction() {
	if s.executor != nil {
		s.executor.RollbackTransaction()
	}
}

func (s *orm) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}

func (s *orm) getOrm(owner string) *ormImpl.Orm {
	s.ownerOrmLock.Lock()
	defer s.ownerOrmLock.Unlock()

	curOrm, ok := s.ownerOrmImplMap[owner]
	if !ok {
		curProvider := _config.getProvider(owner)
		curOrm = ormImpl.New(s.executor, curProvider)

		s.ownerOrmImplMap[owner] = curOrm
	}

	return curOrm
}
