//go:build mysql
// +build mysql

package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/provider"
)

// NewPool new executor pool
func NewPool() database.Pool {
	return mysql.NewPool()
}

// NewExecutor NewExecutor
func NewExecutor(config database.Config) (database.Executor, *cd.Error) {
	return mysql.NewExecutor(config)
}

func NewConfig(dbServer, dbName, username, password string) database.Config {
	return mysql.NewConfig(dbServer, dbName, username, password, "")
}

func NewBuilder(provider provider.Provider, modelCodec codec.Codec) database.Builder {
	return mysql.NewBuilder(provider, modelCodec)
}
