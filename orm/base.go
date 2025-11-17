package orm

import (
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type baseRunner struct {
	vModel        models.Model
	executor      database.Executor
	modelProvider provider.Provider
	modelCodec    codec.Codec
	sqlBuilder    database.Builder

	batchFilter bool
	deepLevel   int
}

func newBaseRunner(
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	batchFilter bool,
	deepLevel int) baseRunner {
	return baseRunner{
		vModel:        vModel,
		batchFilter:   batchFilter,
		deepLevel:     deepLevel,
		executor:      executor,
		modelProvider: provider,
		modelCodec:    modelCodec,
		sqlBuilder:    NewBuilder(provider, modelCodec),
	}
}
