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

func NewConfig(dbServer, dbName, username, password string) database.Config {
	return postgres.NewConfig(dbServer, dbName, username, password)
}

func NewBuilder(provider provider.Provider, modelCodec codec.Codec) database.Builder {
	return postgres.NewBuilder(provider, modelCodec)
}
