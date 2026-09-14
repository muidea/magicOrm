//go:build !mysql
// +build !mysql

package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/postgres"
	"github.com/muidea/magicOrm/provider"
)

// NewPool new executor pool
func NewPool() database.Pool {
	return postgres.NewPool()
}

// NewExecutor NewExecutor
func NewExecutor(config database.Config) (database.Executor, *cd.Error) {
	return postgres.NewExecutor(config)
}

func NewConfig(dbServer, dbName, schemaName, username, password string) database.Config {
	return postgres.NewConfig(dbServer, dbName, schemaName, username, password)
}

func ensureSchema(config database.Config) *cd.Error {
	value, ok := config.(*postgres.Config)
	if !ok {
		return cd.NewError(cd.IllegalParam, "postgres database configuration type is invalid")
	}
	return postgres.EnsureSchema(value)
}

func NewBuilder(provider provider.Provider, modelCodec codec.Codec) database.Builder {
	return postgres.NewBuilder(provider, modelCodec)
}
