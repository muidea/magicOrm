package orm

import (
	"fmt"

	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Orm orm interfalce
type Orm interface {
	RegisterModel(entity interface{}) error
	UnregisterModel(entity interface{})
	Create(entity interface{}) error
	Drop(entity interface{}) error
	Insert(entity interface{}) error
	Update(entity interface{}) error
	Delete(entity interface{}) error
	Query(entity interface{}) error
	BatchQuery(sliceEntity interface{}, filter model.Filter) error
	Release()
}

var _config *ormConfig

func init() {
}

type orm struct {
	executor      executor.Executor
	modelProvider provider.Provider
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
	return &queryFilter{params: map[string]model.FilterItem{}, modelProvider: _config.getProvider()}
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

	return &orm{executor: executor, modelProvider: _config.getProvider()}, nil
}

func (s *orm) RegisterModel(entity interface{}) error {
	return s.modelProvider.RegisterModel(entity)
}

func (s *orm) UnregisterModel(entity interface{}) {
	s.modelProvider.UnregisterModel(entity)
}

func (s *orm) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}
