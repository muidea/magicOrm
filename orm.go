package orm

import (
	"fmt"

	"muidea.com/magicOrm/executor"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/provider"
)

// Orm orm interfalce
type Orm interface {
	Create(obj interface{}) error
	Insert(obj interface{}) error
	Update(obj interface{}) error
	Delete(obj interface{}) error
	Query(obj interface{}) error
	BatchQuery(sliceObj interface{}, filter model.Filter) error
	Drop(obj interface{}) error
	Release()
}

var ormManager *manager

func init() {
	ormManager = newManager()
}

type orm struct {
	executor      executor.Executor
	modelProvider provider.Provider
}

// Initialize InitOrm
func Initialize(user, password, address, dbName string) error {
	cfg := &serverConfig{user: user, password: password, address: address, dbName: dbName}

	ormManager.updateServerConfig(cfg)

	return nil
}

// Uninitialize Uninitialize orm
func Uninitialize() {

}

// NewFilter create new filter
func NewFilter() model.Filter {
	return &queryFilter{params: map[string]model.FilterItem{}, modelProvider: ormManager.getProvider()}
}

// New create new Orm
func New() (Orm, error) {
	cfg := ormManager.getServerConfig()
	if cfg == nil {
		return nil, fmt.Errorf("not define databaes server config")
	}

	executor, err := executor.NewExecutor(cfg.user, cfg.password, cfg.address, cfg.dbName)
	if err != nil {
		return nil, err
	}

	return &orm{executor: executor, modelProvider: ormManager.getProvider()}, nil
}

func (s *orm) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}
