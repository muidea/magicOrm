package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/postgres"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/provider"
)

// NewPool new executor pool
func NewPool() executor.Pool {
	return postgres.NewPool()
}

// NewExecutor NewExecutor
func NewExecutor(config executor.Config) (executor.Executor, *cd.Error) {
	return postgres.NewExecutor(config)
}

func NewConfig(dbServer, dbName, username, password string) executor.Config {
	return postgres.NewConfig(dbServer, dbName, username, password)
}

func NewBuilder(provider provider.Provider, modelCodec codec.Codec) builder.Builder {
	return postgres.NewBuilder(provider, modelCodec)
}
