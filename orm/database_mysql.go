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

func NewConfig(dbServer, dbName, schemaName, username, password string) database.Config {
	return mysql.NewConfig(dbServer, dbName, schemaName, username, password, "")
}

func ensureSchema(config database.Config) *cd.Error {
	value, ok := config.(*mysql.Config)
	if !ok {
		return cd.NewError(cd.IllegalParam, "mysql database configuration type is invalid")
	}
	return mysql.EnsureSchema(value)
}

func NewBuilder(provider provider.Provider, modelCodec codec.Codec) database.Builder {
	return mysql.NewBuilder(provider, modelCodec)
}
