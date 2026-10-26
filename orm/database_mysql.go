//go:build mysql
// +build mysql

package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/database/mysql"
	"github.com/muidea/magicOrm/provider"
)

// newPool new executor pool
func newPool() database.Pool {
	return mysql.NewPool()
}

// newExecutor newExecutor
func newExecutor(config database.Config) (database.Executor, *cd.Error) {
	return mysql.NewExecutor(config)
}

func newConfig(dbServer, dbName, schemaName, username, password string) database.Config {
	return mysql.NewConfig(dbServer, dbName, schemaName, username, password, "")
}

func ensureSchema(config database.Config) *cd.Error {
	value, ok := config.(*mysql.Config)
	if !ok {
		return cd.NewError(cd.IllegalParam, "mysql database configuration type is invalid")
	}
	return mysql.EnsureSchema(value)
}

func newBuilder(provider provider.Provider, modelCodec codec.Codec) database.Builder {
	return mysql.NewBuilder(provider, modelCodec)
}
