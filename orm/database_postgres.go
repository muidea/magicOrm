//go:build !mysql
// +build !mysql

package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/database/postgres"
	"github.com/muidea/magicOrm/provider"
)

// newPool new executor pool
func newPool() database.Pool {
	return postgres.NewPool()
}

// newExecutor newExecutor
func newExecutor(config database.Config) (database.Executor, *cd.Error) {
	return postgres.NewExecutor(config)
}

func newConfig(dbServer, dbName, schemaName, username, password string) database.Config {
	return postgres.NewConfig(dbServer, dbName, schemaName, username, password)
}

func ensureSchema(config database.Config) *cd.Error {
	value, ok := config.(*postgres.Config)
	if !ok {
		return cd.NewError(cd.IllegalParam, "postgres database configuration type is invalid")
	}
	return postgres.EnsureSchema(value)
}

func newBuilder(provider provider.Provider, modelCodec codec.Codec) database.Builder {
	return postgres.NewBuilder(provider, modelCodec)
}
