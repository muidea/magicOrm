package orm

import (
	"fmt"
	"github.com/muidea/magicOrm/model"

	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

var _config *ormConfig

func init() {
}

// Initialize InitOrm
func Initialize(maxConnNum int, user, password, address, dbName string, localProvider bool) error {
	cfg := &serverConfig{user: user, password: password, address: address, dbName: dbName}

	_config = newConfig(localProvider)

	_config.updateServerConfig(cfg)

	return executor.InitializePool(maxConnNum, user, password, address, dbName)
}

// Uninitialize Uninitialize orm
func Uninitialize() {
	_config = nil

	executor.UninitializePool()
}

// NewOrm create new Orm
func NewOrm(owner string) (orm.Orm, error) {
	cfg := _config.getServerConfig()
	if cfg == nil {
		return nil, fmt.Errorf("not define databaes server config")
	}

	executor, err := executor.NewExecutor(cfg.user, cfg.password, cfg.address, cfg.dbName)
	if err != nil {
		return nil, err
	}

	orm := orm.New(executor, _config.getProvider(owner))
	return orm, nil
}

// GetOrm get orm from pool
func GetOrm(owner string) (orm.Orm, error) {
	executor, err := executor.GetExecutor()
	if err != nil {
		return nil, err
	}

	orm := orm.New(executor, _config.getProvider(owner))
	return orm, nil
}

func GetProvider(owner string) provider.Provider {
	return _config.getProvider(owner)
}

func GetFilter(owner string) model.Filter {
	return orm.NewFilter(_config.getProvider(owner))
}
