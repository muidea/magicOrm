package orm

import (
	"fmt"

	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	ormImpl "github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

// Orm orm interfalce
type Orm interface {
	RegisterModel(entity interface{}, owner string) error
	UnregisterModel(entity interface{}, owner string)
	Create(entity interface{}, owner string) error
	Drop(entity interface{}, owner string) error
	Insert(entity interface{}, owner string) error
	Update(entity interface{}, owner string) error
	Delete(entity interface{}, owner string) error
	Query(entity interface{}, owner string) error
	BatchQuery(sliceEntity interface{}, filter model.Filter, owner string) error
	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()
	Release()
}

var _config *ormConfig

func init() {
}

type orm struct {
	executor      executor.Executor
	modelProvider provider.Provider
	ormImpl       *ormImpl.Orm
}

// Initialize InitOrm
func Initialize(user, password, address, dbName string, localProvider bool) error {
	cfg := &serverConfig{user: user, password: password, address: address, dbName: dbName}

	_config = newConfig(localProvider)

	_config.updateServerConfig(cfg)

	return nil
}

// Uninitialize Uninitialize orm
func Uninitialize() {
	_config = nil
}

// NewFilter create new filter
func NewFilter() model.Filter {
	modelProvider := _config.getProvider()
	return ormImpl.NewFilter(modelProvider)
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

	modelProvider := _config.getProvider()
	ormImpl := ormImpl.New(executor, modelProvider)
	return &orm{executor: executor, modelProvider: modelProvider, ormImpl: ormImpl}, nil
}

func (s *orm) RegisterModel(entity interface{}, owner string) error {
	return s.modelProvider.RegisterModel(entity)
}

func (s *orm) UnregisterModel(entity interface{}, owner string) {
	s.modelProvider.UnregisterModel(entity)
}

func (s *orm) Create(entity interface{}, owner string) (err error) {
	err = s.ormImpl.Create(entity)

	return
}

func (s *orm) Drop(entity interface{}, owner string) (err error) {
	err = s.ormImpl.Drop(entity)
	return
}

func (s *orm) Insert(entity interface{}, owner string) (err error) {
	err = s.ormImpl.Insert(entity)
	return
}

func (s *orm) Query(entity interface{}, owner string) (err error) {

	err = s.ormImpl.Query(entity)
	return
}

func (s *orm) Delete(entity interface{}, owner string) (err error) {
	err = s.ormImpl.Delete(entity)

	return
}

func (s *orm) Update(entity interface{}, owner string) (err error) {
	err = s.ormImpl.Update(entity)

	return
}

func (s *orm) BatchQuery(sliceEntity interface{}, filter model.Filter, owner string) (err error) {

	err = s.ormImpl.BatchQuery(sliceEntity, filter)
	return
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
